package hubble

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	ws    *websocket.Conn
	rlock sync.Mutex
	wlock sync.Mutex
}

var unexpectedMessageType = errors.New("Unexpected message type")
var unexpectedMessageFormat = errors.New("Only binary messages are supported")

func NewProxyConnection(proxyUrl string, config *tls.Config) (*Connection, error) {
	// open connection to proxy.
	var dialer = websocket.Dialer{
		TLSClientConfig: config,
	}

	ws, _, err := dialer.Dial(proxyUrl, nil)

	if err != nil {
		return nil, err
	}

	var connection = new(Connection)
	connection.ws = ws

	return connection, nil
}

func NewConnection(ws *websocket.Conn) *Connection {
	var connection = Connection{
		ws: ws,
	}

	return &connection
}

func (conn *Connection) Send(message Message) error {
	conn.wlock.Lock()
	defer conn.wlock.Unlock()
	writer, err := conn.ws.NextWriter(websocket.BinaryMessage)
	defer writer.Close()
	if err != nil {
		return err
	}

	return dumps(writer, message)
}

func (conn *Connection) Receive() (Message, error) {
	conn.rlock.Lock()
	defer conn.rlock.Unlock()
	mode, reader, err := conn.ws.NextReader()
	if err != nil {
		return nil, err
	}

	if mode != websocket.BinaryMessage {
		//only binary messages are supported.
		return nil, unexpectedMessageFormat
	}

	return loads(reader)
}

func (conn *Connection) SendAckOrError(guid string, err error) error {
	ack := NewAckMessage(guid, err == nil, fmt.Sprintf("%v", err))
	return conn.Send(ack)
}

func (conn *Connection) ReceiveAck() error {
	//read ack.
	reply, err := conn.Receive()
	if err != nil {
		return err
	}

	if reply.GetMessageType() != ACK_MESSAGE_TYPE {
		return unexpectedMessageType
	}

	ack := reply.(*AckMessage)
	if !ack.Ok {
		return errors.New(ack.Message)
	}

	return nil
}

func (conn *Connection) Close() error {
	return conn.ws.Close()
}
