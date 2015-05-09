package proxy

import (
	"hubble"
	"errors"
	"log"
	"fmt"
)

const GatewayQueueSize = 512

var unauthenticated = errors.New("Unauthenticated")
var unauthorized = errors.New("Unauthorized")
var gatewayNotRegistered = errors.New("Gateway not registered")

type terminal struct {
	guid string
	gateway *gateway
}

type gateway struct {
	handshake *hubble.HandshakeMessage
	connection *hubble.Connection
	terminals map[string]*terminal
	channel chan *hubble.MessageCapsule
}


var gateways = make(map[string]*gateway)

func newGateway(connection *hubble.Connection,
				handshake *hubble.HandshakeMessage) *gateway {
	gw := new(gateway)
	gw.connection = connection
	gw.handshake = handshake
	gw.terminals = make(map[string]*terminal)
	gw.channel = make(chan *hubble.MessageCapsule, GatewayQueueSize)
	return gw
}

func (gw *gateway) String() string {
	return gw.handshake.Name
}

func (gw *gateway) register() error {
	//1- Authentication
	//TODO:

	//2- Registration
	log.Println(fmt.Sprintf("Registering gateway: %v", gw.handshake.Name))
	gateways[gw.handshake.Name] = gw
	
	return nil
}

func (gw *gateway) unregister() {
	log.Println(fmt.Sprintf("Unegistering gateway: %v", gw.handshake.Name))
	delete(gateways, gw.handshake.Name)

	//TODO: Close all hanging sessions
}

func (gw *gateway) serve() {

}

func (gw *gateway) openSession(intiator *hubble.InitiatorMessage) error {
	endGw, ok := gateways[intiator.Gatename]
	if !ok {
		return gatewayNotRegistered
	}

	endTerm := new(terminal)
	endTerm.guid = intiator.GUID
	endTerm.gateway = endGw

	gw.terminals[intiator.GUID] = endTerm

	startTerm := new(terminal)
	startTerm.guid = intiator.GUID
	startTerm.gateway = gw

	endGw.terminals[intiator.GUID] = startTerm

	endTerm._forward(&hubble.MessageCapsule {
		Mtype: hubble.INITIATOR_MESSAGE_TYPE,
		Message: intiator,
	})

	return nil
}

func (gw *gateway) closeSession(terminator *hubble.TerminatorMessage) {
	terminal, ok := gw.terminals[terminator.GUID]
	if !ok {
		return
	}

	//remove ref from this gateway terminals
	defer delete(gw.terminals, terminator.GUID)
	//remove ref from the other end terminals
	defer delete(terminal.gateway.terminals, terminator.GUID)

	terminal._forward(&hubble.MessageCapsule{
		Mtype: hubble.TERMINATOR_MESSAGE_TYPE,
		Message: terminator,
	}) //TODO: fix!!
}

func (gw *gateway) forward(guid string, mtype uint8, message interface{}) {
	terminal, ok := gw.terminals[guid]
	if !ok {
		return
	}

	terminal._forward(&hubble.MessageCapsule{
		Mtype: mtype,
		Message: message,
	})
}

func (term terminal) _forward(message *hubble.MessageCapsule) {
	//push message to gateway channel
	go func() {
		term.gateway.channel <- message
	}()
}
