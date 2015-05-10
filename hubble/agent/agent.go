package agent

import (
	"hubble"
	"log"
)


//Handshake
func Handshake(conn *hubble.Connection, agentname string, key string) error {
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

func Dispatch(mtype uint8, message hubble.SessionMessage) {
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