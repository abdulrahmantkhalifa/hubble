package hubble

import (
	"net"
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

type AckMessage struct {
	Ok bool
	Message string
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
	ACK_MESSAGE_TYPE: func() interface{} {
		return new(AckMessage)
	},
}
