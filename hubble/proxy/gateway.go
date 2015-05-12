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
	//gw.channel = make(chan *hubble.MessageCapsule, GatewayQueueSize)
	gw.channel = make(chan *hubble.MessageCapsule) //unbuffered for testing
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
	close(gw.channel)

	delete(gateways, gw.handshake.Name)

	for _, terminal := range gw.terminals {
		terminal.terminate()
	}
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

	endTerm.forward(&hubble.MessageCapsule {
		Mtype: hubble.INITIATOR_MESSAGE_TYPE,
		Message: intiator,
	})

	return nil
}

func (gw *gateway) closeSession(terminator *hubble.ConnectionClosedMessage) {
	terminal, ok := gw.terminals[terminator.GUID]
	if ok {
		//remove ref from this gateway terminals
		delete(gw.terminals, terminator.GUID)
		//remove ref from the other end terminals
		terminal.terminate()

		terminal, ok := terminal.gateway.terminals[terminator.GUID]

		if ok {
			//remove ref from this gateway terminals
			delete(gw.terminals, terminator.GUID)
			//remove ref from the other end terminals
			terminal.terminate()
		}	
	}
}

func (gw *gateway) forward(guid string, mtype uint8, message interface{}) {
	terminal, ok := gw.terminals[guid]
	if !ok {
		return
	}

	terminal.forward(&hubble.MessageCapsule{
		Mtype: mtype,
		Message: message,
	})
}

func (term terminal) forward(message *hubble.MessageCapsule) {
	defer func() {
		if err := recover(); err != nil {
			//propable channel is closed.
			//Do nothing.
		}
	}()

	term.gateway.channel <- message
}

func (term terminal) terminate() {
	term.forward(&hubble.MessageCapsule{
		Mtype: hubble.TERMINATOR_MESSAGE_TYPE,
		Message: &hubble.TerminatorMessage{
			GuidMessage: hubble.GuidMessage{term.guid},
		},
	})
}
