package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/Jumpscale/hubble/agent"
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
		remotehost := match[4]

		local, err := strconv.ParseUint(match[1], 10, 16)
		if err != nil {
			log.Fatalf("Invalid port %v", match[1])
		}

		remoteport, err := strconv.ParseUint(match[5], 10, 16)
		if err != nil {
			log.Fatalf("Invalid port %v", match[5])
		}

		key := match[2]
		if key != "" {
			key = key[:len(key)-1]
		}

		gw := match[3]

		tunnel := agent.NewTunnel(uint16(local), gw, key, remotehost, uint16(remoteport))
		tunnels = append(tunnels, tunnel)
	}

	if url == "" {
		printHelp()
		log.Fatal("Missing url")
	}

	if name == "" {
		printHelp()
		log.Fatal("Missing name")
	}

	var config tls.Config

	if ca != "" {
		pem, err := ioutil.ReadFile(ca)
		if err != nil {
			log.Fatal(err)
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
