package hubble 

import (
	"net"
	"net/http"
	"hubble/proxy"
	"hubble/agent"
	"testing"
	"sync"
	"os"
	"os/exec"
	"io"
	"crypto/md5"
	"fmt"
	"errors"
)

const NUM_FILES int = 10
const BLOCK_SIZE = 512 //KB
const FILE_SIZE = 1 //M


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

	wg.Add(1)
	//starting proxy
	go func () {
		defer wg.Done()

		http.HandleFunc("/", proxy.ProxyHandler)
		listner, err := net.Listen("tcp", ":9999")
		if err != nil {
			fmt.Println(err)

		}
		go http.Serve(listner, nil)
		
	}()

	//wait until proxy is ready before starting agents.
	wg.Wait()

	//start 1st agents
	//the first agent a1 should serve tunnel 7777:gw2:127.0.0.1:5555
	tunnel := agent.NewTunnel(7777, "gw2", net.ParseIP("127.0.0.1"), 5555)

	agent.Agent("gw1", "", "ws://localhost:9999/", []*agent.Tunnel{tunnel}, nil)
	//start second agent
	agent.Agent("gw2", "", "ws://localhost:9999/", make([]*agent.Tunnel, 0), nil)

	//now we need to start a file server that serves on port 5555
	tempdir := fmt.Sprintf("%s/%s", os.TempDir(), "hubble_t")
	err := os.MkdirAll(tempdir, os.ModeDir | os.ModePerm)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer os.RemoveAll(tempdir)

	wg.Add(1)

	go func () {
		defer wg.Done()
		listner, err := net.Listen("tcp", ":5555")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		go http.Serve(listner, http.FileServer(http.Dir(tempdir)))
	} ()

	wg.Wait()
	
	

	//Create files to serve.
	for i := 0; i < NUM_FILES; i++ {
		fname := fmt.Sprintf("file-%d", i)
		fpath := fmt.Sprintf("%s/%s", tempdir, fname)
		cmd := exec.Command("dd",
							"if=/dev/urandom",
							fmt.Sprintf("of=%s", fpath),
							fmt.Sprintf("bs=%dk", BLOCK_SIZE),
							fmt.Sprintf("count=%d", FILE_SIZE * 1024 / BLOCK_SIZE))

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
	response, err := http.Get(fmt.Sprintf("http://localhost:7777/%s", fname))
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
		go func () {
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
		} ()
	}
	t.Log("Waiting for downloads to finish")
	downloadWg.Wait()
}

func BenchmarkDownload(b *testing.B) {
	fname := "file-0"
	for i := 0; i < b.N; i ++ {
		download(fname)
	}
}