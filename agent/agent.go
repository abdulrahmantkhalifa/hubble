
package main

import (
	"hubble"
	"fmt"
	"log"
	"net"
)

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
func Handshake(conn *hubble.Connection, agentname string, key string) error {
	message := hubble.HandshakeMessage {
		Name: agentname,
		Key: key,
	}

	err := conn.Send(hubble.HANDSHAKE_MESSAGE_TYPE, &message)
	if err != nil {
		return err
	}

	//read ack.
	err = conn.ReceiveAck()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	conn, err := hubble.NewProxyConnection("ws://localhost:8080/")
	if err != nil {
		log.Fatal("Failed to connect to proxy", err)
	}
	
	err = Handshake(conn, "gw1", "password")
	if err != nil {
		log.Fatal(err)
	}

	//send initiator message

	err = conn.Send(hubble.INITIATOR_MESSAGE_TYPE, &hubble.InitiatorMessage{
		GUID: "First Session",
		Ip: net.ParseIP("172.0.0.1"),
		Port: 22,
		Gatename: "gw1",
	})

	err = conn.ReceiveAck()
	if err != nil {
		log.Fatal(err)
	}

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