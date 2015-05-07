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
	channel chan *capsule
}

type capsule struct {
	mtype uint8
	message interface{}
}

var gateways = make(map[string]*gateway)

func newGateway(connection *hubble.Connection,
				handshake *hubble.HandshakeMessage) *gateway {
	gw := new(gateway)
	gw.connection = connection
	gw.handshake = handshake
	gw.terminals = make(map[string]*terminal)
	gw.channel = make(chan *capsule, GatewayQueueSize)
	return gw
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

	endTerm._forward(&capsule {
		mtype: hubble.INITIATOR_MESSAGE_TYPE,
		message: intiator,
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

	terminal._forward(&capsule{
		mtype: hubble.TERMINATOR_MESSAGE_TYPE,
		message: terminator,
	}) //TODO: fix!!
}

func (term terminal) _forward(message *capsule) {
	//push message to gateway channel
	go func() {
		term.gateway.channel <- message
	}()
}

func (term terminal) forward(data *hubble.DataMessage) {
	//push message to gateway channel
	term._forward(&capsule{
		mtype: hubble.DATA_MESSAGE_TYPE,
		message: data,
	})
}