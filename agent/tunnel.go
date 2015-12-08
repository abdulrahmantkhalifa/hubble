package agent

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/Jumpscale/hubble"
	"github.com/pborman/uuid"
)

type Tunnel struct {
	local      uint16
	remotehost string
	remoteport uint16
	gateway    string
	key        string
	listener   net.Listener
}

type ctrlChan chan int

func NewTunnel(local uint16, gateway, key, remotehost string, remoteport uint16) *Tunnel {
	tunnel := new(Tunnel)
	tunnel.local = local
	tunnel.remotehost = remotehost
	tunnel.gateway = gateway
	tunnel.key = key
	tunnel.remoteport = remoteport

	return tunnel
}

func (tunnel *Tunnel) Local() uint16 {
	return tunnel.local
}

func (tunnel *Tunnel) Host() string {
	return tunnel.remotehost
}

func (tunnel *Tunnel) Gateway() string {
	return tunnel.gateway
}

func (tunnel *Tunnel) RemotePort() uint16 {
	return tunnel.remoteport
}

func (tunnel *Tunnel) String() string {
	return fmt.Sprintf("%v:%v:%v:%v", tunnel.local, tunnel.gateway, tunnel.remotehost, tunnel.remoteport)
}

//Open the tunnel on local side and server over the given connection to the proxy.
func (tunnel *Tunnel) start(sessions sessionsStore, conn *hubble.Connection) error {
	log.Println("Starting tunnel", tunnel)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tunnel.local))

	if err != nil {
		log.Printf("Failed to listen on port %v: %v\n", tunnel.local, err)
		return err
	}

	if tunnel.local == 0 {
		addr := listener.Addr().String()
		if _, port, err := net.SplitHostPort(addr); err == nil {
			if i, err := strconv.ParseUint(port, 10, 16); err == nil {
				tunnel.local = uint16(i)
				log.Println("Listening on dynamic port", tunnel.local)
			}
		}
	}

	tunnel.listener = listener
	ctrl := make(ctrlChan)

	go func() {
		// open socket and wait for connections
		defer func() {
			listener.Close()
			//send kill signal to all sessions built on that tunnel.
			for {
				select {
				case ctrl <- 1:
				case <-time.After(5 * time.Second):
					log.Println("All tunnel sessions are terminated")
					return
				}
			}
		}()

		for {
			socket, err := listener.Accept()
			if err != nil {
				log.Println("Tunnel", tunnel, "closed", err)
				return
			}

			go tunnel.handle(sessions, conn, socket, ctrl)
		}
	}()

	return nil
}

func (tunnel *Tunnel) stop() {
	log.Println("Terminating tunnel", tunnel)
	if tunnel.listener != nil {
		tunnel.listener.Close()
	}
}

func (tunnel *Tunnel) handle(sessions sessionsStore, conn *hubble.Connection, socket net.Conn, ctrl ctrlChan) {
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
		tunnel.remotehost, tunnel.remoteport, tunnel.gateway, tunnel.key))

	if err != nil {
		log.Printf("Failed to start session %v to %v: %v\n", guid, tunnel, err)
		return
	}

	//read first message. must be ack.

	//2- recieve ack
	log.Println("Waiting for ack from:", tunnel.gateway)
	select {
	case message, ok := <-channel:
		if !ok || message.GetMessageType() != hubble.ACK_MESSAGE_TYPE {
			log.Println("Expecting ack message, got: ", message.GetMessageType())
			return
		}

		ack := message.(*hubble.AckMessage)
		if !ack.Ok {
			//failed to start session!
			return
		}
	case <-ctrl:
		//currently only ctrl signal is to terminate
		return
	case <-time.After(30 * time.Second):
		//timedout, return
		return
	}

	log.Printf("Session %v started...", guid)

	serveSession(guid, conn, channel, socket, ctrl)
}
