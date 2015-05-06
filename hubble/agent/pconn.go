package agent

import (
	"hubble"
	"net/url"
	"net"
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


func (conn *ProxyConnection) Send(mtype uint8, message interface{}) error {
	writer, err := conn.ws.NextWriter(websocket.BinaryMessage)
	defer writer.Close()
	if err != nil {
		return err
	}

	return hubble.Dumps(writer, mtype, message)
}

//Wrapper for handshake
func (conn *ProxyConnection) Initialize(agentname string, key string) error {
	message := hubble.HandshakeMessage {
		Name: agentname,
		Key: key,
	}

	return conn.Send(hubble.HANDSHAKE_MESSAGE_TYPE, &message)
}