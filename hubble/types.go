package hubble

import (
	"net"
	"fmt"
	"errors"
)

type MessageType uint8

const PROTOCOL_VERSION_0_1 string = "0.1"

const INVALID_MESSAGE_TYPE MessageType = 0
const HANDSHAKE_MESSAGE_TYPE MessageType = 1
const INITIATOR_MESSAGE_TYPE MessageType = 2
const DATA_MESSAGE_TYPE MessageType = 3
const TERMINATOR_MESSAGE_TYPE MessageType = 4
const CONNECTION_CLOSED_MESSAGE_TYPE MessageType = 5

const ACK_MESSAGE_TYPE MessageType = 255

var unknownMessageType = errors.New("Unknow message type")



// type MessageCapsule struct {
// 	Mtype uint8
// 	Message interface{}
// }

//Agent level messages
type Message interface {
	GetMessageType() MessageType
}

type TypedMessage struct {
	mtype MessageType
}


func (msg *TypedMessage) GetMessageType() MessageType {
	return msg.mtype
}

type SessionMessage interface {
	Message
	GetGUID() string
}

//GuidMessage
type GuidMessage struct {
	GUID string
}

//Handshake message
type HandshakeMessage struct {
	TypedMessage
	Version string
	Name string
	Key string
}

func (msg *GuidMessage) GetGUID() string {
	return msg.GUID
}

//Initiator message
type InitiatorMessage struct {
	TypedMessage
	GuidMessage
	Ip net.IP
	Port uint16
	Gatename string
}

//Data message
type DataMessage struct {
	TypedMessage
	GuidMessage
	Order int64
	Data []byte
}

//Terminator message
type TerminatorMessage struct {
	TypedMessage
	GuidMessage
}

type ConnectionClosedMessage TerminatorMessage

//Ack message
type AckMessage struct {
	TypedMessage
	GuidMessage
	Ok bool
	Message string
}

func (msg *AckMessage) GetMessageType() MessageType {
	return msg.mtype
}

func (msg *InitiatorMessage) String() string {
	return fmt.Sprintf("%v->%v:%v id(%v)", msg.Gatename, msg.Ip, msg.Port, msg.GUID)
}

func (msg *TerminatorMessage) String() string {
	return fmt.Sprintf("id(%v)", msg.GUID)
}

func NewHandshakeMessage(name string, key string) *HandshakeMessage {
	return &HandshakeMessage {
		TypedMessage: TypedMessage{HANDSHAKE_MESSAGE_TYPE},
		Version: PROTOCOL_VERSION_0_1,
		Name: name,
		Key: key,
	}
}

func NewInitiatorMessage(guid string, ip net.IP, port uint16, gatename string) *InitiatorMessage {
	return &InitiatorMessage {
		TypedMessage: TypedMessage{INITIATOR_MESSAGE_TYPE},
		GuidMessage: GuidMessage{guid},
		Ip: ip,
		Port: port,
		Gatename: gatename,
	}
}

func NewDataMessage(guid string, order int64, data []byte) *DataMessage {
	return &DataMessage{
		TypedMessage: TypedMessage{DATA_MESSAGE_TYPE},
		GuidMessage: GuidMessage{guid},
		Order: order,
		Data: data,
	}
}

func NewTerminatorMessage(guid string) *TerminatorMessage {
	return &TerminatorMessage{
		TypedMessage: TypedMessage{TERMINATOR_MESSAGE_TYPE},
		GuidMessage: GuidMessage{guid},
	}
}

func NewConnectionClosedMessage(guid string) *ConnectionClosedMessage {
	return &ConnectionClosedMessage {
		TypedMessage: TypedMessage{CONNECTION_CLOSED_MESSAGE_TYPE},
		GuidMessage: GuidMessage{guid},
	}
}

func NewAckMessage(guid string, ok bool, message string) *AckMessage {
	return &AckMessage {
		TypedMessage: TypedMessage{ACK_MESSAGE_TYPE},
		GuidMessage: GuidMessage{guid},
		Ok: ok,
		Message: message,
	}
}

var messageTypes = map[MessageType]func() Message {
	HANDSHAKE_MESSAGE_TYPE: func() Message {
		return &HandshakeMessage {
			TypedMessage: TypedMessage{HANDSHAKE_MESSAGE_TYPE},
		}
	},
	INITIATOR_MESSAGE_TYPE: func() Message {
		return &InitiatorMessage {
			TypedMessage: TypedMessage{INITIATOR_MESSAGE_TYPE},
		}
	},
	DATA_MESSAGE_TYPE: func() Message {
		return &DataMessage {
			TypedMessage: TypedMessage{DATA_MESSAGE_TYPE},
		}
	},
	TERMINATOR_MESSAGE_TYPE: func() Message {
		return &TerminatorMessage {
			TypedMessage: TypedMessage{TERMINATOR_MESSAGE_TYPE},
		}
	},
	CONNECTION_CLOSED_MESSAGE_TYPE: func() Message {
		return &ConnectionClosedMessage{
			TypedMessage: TypedMessage{CONNECTION_CLOSED_MESSAGE_TYPE},
		}
	},
	ACK_MESSAGE_TYPE: func() Message {
		return &AckMessage {
			TypedMessage: TypedMessage{ACK_MESSAGE_TYPE},
		}
	},
}


func NewMessage(mtype MessageType) (Message, error) {
	initiator, ok := messageTypes[mtype]
	
	if !ok {
		return nil, unknownMessageType
	}

	msg := initiator()
	return msg, nil
}