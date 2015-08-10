package events

type GatewayRegistrationEvent struct {
	Gateway string
}

type GatewayUnregistrationEvent struct {
	Gateway string
}
