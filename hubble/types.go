package hubble

import (
	"net"
	"fmt"
)

type Tunnel struct {
	Local uint16
	Ip net.IP
	Remote uint16
	Gateway string
}

type Session struct {
	tunnel *Tunnel
	guid [16]byte
}


const HANDSHAKE_MESSAGE_TYPE uint8 = 1
const INITIATOR_MESSAGE_TYPE uint8 = 2
const DATA_MESSAGE_TYPE uint8 = 3
const TERMINATOR_MESSAGE_TYPE uint8 = 4

const ACK_MESSAGE_TYPE uint8 = 255

type HandshakeMessage struct {
	Name string
	Key string
}

type InitiatorMessage struct {
	GUID string
	Ip net.IP
	Port uint16
	Gatename string
}

type DataMessage struct {
	GUID string
	Data []byte
}

type TerminatorMessage struct {
	GUID string
}

type AckMessage struct {
	Ok bool
	Message string
}

func (msg *InitiatorMessage) String() string {
	return fmt.Sprintf("%v->%v:%v id(%v)", msg.Gatename, msg.Ip, msg.Port, msg.GUID)
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
