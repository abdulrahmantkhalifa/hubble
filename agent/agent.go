package agent

import (
	"git.aydo.com/0-complexity/hubble"
	"log"
	"crypto/tls"
)


//Handshake
func handshake(conn *hubble.Connection, agentname string, key string) error {

	message := hubble.NewHandshakeMessage(agentname, key)

	err := conn.Send(message)
	if err != nil {
		return err
	}

	//Next message must be an ack message, read ack.
	err = conn.ReceiveAck()
	if err != nil {
		return err
	}

	return nil
}

func dispatch(sessions sessionsStore, message hubble.SessionMessage) {
	defer func () {
		if err := recover(); err != nil {
			//Can't send data to session channel?!. Please don't panic, chill out and
			//relax, it's probably closed. Do nothing.
		}
	} ()
	session, ok := sessions[message.GetGUID()]

	if !ok {
		if message.GetMessageType() != hubble.TERMINATOR_MESSAGE_TYPE {
			log.Println("Message to unknow session received: ", message.GetGUID(), message.GetMessageType())
		}

		return
	}

	session <- message
}

func Agent(name string, key string, url string, tunnels []*Tunnel,config *tls.Config) (err error) {
	//1- intialize connection to proxy
	conn, err := hubble.NewProxyConnection(url, config)
	if err != nil {
		return
	}

	//2- registration
	err = handshake(conn, name, key)
	if err != nil {
		return
	}

	sessions := make(sessionsStore)

	go func () {
		//receive all messages.
		log.Println("Start receiving loop")
		for {
			message, err := conn.Receive()
			if err != nil {
				//we should check error types to take a decistion. for now just exit
				log.Fatalf("Receive loop failed: %v", err)
			}
			switch message.GetMessageType() {
				case hubble.INITIATOR_MESSAGE_TYPE:
					initiator := message.(*hubble.InitiatorMessage)
					startLocalSession(sessions, conn, initiator)
				default:
					dispatch(sessions, message.(hubble.SessionMessage))
			}
		}
	}()

	for _, tunnel := range tunnels {
		tunnel.serve(sessions, conn)
	}

	return nil
}
