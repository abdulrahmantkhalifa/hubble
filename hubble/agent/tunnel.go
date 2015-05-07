package agent

import (
	"hubble"
	"net"
	"fmt"
	"log"
	"code.google.com/p/go-uuid/uuid"
)

type Tunnel struct {
	local uint16
	ip net.IP
	remote uint16
	gateway string
}

func NewTunnel(local uint16, gateway string, ip net.IP, remote uint16) *Tunnel {
	tunnel := new(Tunnel)
	tunnel.local = local
	tunnel.ip = ip
	tunnel.gateway = gateway
	tunnel.remote = remote

	return tunnel
}

func (tunnel *Tunnel) String() string {
	return fmt.Sprintf("%v:%v:%v:%v", tunnel.local, tunnel.gateway, tunnel.ip, tunnel.remote)
}

//Open the tunnel on local side and server over the given connection to the proxy.
func (tunnel *Tunnel) Serve(conn *hubble.Connection) {
	go func() {
		// open socket and wait for connections
		listner, err := net.Listen("tcp", fmt.Sprintf(":%d", tunnel.local))
		defer listner.Close()

		if err != nil {
			log.Printf("Failed to listing on port %v: %v\n", tunnel.local, err)
			return
		}

		for {
			socket, err := listner.Accept()
			if err != nil {
				log.Println(err)
				socket.Close()
			}

			go tunnel.handle(conn, socket)
		}
	}()
}

func (tunnel *Tunnel) handle(conn *hubble.Connection, socket net.Conn) {
	guid := uuid.New()
	
	defer func() {
		log.Printf("Terminating session %v on tunnel %v\n", guid, tunnel)
		socket.Close()
	}
	
	//1- send initiator message ...
	log.Printf("Starting session %v on tunnel %v", guid, tunnel)

	err := conn.Send(hubble.INITIATOR_MESSAGE_TYPE, &hubble.InitiatorMessage {
		GUID: guid,
		Ip: tunnel.ip,
		Port: tunnel.remote,
		Gatename: tunnel.gateway,
	})

	if err != nil {
		log.Printf("Failed to start session %v to %v: %v\n", guid, tunnel, err)
		return
	}

	//2- recieve ack
	err = conn.ReceiveAck()
	if err != nil {
		log.Println(err)
		return
	}

	//otherwise start forwarding routines
}