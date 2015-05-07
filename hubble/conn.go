package hubble

import (
	"net/url"
	"net"
	"fmt"
	"errors"
	"github.com/gorilla/websocket"
)

type Connection struct {
	ws *websocket.Conn
}

var unexpectedMessageType = errors.New("Unexpected message type")
var unexpectedMessageFormat = errors.New("Only binary messages are supported")


func NewProxyConnection(proxyUrl string) (*Connection, error) {
	// open connection to proxy.
	url, _ := url.Parse(proxyUrl)
	proxy_conn, err := net.Dial("tcp", url.Host)	

	if err != nil {
		return nil, err
	}

	ws, _, err := websocket.NewClient(proxy_conn, url, nil, 1024, 1024)
	if err != nil {
		return nil, err
	}

	var connection = new(Connection)
	connection.ws = ws
	
	return connection, nil
}

func NewConnection(ws *websocket.Conn) *Connection {
	var connection = Connection {
		ws: ws,
	}

	return &connection
}


func (conn *Connection) Send(mtype uint8, message interface{}) error {
	writer, err := conn.ws.NextWriter(websocket.BinaryMessage)
	defer writer.Close()
	if err != nil {
		return err
	}

	return dumps(writer, mtype, message)
}

func (conn *Connection) Receive() (uint8, interface{}, error) {
	mode, reader, err := conn.ws.NextReader()
	if err != nil {
		return 0, nil, err
	}

	if mode != websocket.BinaryMessage {
		//only binary messages are supported.
		return 0, nil, unexpectedMessageFormat
	}

	return loads(reader)
}

func (conn *Connection) SendAckOrError(err error) error {
	return conn.Send(ACK_MESSAGE_TYPE, &AckMessage {
		Ok: err == nil,
		Message: fmt.Sprintf("%v", err),
	})
}

func (conn *Connection) ReceiveAck() error {
	//read ack.
	mtype, reply, err := conn.Receive()
	if err != nil {
		return err
	}

	if mtype != ACK_MESSAGE_TYPE {
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