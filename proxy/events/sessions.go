package events

import "github.com/g8os/hubble/auth"

type OpenSessionEvent struct {
	SourceGateway     string
	ConnectionRequest auth.ConnectionRequest

	Success bool
	Error   string
}

type CloseSessionEvent struct {
	Gateway       string
	ConnectionKey string
}
