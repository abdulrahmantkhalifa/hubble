package logging

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventLogger(t *testing.T) {
	event := "event"

	logger := &testEventLogger{}
	InstallEventLogger(logger)

	err := LogEvent(event)
	assert.Nil(t, err, "Did not expect error: %v", err)
	assert.Equal(t, logger.event, event, "Events don't match.")

	logger.err = errors.New("test error")
	err = LogEvent(event)
	assert.Equal(t, err, logger.err, "Errors don't match.")
}

type testEventLogger struct {
	event Event
	err   error
}

func (logger *testEventLogger) Log(event Event) error {
	logger.event = event
	return logger.err
}
