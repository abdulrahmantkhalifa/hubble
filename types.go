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

//Agent level messages
type Message interface {
	GetMessageType() MessageType
}

type SessionMessage interface {
	Message
	GetGUID() string
}

//GuidMessage
type GuidMessage struct {
	GUID string
}

func (msg *GuidMessage) GetGUID() string {
	return msg.GUID
}

//Handshake message
type HandshakeMessage struct {
	Version string
	Name string
	Key string
}

func (msg *HandshakeMessage) GetMessageType() MessageType {
	return HANDSHAKE_MESSAGE_TYPE
}

//Initiator message
type InitiatorMessage struct {
	GuidMessage
	Ip net.IP
	Port uint16
	Gatename string
}

func (msg *InitiatorMessage) GetMessageType() MessageType {
	return INITIATOR_MESSAGE_TYPE
}

//Data message
type DataMessage struct {
	GuidMessage
	Order int64
	Data []byte
}

func (msg *DataMessage) GetMessageType() MessageType {
	return DATA_MESSAGE_TYPE
}

//Terminator message
type TerminatorMessage struct {
	GuidMessage
}

func (msg *TerminatorMessage) GetMessageType() MessageType {
	return TERMINATOR_MESSAGE_TYPE
}

type ConnectionClosedMessage TerminatorMessage

func (msg *ConnectionClosedMessage) GetMessageType() MessageType {
	return CONNECTION_CLOSED_MESSAGE_TYPE
}

//Ack message
type AckMessage struct {
	GuidMessage
	Ok bool
	Message string
}

func (msg *AckMessage) GetMessageType() MessageType {
	return ACK_MESSAGE_TYPE
}


func (msg *InitiatorMessage) String() string {
	return fmt.Sprintf("%v->%v:%v id(%v)", msg.Gatename, msg.Ip, msg.Port, msg.GUID)
}

func (msg *TerminatorMessage) String() string {
	return fmt.Sprintf("id(%v)", msg.GUID)
}

func NewHandshakeMessage(name string, key string) *HandshakeMessage {
	return &HandshakeMessage {
		Version: PROTOCOL_VERSION_0_1,
		Name: name,
		Key: key,
	}
}

func NewInitiatorMessage(guid string, ip net.IP, port uint16, gatename string) *InitiatorMessage {
	return &InitiatorMessage {
		GuidMessage: GuidMessage{guid},
		Ip: ip,
		Port: port,
		Gatename: gatename,
	}
}

func NewDataMessage(guid string, order int64, data []byte) *DataMessage {
	return &DataMessage{
		GuidMessage: GuidMessage{guid},
		Order: order,
		Data: data,
	}
}

func NewTerminatorMessage(guid string) *TerminatorMessage {
	return &TerminatorMessage{
		GuidMessage: GuidMessage{guid},
	}
}

func NewConnectionClosedMessage(guid string) *ConnectionClosedMessage {
	return &ConnectionClosedMessage {
		GuidMessage: GuidMessage{guid},
	}
}

func NewAckMessage(guid string, ok bool, message string) *AckMessage {
	return &AckMessage {
		GuidMessage: GuidMessage{guid},
		Ok: ok,
		Message: message,
	}
}

var messageTypes = map[MessageType]func() Message {
	HANDSHAKE_MESSAGE_TYPE: func() Message {
		return &HandshakeMessage {}
	},
	INITIATOR_MESSAGE_TYPE: func() Message {
		return &InitiatorMessage {}
	},
	DATA_MESSAGE_TYPE: func() Message {
		return &DataMessage {}
	},
	TERMINATOR_MESSAGE_TYPE: func() Message {
		return &TerminatorMessage {}
	},
	CONNECTION_CLOSED_MESSAGE_TYPE: func() Message {
		return &ConnectionClosedMessage{}
	},
	ACK_MESSAGE_TYPE: func() Message {
		return &AckMessage {}
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