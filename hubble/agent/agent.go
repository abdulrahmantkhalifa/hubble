package agent

import (
	"hubble"
	"log"
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

func dispatch(mtype uint8, message hubble.SessionMessage) {
	go func() {
		defer func (){
			if err := recover(); err != nil {
				//Can't send data to session channel?!. Please don't panic, chill out and 
				//relax, it's probably closed. Do nothing.
			}
		} ()
		session := sessions[message.GetGUID()]

		if session == nil {
			log.Println("Message to unknow session received: ", message.GetGUID(), mtype)
			return
		}

		capsule := new(hubble.MessageCapsule)
		capsule.Mtype = mtype
		capsule.Message = message

		session <- capsule
	}()
}

func Agent(name string, key string, url string, tunnels []*Tunnel) {
	//1- intialize connection to proxy
	conn, err := hubble.NewProxyConnection(url)
	if err != nil {
		log.Fatal("Failed to connect to proxy", err)
	}
	
	//2- registration
	err = handshake(conn, name, key)
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
			switch mtype {
				case hubble.INITIATOR_MESSAGE_TYPE:
					//TODO: Make a connection to service.
					initiator := message.(*hubble.InitiatorMessage)
					//send ack (debug)
					startLocalSession(conn, initiator)
				default:
					dispatch(mtype, message.(hubble.SessionMessage))
			}
		}
	}()

	for _, tunnel := range tunnels {
		tunnel.serve(conn)
	}
}