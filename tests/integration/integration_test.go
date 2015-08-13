package hubble

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"testing"

	"github.com/Jumpscale/hubble/agent"
	"github.com/Jumpscale/hubble/auth"
	"github.com/Jumpscale/hubble/logging"
	"github.com/Jumpscale/hubble/proxy"
	"github.com/Jumpscale/hubble/proxy/events"
	"github.com/stretchr/testify/assert"
)

const NUM_FILES int = 10
const BLOCK_SIZE = 512 //KB
const FILE_SIZE = 1    //M

var eventLogger *testingEventLogger
var proxyPort = uint16(0)
var fileServerPort = uint16(0)
var localTunnelPort = uint16(0)

func md5sum_r(reader io.Reader) (string, error) {
	md5 := md5.New()
	buffer := make([]byte, 1024)
	for {
		count, err := reader.Read(buffer)

		if err != nil && err != io.EOF {
			return "", err
		}

		md5.Write(buffer[:count])

		if err == io.EOF {
			break
		}
	}

	return fmt.Sprintf("%x", md5.Sum(nil)), nil
}

func md5sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	return md5sum_r(file)
}

var hashes = make(map[string]string)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup

	auth.Install(auth.NewAcceptAllModule())

	eventLogger = &testingEventLogger{}
	logging.InstallEventLogger(eventLogger)

	wg.Add(1)
	//starting proxy
	go func() {
		defer wg.Done()

		http.HandleFunc("/", proxy.ProxyHandler)
		listner, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Println(err)

		}

		proxyPort, err = portForListener(listner)
		if err != nil {
			fmt.Println(err)
		}

		go http.Serve(listner, nil)

	}()

	//wait until proxy is ready before starting agents.
	wg.Wait()

	//now we need to start a file server that serves some files
	tempdir := fmt.Sprintf("%s/%s", os.TempDir(), "hubble_t")
	err := os.MkdirAll(tempdir, os.ModeDir|os.ModePerm)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer os.RemoveAll(tempdir)

	wg.Add(1)

	go func() {
		defer wg.Done()
		listner, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fileServerPort, err = portForListener(listner)
		if err != nil {
			fmt.Println(err)
		}

		go http.Serve(listner, http.FileServer(http.Dir(tempdir)))
	}()

	wg.Wait()

	//start 1st agents
	//the first agent a1 should serve tunnel (dynamic):gw2:127.0.0.1:(fileServerPort)
	tunnel := agent.NewTunnel(0, "gw2", "connection_key", net.ParseIP("127.0.0.1"), fileServerPort)
	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/", proxyPort)

	gw1 := agent.NewAgent(wsURL, "gw1", "", nil)
	gw1.Start(nil)

	gw1.AddTunnel(tunnel)
	localTunnelPort = tunnel.Local()

	//start second agent
	gw2 := agent.NewAgent(wsURL, "gw2", "", nil)
	gw2.Start(nil)

	//Create files to serve.
	for i := 0; i < NUM_FILES; i++ {
		fname := fmt.Sprintf("file-%d", i)
		fpath := fmt.Sprintf("%s/%s", tempdir, fname)
		cmd := exec.Command("dd",
			"if=/dev/urandom",
			fmt.Sprintf("of=%s", fpath),
			fmt.Sprintf("bs=%dk", BLOCK_SIZE),
			fmt.Sprintf("count=%d", FILE_SIZE*1024/BLOCK_SIZE))

		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		//calculate md5sum for the file
		hash, err := md5sum(fpath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		hashes[fname] = hash
	}

	os.Exit(m.Run())
}

//download a file and return its md5sum hash
func download(fname string) (hash string, err error) {
	//we go over the forwarded port of course
	localURL := fmt.Sprintf("http://127.0.0.1:%d/%s", localTunnelPort, fname)
	response, err := http.Get(localURL)
	if err != nil {
		return
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = errors.New("Invalid status code")
		return
	}

	hash, err = md5sum_r(response.Body)
	return
}

func TestDownload(t *testing.T) {
	//now the status is the following
	//1- We have a proxy running
	//2- We have 2 agents running
	//3- We have a file server running.
	//4- We have files served by the file server

	//now what we need to do is to download all files
	//and then calculate md5sum of the downloaded files to make sure
	//they are all okay.

	var downloadWg sync.WaitGroup
	downloadWg.Add(NUM_FILES)
	for i := 0; i < NUM_FILES; i++ {
		fname := fmt.Sprintf("file-%d", i)
		go func() {
			defer downloadWg.Done()

			downloaded_hash, err := download(fname)
			if err != nil {
				t.Error(err)
			}

			t.Log(fname, hashes[fname], downloaded_hash)
			if hashes[fname] != downloaded_hash {
				t.Errorf("File: %s with has %s has wrong downloaded hash %s",
					fname, hashes[fname], downloaded_hash)
			}
		}()
	}
	t.Log("Waiting for downloads to finish")
	downloadWg.Wait()

	//check if events are triggered
	e := eventLogger.events[0]
	if registration, ok := e.(events.GatewayRegistrationEvent); true {
		assert.True(t, ok, "Expected GatewayRegistrationEvent, got %T(%v)", e, e)
		assert.Equal(t, registration.Gateway, "gw1", "Expected gw1 registration first.")
	}

	e = eventLogger.events[1]
	if registration, ok := e.(events.GatewayRegistrationEvent); true {
		assert.True(t, ok, "Expected GatewayRegistrationEvent, got %T(%v)", e, e)
		assert.Equal(t, registration.Gateway, "gw2", "Expected gw2 registration first.")
	}

	for _, ev := range eventLogger.events[2:] {
		switch e := ev.(type) {
		case events.OpenSessionEvent:
			assert.Equal(t, e.SourceGateway, "gw1", "Expected gw1 as source.")
			assert.True(t, e.Success, "Expected open session success.")
			assert.NotNil(t, e.Error, "Did not expect error: %v", e.Error)
			assert.Equal(t, e.ConnectionRequest.IP.String(), "127.0.0.1", "Destination IP doesn't match.")
			assert.Equal(t, e.ConnectionRequest.Gatename, "gw2", "Destination gateway doesn't match.")
			assert.Equal(t, e.ConnectionRequest.Key, "connection_key", "Connection key doesn't match.")
			assert.Equal(t, e.ConnectionRequest.Port, fileServerPort, "Local tunnel port doesn't match.")

		case events.CloseSessionEvent:
			if e.Gateway != "gw1" && e.Gateway != "gw2" {
				assert.True(t, false, "Unknown closed session at gateway: %v", e.Gateway)
			}
			assert.Equal(t, e.ConnectionKey, "connection_key", "Connection key doesn't match.")
		}
	}
}

func BenchmarkDownload(b *testing.B) {
	fname := "file-0"
	for i := 0; i < b.N; i++ {
		download(fname)
	}
}

type testingEventLogger struct {
	eventsLock sync.Mutex
	events     []logging.Event
}

func (logger *testingEventLogger) Log(event logging.Event) error {
	logger.eventsLock.Lock()
	logger.events = append(logger.events, event)
	logger.eventsLock.Unlock()
	return nil
}

func portForListener(l net.Listener) (port uint16, err error) {
	_, s, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return
	}

	i, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return
	}

	port = uint16(i)
	return
}
