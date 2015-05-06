
package main

import (
	"hubble"
	"fmt"
	"log"
	"net"
)

// func handle(conn net.Conn) {
// 	var ssh_conn, err = net.Dial("tcp", "localhost:22")
// 	if err != nil {
// 		log.Printf("Failed to connect to ssh service: %v", err)
// 	}

// 	go netwpoc.Pipe(conn, ssh_conn)
// }


func ServiceTunnel(tunnel *hubble.Tunnel) {
	// open socket and wait for connections
	listner, err := net.Listen("tcp", fmt.Sprintf(":%d", tunnel.Local))
	defer listner.Close()

	if err != nil {
		log.Printf("Failed to listing on port: %v", err)
		return
	}

	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Printf("accept socket error: %v", err)
			conn.Close()
		}

		//go handle(conn)
	}
}

//Wrapper for handshake
func Handshake(conn *hubble.ProxyConnection, agentname string, key string) error {
	message := hubble.HandshakeMessage {
		Name: agentname,
		Key: key,
	}

	return conn.Send(hubble.HANDSHAKE_MESSAGE_TYPE, &message)
}

func main() {
	conn, err := hubble.NewProxyConnection("ws://localhost:8080/")
	if err != nil {
		log.Fatal("Failed to connect to proxy", err)
	}
	
	Handshake(conn, "gw1", "password")
	// ws.WriteMessage(websocket.TextMessage, []byte("Hello Proxy"))

	// tunnels := []hubble.Tunnel {
	// 	//tunnel to ssh(22)  proxy->gw1->127.0.0.1
	// 	{Gateway: "gw1",
	// 	 Ip: net.ParseIP("127.0.0.1"),
	// 	 Local: 2015,
	// 	 Remote: 22},
	// }

	// for _, tunnel := range tunnels {
	// 	go ServiceTunnel(&tunnel)
	// }

	// //wait forever
	select {}
}