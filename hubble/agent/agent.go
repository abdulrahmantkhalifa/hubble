package agent

import (
	"hubble"
	"log"
	"crypto/tls"
)


//Handshake
func handshake(conn *hubble.Connection, agentname string, key string) error {
	message := hubble.HandshakeMessage {
		Version: hubble.PROTOCOL_VERSION_0_1,
		Name: agentname,
		Key: key,
	}

	err := conn.Send(hubble.HANDSHAKE_MESSAGE_TYPE, &message)
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

func dispatch(sessions sessionsStore, mtype uint8, message hubble.SessionMessage) {
	defer func () {
		if err := recover(); err != nil {
			//Can't send data to session channel?!. Please don't panic, chill out and 
			//relax, it's probably closed. Do nothing.
		}
	} ()
	session, ok := sessions[message.GetGUID()]

	if !ok {
		if mtype != hubble.TERMINATOR_MESSAGE_TYPE {
			log.Println("Message to unknow session received: ", message.GetGUID(), mtype)
		}

		return
	}

	capsule := new(hubble.MessageCapsule)
	capsule.Mtype = mtype
	capsule.Message = message

	session <- capsule
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
			mtype, message, err := conn.Receive()
			if err != nil {
				//we should check error types to take a decistion. for now just exit
				log.Fatalf("Receive loop failed: %v", err)
			}
			switch mtype {
				case hubble.INITIATOR_MESSAGE_TYPE:
					initiator := message.(*hubble.InitiatorMessage)
					startLocalSession(sessions, conn, initiator)
				default:
					dispatch(sessions, mtype, message.(hubble.SessionMessage))
			}
		}
	}()

	for _, tunnel := range tunnels {
		tunnel.serve(sessions, conn)
	}

	return nil
}