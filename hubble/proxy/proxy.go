package proxy

import (
	"hubble"
	"log"
	"fmt"
	"errors"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(request *http.Request) bool { return true },
}

var unauthenticated = errors.New("Unauthenticated")
var unauthorized = errors.New("Unauthorized")


type Gateway struct {
	handshake *hubble.HandshakeMessageHeader
	connection *hubble.ProxyConnection
}

var gateways = make(map[string]*Gateway)


func register(gateway *Gateway) error {
	//1- Authentication
	log.Println(fmt.Sprintf("Registering gateway: %v", gateway.handshake.Name))
	gateways[gateway.handshake.Name] = gateway
	return nil
}

func unregister(gateway *Gateway) {
	log.Println(fmt.Sprintf("Unegistering gateway: %v", gateway.handshake.Name))
	delete(gateways, gateway.handshake.Name)
}

func readNextMessage(conn *websocket.Conn) (uint8, interface{}, error) {
	mode, reader, err := conn.NextReader()
	if err != nil {
		return 0, nil, err
	}

	if mode != websocket.BinaryMessage {
		//only binary messages are supported.
		return 0, nil, errors.New("Only binary messages are supported")
	}

	return hubble.Loads(reader)
}

func handler(ws *websocket.Conn, request *http.Request) {
	conn := hubble.NewConnection(ws)
	defer conn.Close()

	//1- Handshake
	mtype, message, err := conn.Receive()
	
	if err != nil {
		log.Println(err)
		return
	}

	if mtype != hubble.HANDSHAKE_MESSAGE_HEADER_TYPE {
		log.Println(fmt.Sprintf("Expecting handshake message, got %v", mtype))
		return
	}

	handshake := message.(*hubble.HandshakeMessageHeader)

	var gateway = Gateway{
		handshake: handshake,
		connection: conn,
	}

	err = register(&gateway)
	//if registration went fine, proceed with reading the rest of the messages.
	//else we should send a proper error message to agent.
	defer unregister(&gateway)

	//Read loop
	for {
		mtype, message, err := conn.Receive()
		if err != nil {
			break
		}
		// switch mtype {
		// 	case hubble.INITIATOR_MESSAGE_TYPE
		// }
		log.Println(mtype, message, err)
	}
}
//The http handler for the websockets
func ProxyHandler(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
        log.Println(err)
        return
    }

	go handler(conn, request)
}