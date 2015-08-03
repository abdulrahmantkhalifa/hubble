
package main

import (
	"github.com/Jumpscale/hubble/agent"
	"log"
	"fmt"
	"net"
	"flag"
	"regexp"
	"crypto/tls"
	"crypto/x509"
	"strconv"
	"io/ioutil"
	"time"
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
		fmt.Println("agent [options] [[local port:gateway:remot ip:remot port]...]")
		flag.PrintDefaults()
	}

	if help {
		printHelp()
		return
	}

	tunnels_def := flag.Args()
	tunnels := make([]*agent.Tunnel, 0)

	//tunnel is defined as lport:gw:ip:port
	re := regexp.MustCompile("(\\d+):([^:]+):([^:]+):(\\d+)")

	for _, tunnel_def := range tunnels_def {
		match := re.FindStringSubmatch(tunnel_def)
		ip := net.ParseIP(match[3])
		if ip == nil {
			log.Fatalf("Invalid ip address %v", match[3])
		}

		local, err := strconv.ParseUint(match[1], 10, 16)
		if err != nil {
			log.Fatalf("Invalid port %v", match[1])
		}

		remote, err := strconv.ParseUint(match[4], 10, 16)
		if err != nil {
			log.Fatalf("Invalid port %v", match[4])
		}
		tunnel := agent.NewTunnel(uint16(local), match[2], ip, uint16(remote))
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

	var onExit func (agt agent.Agent, err error)

	onExit = func (agt agent.Agent, err error) {
		if err != nil {
			go func(){
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
