package events

import "github.com/Jumpscale/hubble/auth"

type OpenSessionEvent struct {
	SourceGateway     string
	ConnectionRequest auth.ConnectionRequest

	Success bool
	Error   string
}
