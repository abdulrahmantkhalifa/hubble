package events

import "time"

type GatewayRegistrationEvent struct {
	Time    time.Time
	Gateway string
}

type GatewayUnregistrationEvent struct {
	Time    time.Time
	Gateway string
}
