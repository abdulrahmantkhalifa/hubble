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


const HANDSHAKE_MESSAGE_HEADER_TYPE uint8 = 1
const INITIATOR_MESSAGE_HEADER_TYPE uint8 = 2
const DATA_MESSAGE_HEADER_TYPE uint8 = 3

const ACK_MESSAGE_HEADER_TYPE uint8 = 254
const ERROR_MESSAGE_HEADER_TYPE uint8 = 255


type HandshakeMessageHeader struct {
	Name string
	Key string
}

type InitiatorMessageHeader struct {

}

type DataMessageHeader struct {

}

type ErrorMessageHeader struct {
	Error string
}

type AckMessageHeader struct {

}

var MessageHeaderTypes = map[uint8]func() interface{} {
	HANDSHAKE_MESSAGE_HEADER_TYPE: func() interface{} {
		return new(HandshakeMessageHeader)
	},
	INITIATOR_MESSAGE_HEADER_TYPE: func() interface{} {
		return new(InitiatorMessageHeader)
	},
	DATA_MESSAGE_HEADER_TYPE: func() interface{} {
		return new(DataMessageHeader)
	},


	ACK_MESSAGE_HEADER_TYPE: func() interface{} {
		return new(AckMessageHeader)
	},
	ERROR_MESSAGE_HEADER_TYPE: func() interface{} {
		return new(ErrorMessageHeader)
	},
}
