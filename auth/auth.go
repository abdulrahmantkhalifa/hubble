package auth

import "errors"

// ConnectionRequest holds information about a client that wants to open a
// tunnel to the specified destination.
type ConnectionRequest struct {
	RemoteHost string // Destination host
	RemotePort uint16 // Destination port
	Gatename   string // Destination gate
	Key        string // Authorization key
}

// An AuthorizationModule should accept or decline an authorization request.
type AuthorizationModule interface {
	// Accept or decline the given authorization request
	Connect(r ConnectionRequest) (bool, error)
}

var installedAuthorizationModule AuthorizationModule

// Use the specified authorization module for authorizing requests
func Install(module AuthorizationModule) {
	installedAuthorizationModule = module
}

// Authorize or decline a connection request
func Connect(r ConnectionRequest) (bool, error) {
	if installedAuthorizationModule == nil {
		return false, errors.New("No authorization module installed.")
	}

	return installedAuthorizationModule.Connect(r)
}
