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
	handshake *hubble.HandshakeMessage
	connection *hubble.ProxyConnection
}

var gateways = make(map[string]*Gateway)


func register(gateway *Gateway) error {
	//1- Authentication
	//TODO:

	//2- Registration
	log.Println(fmt.Sprintf("Registering gateway: %v", gateway.handshake.Name))
	gateways[gateway.handshake.Name] = gateway
	
	return nil
}

func unregister(gateway *Gateway) {
	log.Println(fmt.Sprintf("Unegistering gateway: %v", gateway.handshake.Name))
	delete(gateways, gateway.handshake.Name)
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

	if mtype != hubble.HANDSHAKE_MESSAGE_TYPE {
		log.Println(fmt.Sprintf("Expecting handshake message, got %v", mtype))
		return
	}

	handshake := message.(*hubble.HandshakeMessage)

	var gateway = Gateway{
		handshake: handshake,
		connection: conn,
	}

	err = register(&gateway)
	ack := hubble.AckMessage {
		Ok: err == nil,
		Message: fmt.Sprintf("%v", err),
	}

	log.Println(ack)
	conn.Send(hubble.ACK_MESSAGE_TYPE, ack)
	
	if err != nil {
		return
	}

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