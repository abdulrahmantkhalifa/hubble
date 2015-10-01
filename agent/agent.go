package agent

import (
	"crypto/tls"
	"github.com/Jumpscale/hubble"
	"log"
)

type OnExit func(Agent, error)

type Agent interface {
	Proxy() string
	Name() string
	Key() string
	Start(OnExit) error
	Stop() error
	AddTunnel(*Tunnel) error
	RemoveTunnel(*Tunnel)
	Tunnels() []*Tunnel
}

type agentImpl struct {
	name      string
	key       string
	proxy     string
	tlsConfig *tls.Config
	tunnels   map[string]*Tunnel
	conn      *hubble.Connection
	sessions  sessionsStore
}

func NewAgent(proxy string, name string, key string, tlsConfig *tls.Config) Agent {
	return &agentImpl{
		proxy:     proxy,
		name:      name,
		key:       key,
		tlsConfig: tlsConfig,
		tunnels:   make(map[string]*Tunnel),
		sessions:  make(sessionsStore),
	}
}

func (agent *agentImpl) Proxy() string {
	return agent.proxy
}

func (agent *agentImpl) Name() string {
	return agent.name
}

func (agent *agentImpl) Key() string {
	return agent.key
}

//Handshake
func (agent *agentImpl) handshake() error {

	message := hubble.NewHandshakeMessage(agent.name, agent.key)

	err := agent.conn.Send(message)
	if err != nil {
		return err
	}

	//Next message must be an ack message, read ack.
	err = agent.conn.ReceiveAck()
	if err != nil {
		return err
	}

	return nil
}

func (agent *agentImpl) dispatch(message hubble.SessionMessage) {
	defer func() {
		if err := recover(); err != nil {
			//Can't send data to session channel?!. Please don't panic, chill out and
			//relax, it's probably closed. Do nothing.
		}
	}()
	session, ok := agent.sessions[message.GetGUID()]

	if !ok {
		if message.GetMessageType() != hubble.TERMINATOR_MESSAGE_TYPE {
			log.Println("Message to unknow session received: ", message.GetGUID(), message.GetMessageType())
		}

		return
	}

	session <- message
}

func (agent *agentImpl) Start(onExit OnExit) (err error) {
	//1- intialize connection to proxy
	conn, err := hubble.NewProxyConnection(agent.proxy, agent.tlsConfig)

	defer func() {
		if err != nil && onExit != nil {
			onExit(agent, err)
		}
	}()

	if err != nil {
		return
	}
	agent.conn = conn
	//2- registration
	err = agent.handshake()
	if err != nil {
		return
	}

	go func() {
		//receive all messages.
		log.Println("Start receiving loop")
		err = nil
		defer func() {
			for _, tunnel := range agent.tunnels {
				tunnel.stop()
			}

			if onExit != nil {
				onExit(agent, err)
			}
		}()

		for {
			var message hubble.Message
			message, err = conn.Receive()
			if err != nil {
				//we should check error types to take a decistion. for now just exit
				log.Println("Receive loop failed", err)
				return
			}
			switch message.GetMessageType() {
			case hubble.INITIATOR_MESSAGE_TYPE:
				initiator := message.(*hubble.InitiatorMessage)
				startLocalSession(agent.sessions, conn, initiator)
			default:
				agent.dispatch(message.(hubble.SessionMessage))
			}
		}
	}()

	for _, tunnel := range agent.tunnels {
		tunnel.start(agent.sessions, agent.conn)
	}

	return nil
}

func (agent *agentImpl) Stop() (err error) {
	//close connection, force receive loop to die.
	if agent.conn != nil {
		agent.conn.Close()
		agent.conn = nil
		for _, tunnel := range agent.tunnels {
			tunnel.stop()
		}
		agent.tunnels = make(map[string]*Tunnel)
	}

	return nil
}

func (agent *agentImpl) AddTunnel(tunnel *Tunnel) error {
	if _, ok := agent.tunnels[tunnel.String()]; ok {
		//not re-adding tunnel
		return nil
	}

	agent.tunnels[tunnel.String()] = tunnel

	if agent.conn != nil {
		return tunnel.start(agent.sessions, agent.conn)
	}

	return nil
}

func (agent *agentImpl) RemoveTunnel(tunnel *Tunnel) {
	if tunnel, ok := agent.tunnels[tunnel.String()]; ok {
		defer delete(agent.tunnels, tunnel.String())

		tunnel.stop()
	}
}

func (agent *agentImpl) Tunnels() []*Tunnel {
	tunnels := make([]*Tunnel, 0, len(agent.tunnels))
	for _, tunnel := range agent.tunnels {
		tunnels = append(tunnels, tunnel)
	}

	return tunnels
}
