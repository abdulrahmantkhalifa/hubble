package proxy

import (
	"hubble"
	"log"
	"fmt"
	"net/http"
	"errors"
	"github.com/gorilla/websocket"
)

var invalidProtocolVersion = errors.New("Invalid protocol version")

var upgrader = websocket.Upgrader {
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(request *http.Request) bool { return true },
}


func initiatorMessage(gw *gateway, mtype uint8, message interface{}) {
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

func connectionClosedMessage(gw *gateway, mtype uint8, message interface{}) {
	terminator := message.(*hubble.ConnectionClosedMessage)
	log.Println("Ending Session:", gw, terminator)
	gw.closeSession(terminator)
}

func forward(gw *gateway, mtype uint8, message interface{}) {
	msg := message.(hubble.SessionMessage)
	gw.forward(msg.GetGUID(), mtype, message)
}

var messageHandlers = map[uint8] func (*gateway, uint8, interface{}) {
	hubble.INITIATOR_MESSAGE_TYPE: initiatorMessage,
	hubble.CONNECTION_CLOSED_MESSAGE_TYPE: connectionClosedMessage,
	hubble.DATA_MESSAGE_TYPE: forward,
	hubble.ACK_MESSAGE_TYPE: forward,
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
			msgCap := <- gw.channel
			if msgCap == nil {
				//channel has been closed
				return
			}

			err := conn.Send(msgCap.Mtype, msgCap.Message)
			if err != nil {
				log.Println("Failed to forward message to gateway:", gw)
			}
		}
	}()

	//dispatch loop
	for {
		mtype, message, err := conn.Receive()
		if err != nil {
			log.Println(err)
			break
		}

		msgHandler, ok := messageHandlers[mtype]
		if !ok {
			log.Println("Unknown message type:", mtype)
			continue
		}

		msgHandler(gw, mtype, message)
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