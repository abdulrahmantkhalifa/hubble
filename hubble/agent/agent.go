package agent

import (
	"hubble"
	"log"
)


//Handshake
func Handshake(conn *hubble.Connection, agentname string, key string) error {
	message := hubble.HandshakeMessage {
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

func Dispatch(mtype uint8, message hubble.SessionMessage) {
	go func() {
		session := sessions[message.GetGUID()]

		if session == nil {
			log.Println("Message to unknow session received: ", session)
			return
		}

		capsule := new(hubble.MessageCapsule)
		capsule.Mtype = mtype
		capsule.Message = message

		session <- capsule
	}()
}