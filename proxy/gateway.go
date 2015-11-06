package proxy

import (
	"errors"
	"fmt"
	"log"

	"github.com/Jumpscale/hubble"
	"github.com/Jumpscale/hubble/auth"
	"github.com/Jumpscale/hubble/logging"
	"github.com/Jumpscale/hubble/proxy/events"
)

const GatewayQueueSize = 512

var unauthenticated = errors.New("Unauthenticated")
var unauthorized = errors.New("Unauthorized")
var gatewayNotRegistered = errors.New("Gateway not registered")

type terminal struct {
	guid          string
	gateway       *gateway
	connectionKey string
}

type gateway struct {
	handshake  *hubble.HandshakeMessage
	connection *hubble.Connection
	terminals  map[string]*terminal
	channel    chan hubble.Message
}

var gateways = make(map[string]*gateway)

func newGateway(connection *hubble.Connection,
	handshake *hubble.HandshakeMessage) *gateway {
	gw := new(gateway)
	gw.connection = connection
	gw.handshake = handshake
	gw.terminals = make(map[string]*terminal)
	gw.channel = make(chan hubble.Message)
	return gw
}

func (gw *gateway) String() string {
	return gw.handshake.Name
}

func (gw *gateway) register() error {
	//1- Authentication
	//TODO:

	//2- Registration
	logging.LogEvent(events.GatewayRegistrationEvent{
		Gateway: gw.handshake.Name,
	})

	log.Println(fmt.Sprintf("Registering gateway: %v", gw.handshake.Name))
	gateways[gw.handshake.Name] = gw

	return nil
}

func (gw *gateway) unregister() {
	logging.LogEvent(events.GatewayUnregistrationEvent{
		Gateway: gw.handshake.Name,
	})
	log.Println(fmt.Sprintf("Unegistering gateway: %v", gw.handshake.Name))
	close(gw.channel)

	delete(gateways, gw.handshake.Name)

	for _, terminal := range gw.terminals {
		terminal.terminate()
	}
}

func (gw *gateway) openSession(intiator *hubble.InitiatorMessage) error {
	req := auth.ConnectionRequest{
		RemoteHost: intiator.RemoteHost,
		RemotePort: intiator.RemotePort,
		Gatename:   intiator.Gatename,
		Key:        intiator.Key,
	}
	ok, err := auth.Connect(req)
	if err != nil {
		log.Println("auth error:", err)
		logging.LogEvent(events.OpenSessionEvent{
			SourceGateway:     gw.handshake.Name,
			ConnectionRequest: req,
			Success:           false,
			Error:             "Authorziation error: " + err.Error(),
		})
		return errors.New("Authorization error.")
	}
	if !ok {
		log.Println("Session authorization request declined")
		logging.LogEvent(events.OpenSessionEvent{
			SourceGateway:     gw.handshake.Name,
			ConnectionRequest: req,
			Success:           false,
			Error:             "Unauthorized",
		})
		return errors.New("Unauthorized")
	}

	logging.LogEvent(events.OpenSessionEvent{
		SourceGateway:     gw.handshake.Name,
		ConnectionRequest: req,
		Success:           true,
	})

	endGw, ok := gateways[intiator.Gatename]
	if !ok {
		return gatewayNotRegistered
	}

	endTerm := new(terminal)
	endTerm.guid = intiator.GUID
	endTerm.gateway = endGw
	endTerm.connectionKey = intiator.Key

	gw.terminals[intiator.GUID] = endTerm

	startTerm := new(terminal)
	startTerm.guid = intiator.GUID
	startTerm.gateway = gw
	startTerm.connectionKey = intiator.Key

	endGw.terminals[intiator.GUID] = startTerm

	endTerm.forward(intiator)

	return nil
}

func (gw *gateway) closeSession(terminator *hubble.ConnectionClosedMessage) {
	terminal, ok := gw.terminals[terminator.GUID]

	if ok {
		logging.LogEvent(events.CloseSessionEvent{
			Gateway:       gw.handshake.Name,
			ConnectionKey: terminal.connectionKey,
		})

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

func (gw *gateway) forward(guid string, message hubble.Message) {
	terminal, ok := gw.terminals[guid]
	if !ok {
		return
	}

	terminal.forward(message)
}

func (term terminal) forward(message hubble.Message) {
	defer func() {
		if err := recover(); err != nil {
			//channel is closed.
			//Do nothing.
		}
	}()

	term.gateway.channel <- message
}

func (term terminal) terminate() {
	terminate := hubble.NewTerminatorMessage(term.guid)
	term.forward(terminate)
}
