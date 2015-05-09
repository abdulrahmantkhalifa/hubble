package proxy

import (
	"hubble"
	"log"
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(request *http.Request) bool { return true },
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

		switch mtype {
			case hubble.INITIATOR_MESSAGE_TYPE:
				initiator := message.(*hubble.InitiatorMessage)
				log.Println("New Session", initiator)
				err := gw.openSession(initiator)
				if err != nil {
					//in case local session pipe starting failes, we send 
					//error to agent. otherwise we wait for ack from
					//the other end
					conn.SendAckOrError(initiator.GUID, err)
				}
			case hubble.TERMINATOR_MESSAGE_TYPE:
				//close session.
				terminator := message.(*hubble.TerminatorMessage)
				log.Println("Ending Session", terminator)
				gw.closeSession(terminator)
			case hubble.DATA_MESSAGE_TYPE, hubble.ACK_MESSAGE_TYPE:
				//just forward
				msg := message.(hubble.SessionMessage)
				gw.forward(msg.GetGUID(), mtype, message)
			default:
				log.Println("Unknown message type:", mtype)
		}
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