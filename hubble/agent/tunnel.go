package agent

import (
	"hubble"
	"net"
	"fmt"
	"log"
	"io"
	"sync"
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
		delete(sessions, guid)
		socket.Close()
	}()
	
	channel := make(chan *hubble.MessageCapsule, sessionQueueSize)
	
	sessions[guid] = channel

	//1- send initiator message ...
	log.Printf("Starting session %v on tunnel %v", guid, tunnel)

	err := conn.Send(hubble.INITIATOR_MESSAGE_TYPE, &hubble.InitiatorMessage {
		GuidMessage: hubble.GuidMessage{guid},
		Ip: tunnel.ip,
		Port: tunnel.remote,
		Gatename: tunnel.gateway,
	})
	
	if err != nil {
		log.Printf("Failed to start session %v to %v: %v\n", guid, tunnel, err)
		return
	}

	//read first message. must be ack.

	//2- recieve ack
	log.Println("Waiting for ack from:", tunnel.gateway)
	msgCap := <- channel
	ack := msgCap.Message.(*hubble.AckMessage)
	if !ack.Ok {
		//failed to start session!
		log.Println(ack.Message)
		return
	}

	log.Printf("Session %v started...", guid)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		//socket -> proxy 
		defer wg.Done()
		log.Println("Start session sender")
		
		buffer := make([]byte, 1024)
		for {
			count, read_err := socket.Read(buffer)
			if read_err != nil && read_err != io.EOF {
				log.Printf("Failer on session %v %v: %v", guid, tunnel, read_err)
				return
			}

			err = conn.Send(hubble.DATA_MESSAGE_TYPE, &hubble.DataMessage{
				GuidMessage: hubble.GuidMessage{guid},
				Data: buffer[0:count],
			})

			if err != nil {
				//failed to forward data to proxy
				log.Printf("Failer on session %v %v: %v", guid, tunnel, err)
				return
			}

			if read_err == io.EOF {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("Start sessoin receiver")
		
		for {
			msgCap := <- channel
			log.Println(msgCap.Mtype, msgCap.Message)
		}
	}()

	wg.Wait()
}
