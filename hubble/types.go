package hubble

import (
	"net"
	"fmt"
)

const HANDSHAKE_MESSAGE_TYPE uint8 = 1
const INITIATOR_MESSAGE_TYPE uint8 = 2
const DATA_MESSAGE_TYPE uint8 = 3
const TERMINATOR_MESSAGE_TYPE uint8 = 4

const ACK_MESSAGE_TYPE uint8 = 255


type MessageCapsule struct {
	Mtype uint8
	Message interface{}
}

//Agent level messages

//Handshake message
type HandshakeMessage struct {
	Name string
	Key string
}

//Session level messages

//GuidMessage
type GuidMessage struct {
	GUID string
}

type SessionMessage interface {
	GetGUID() string
}

func (msg *GuidMessage) GetGUID() string {
	return msg.GUID
}

//Initiator message
type InitiatorMessage struct {
	GuidMessage
	Ip net.IP
	Port uint16
	Gatename string
}

//Data message
type DataMessage struct {
	GuidMessage
	Data []byte
}

//Terminator message
type TerminatorMessage struct {
	GuidMessage
}

//Ack message
type AckMessage struct {
	GuidMessage
	Ok bool
	Message string
}


func (msg *InitiatorMessage) String() string {
	return fmt.Sprintf("%v->%v:%v id(%v)", msg.Gatename, msg.Ip, msg.Port, msg.GUID)
}

func (msg *TerminatorMessage) String() string {
	return fmt.Sprintf("id(%v)", msg.GUID)
}

var MessageTypes = map[uint8]func() interface{} {
	HANDSHAKE_MESSAGE_TYPE: func() interface{} {
		return new(HandshakeMessage)
	},
	INITIATOR_MESSAGE_TYPE: func() interface{} {
		return new(InitiatorMessage)
	},
	DATA_MESSAGE_TYPE: func() interface{} {
		return new(DataMessage)
	},
	TERMINATOR_MESSAGE_TYPE: func() interface{} {
		return new(TerminatorMessage)
	},
	ACK_MESSAGE_TYPE: func() interface{} {
		return new(AckMessage)
	},
}
