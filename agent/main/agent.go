package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/Jumpscale/hubble/agent"
	"github.com/Jumpscale/hubble/logging"
)

func main() {
	var url string
	var name string
	var key string
	var ca string

	var help bool

	flag.BoolVar(&help, "h", false, "Print this help screen")
	flag.StringVar(&url, "url", "", "WebSocket url to proxy server in form 'ws://host:port/path")
	flag.StringVar(&name, "name", "", "Agent name, will be used by other agents to redirect connections to this agent")
	flag.StringVar(&key, "key", "", "Agent key")
	flag.StringVar(&ca, "ca", "", "Certificate Autority to trust")
	flag.Parse()

	printHelp := func() {
		fmt.Println("agent [options] [[local port:[key@]gateway:remot ip:remot port]...]")
		flag.PrintDefaults()
	}

	if help {
		printHelp()
		return
	}

	tunnels_def := flag.Args()
	tunnels := make([]*agent.Tunnel, 0)

	//tunnel is defined as lport:[key@]gw:ip:port
	re := regexp.MustCompile("(\\d+):(.*@)?([^:]+):([^:]+):(\\d+)")

	for _, tunnel_def := range tunnels_def {
		match := re.FindStringSubmatch(tunnel_def)
		ip := net.ParseIP(match[4])
		if ip == nil {
			logging.Fatalf("Invalid ip address %v", match[4])
		}

		local, err := strconv.ParseUint(match[1], 10, 16)
		if err != nil {
			logging.Fatalf("Invalid port %v", match[1])
		}

		remote, err := strconv.ParseUint(match[5], 10, 16)
		if err != nil {
			logging.Fatalf("Invalid port %v", match[5])
		}

		key := match[2]
		if key != "" {
			key = key[:len(key)-1]
		}

		gw := match[3]

		tunnel := agent.NewTunnel(uint16(local), gw, key, ip, uint16(remote))
		tunnels = append(tunnels, tunnel)
	}

	if url == "" {
		printHelp()
		logging.Fatal("Missing url")
	}

	if name == "" {
		printHelp()
		logging.Fatal("Missing name")
	}

	var config tls.Config

	if ca != "" {
		pem, err := ioutil.ReadFile(ca)
		if err != nil {
			logging.Fatal(err)
		}

		config.RootCAs = x509.NewCertPool()
		config.RootCAs.AppendCertsFromPEM(pem)
	}

	agt := agent.NewAgent(url, name, key, &config)

	var onExit func(agt agent.Agent, err error)

	onExit = func(agt agent.Agent, err error) {
		if err != nil {
			go func() {
				time.Sleep(10 * time.Second)
				agt.Start(onExit)
			}()
		}
	}

	agt.Start(onExit)

	for _, tunnel := range tunnels {
		agt.AddTunnel(tunnel)
	}

	//wait forever
	select {}
}
