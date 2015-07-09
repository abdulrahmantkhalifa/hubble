package agent

import (
	"git.aydo.com/0-complexity/hubble"
	"net"
	"fmt"
	"log"
	"time"
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
func (tunnel *Tunnel) serve(sessions sessionsStore, conn *hubble.Connection) {
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

			go tunnel.handle(sessions, conn, socket)
		}
	}()
}

func (tunnel *Tunnel) handle(sessions sessionsStore, conn *hubble.Connection, socket net.Conn) {
	guid := uuid.New()

	defer func() {
		log.Printf("Session %v on tunnel %v terminated\n", guid, tunnel)
		socket.Close()
	}()

	channel := registerSession(sessions, guid)
	defer unregisterSession(sessions, guid)

	//1- send initiator message ...
	log.Printf("Starting session %v on tunnel %v", guid, tunnel)

	err := conn.Send(hubble.NewInitiatorMessage(guid,
		tunnel.ip, tunnel.remote, tunnel.gateway))

	if err != nil {
		log.Printf("Failed to start session %v to %v: %v\n", guid, tunnel, err)
		return
	}

	//read first message. must be ack.

	//2- recieve ack
	log.Println("Waiting for ack from:", tunnel.gateway)
	select {
		case message, ok := <- channel:
			if !ok || message.GetMessageType() != hubble.ACK_MESSAGE_TYPE {
				log.Println("Expecting ack message, got: ", message.GetMessageType())
				return
			}

			ack := message.(*hubble.AckMessage)
			if !ack.Ok {
				//failed to start session!
				return
			}
		case <- time.After(30 * time.Second):
			//timedout, return
			return
	}

	log.Printf("Session %v started...", guid)

	serveSession(guid, conn, channel, socket)
}

