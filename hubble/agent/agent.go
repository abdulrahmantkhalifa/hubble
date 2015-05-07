package agent

import (
	"hubble"
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

	//read ack.
	err = conn.ReceiveAck()
	if err != nil {
		return err
	}

	return nil
}