package auth

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectWithoutModule(t *testing.T) {
	Install(nil)

	request := validConnectionRequest()
	flag, err := Connect(request)
	assert.False(t, flag, "Should decline connection requests.")
	assert.NotNil(t, err, "Expected 'no module installed' error.")
}

func TestConnectWithValidModule(t *testing.T) {
	request := validConnectionRequest()

	Install(newValidTestAuthModule(true, errors.New("test error")))
	flag, err := Connect(request)
	assert.True(t, flag, "Expected connection request to be granted.")
	assert.NotNil(t, err, "Expected error to be passed on.")
}

func validConnectionRequest() ConnectionRequest {
	return ConnectionRequest{
		IP:       net.ParseIP("127.0.0.1"),
		Port:     22,
		Gatename: "target",
		Key:      "key",
	}
}

type validTestAuthModule struct {
	grant bool
	err   error
}

func newValidTestAuthModule(shouldGrant bool, err error) AuthorizationModule {
	return validTestAuthModule{
		grant: shouldGrant,
		err:   err,
	}
}

func (auth validTestAuthModule) Connect(r ConnectionRequest) (bool, error) {
	return auth.grant, auth.err
}
