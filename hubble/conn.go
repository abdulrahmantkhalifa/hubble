package hubble

import (
	"net/url"
	"net"
	"errors"
	"github.com/gorilla/websocket"
)

type ProxyConnection struct {
	ws *websocket.Conn
}

func NewProxyConnection(proxyUrl string) (*ProxyConnection, error) {
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

	var connection = new(ProxyConnection)
	connection.ws = ws
	
	return connection, nil
}

func NewConnection(ws *websocket.Conn) *ProxyConnection {
	var connection = ProxyConnection {
		ws: ws,
	}

	return &connection
}


func (conn *ProxyConnection) Send(mtype uint8, message interface{}) error {
	writer, err := conn.ws.NextWriter(websocket.BinaryMessage)
	defer writer.Close()
	if err != nil {
		return err
	}

	return Dumps(writer, mtype, message)
}

func (conn *ProxyConnection) Receive() (uint8, interface{}, error) {
	mode, reader, err := conn.ws.NextReader()
	if err != nil {
		return 0, nil, err
	}

	if mode != websocket.BinaryMessage {
		//only binary messages are supported.
		return 0, nil, errors.New("Only binary messages are supported")
	}

	return Loads(reader)
}

func (conn *ProxyConnection) Close() error {
	return conn.ws.Close()
}