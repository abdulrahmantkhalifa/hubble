package proxy

import (
	"errors"
	"fmt"
	"github.com/Jumpscale/hubble"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var invalidProtocolVersion = errors.New("Invalid protocol version")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(request *http.Request) bool { return true },
}

func initiatorMessage(gw *gateway, message hubble.Message) {
	initiator := message.(*hubble.InitiatorMessage)
	log.Println("New Session", initiator)
	err := gw.openSession(initiator)
	if err != nil {
		//in case local session pipe starting failes, we send
		//error to agent. otherwise we wait for ack from
		//the other end
		gw.connection.SendAckOrError(initiator.GUID, err)
	}
}

func connectionClosedMessage(gw *gateway, message hubble.Message) {
	terminator := message.(*hubble.ConnectionClosedMessage)
	log.Println("Ending Session:", gw, terminator)
	gw.closeSession(terminator)
}

func forward(gw *gateway, message hubble.Message) {
	msg := message.(hubble.SessionMessage)
	gw.forward(msg.GetGUID(), message)
}

var messageHandlers = map[hubble.MessageType]func(*gateway, hubble.Message){
	hubble.INITIATOR_MESSAGE_TYPE:         initiatorMessage,
	hubble.CONNECTION_CLOSED_MESSAGE_TYPE: connectionClosedMessage,
	hubble.DATA_MESSAGE_TYPE:              forward,
	hubble.ACK_MESSAGE_TYPE:               forward,
}

func handler(ws *websocket.Conn, request *http.Request) {
	conn := hubble.NewConnection(ws)
	defer conn.Close()

	//1- Handshake
	message, err := conn.Receive()

	if err != nil {
		return
	}

	if message.GetMessageType() != hubble.HANDSHAKE_MESSAGE_TYPE {
		log.Println(fmt.Sprintf("Expecting handshake message, got %v", message.GetMessageType()))
		return
	}

	handshake := message.(*hubble.HandshakeMessage)

	if handshake.Version != hubble.PROTOCOL_VERSION_0_1 {
		conn.SendAckOrError("", invalidProtocolVersion)
		return
	}

	gw := newGateway(conn, handshake)

	err = gw.register()
	conn.SendAckOrError("", err)

	if err != nil {
		return
	}

	defer gw.unregister()

	go func() {
		for {
			msgCap := <-gw.channel
			if msgCap == nil {
				//channel has been closed
				return
			}

			err := conn.Send(msgCap)
			if err != nil {
				log.Println("Failed to forward message to gateway:", gw)
			}
		}
	}()

	//dispatch loop
	for {
		message, err := conn.Receive()
		if err != nil {
			break
		}

		msgHandler, ok := messageHandlers[message.GetMessageType()]
		if !ok {
			log.Println("Unknown message type:", message.GetMessageType())
			continue
		}

		msgHandler(gw, message)
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
