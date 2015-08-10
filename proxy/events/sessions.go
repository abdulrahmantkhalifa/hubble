package events

import (
	"time"

	"github.com/Jumpscale/hubble/auth"
)

type OpenSessionEvent struct {
	Time              time.Time
	SourceGateway     string
	ConnectionRequest auth.ConnectionRequest

	Success bool
	Error   string
}
