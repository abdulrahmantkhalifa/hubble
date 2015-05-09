
package main

import (
	"hubble"
	"hubble/agent"
	"log"
	"net"
	"flag"
	"regexp"
	"strconv"
)


func main() {
	var url string
	var name string
	var key string

	flag.StringVar(&url, "url", "", "WebSocket url to proxy server in form 'ws://host:port/path")
	flag.StringVar(&name, "name", "", "Agent name, will be used by other agents to redirect connections to this agent")
	flag.StringVar(&key, "key", "", "Agent key")
	flag.Parse()

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
		flag.PrintDefaults()
		log.Fatal("Missing url")
	}

	if name == "" {
		flag.PrintDefaults()
		log.Fatal("Missing name")
	}

	// if key == "" {
	// 	flag.PrintDefaults()
	// 	log.Fatal("Missing key")
	// }

	//1- intialize connection to proxy
	conn, err := hubble.NewProxyConnection(url)
	if err != nil {
		log.Fatal("Failed to connect to proxy", err)
	}
	
	//2- registration
	err = agent.Handshake(conn, name, key)
	if err != nil {
		log.Fatal(err)
	}

	go func () {
		//receive all messages.
		log.Println("Start receiving loop")
		for {
			mtype, message, err := conn.Receive()
			if err != nil {
				//we should check error types to take a decistion. for now just exit
				log.Fatalf("Receive loop failed: %v", err)
			}
			log.Println(mtype, message)
			switch mtype {
				case hubble.INITIATOR_MESSAGE_TYPE:
					//TODO: Make a connection to service.
					initiator := message.(*hubble.InitiatorMessage)
					//send ack (debug)
					conn.SendAckOrError(initiator.GUID, nil)
				default:
					agent.Dispatch(mtype, message.(hubble.SessionMessage))
			}
		}
	}()

	for _, tunnel := range tunnels {
		tunnel.Serve(conn)
	}

	//wait forever
	select {}
}