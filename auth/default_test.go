package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcceptAllModule(t *testing.T) {
	request := validConnectionRequest()

	Install(NewAcceptAllModule())
	flag, err := Connect(request)
	assert.True(t, flag, "Expected connection request to be granted.")
	assert.Nil(t, err, "Did not expect error: %v", err)
}

func TestDeclineAllModule(t *testing.T) {
	request := validConnectionRequest()

	Install(NewDeclineAllModule())
	flag, err := Connect(request)
	assert.False(t, flag, "Expected connection request to be denied.")
	assert.Nil(t, err, "Did not expect error: %v", err)
}
