package logging

import "log"

type Event interface{}

// EventLogger is an interface that accepts events. You can make your types
// adhere to this interface to process events.
type EventLogger interface {
	Log(event Event) error
}

var eventLogger EventLogger = defaultEventLogger{}

// InstallEventLogger replaces the current event logger.
func InstallEventLogger(logger EventLogger) {
	if logger == nil {
		eventLogger = defaultEventLogger{}
	} else {
		eventLogger = logger
	}
}

// LogEvent makes the currently installed event logger process the event.
func LogEvent(event Event) error {
	log.Printf("Logging event: %T%v", event, event)
	err := eventLogger.Log(event)
	if err != nil {
		log.Print("Could not log event:", err)
	}
	return err
}

type defaultEventLogger struct{}

func (defaultEventLogger) Log(event Event) error {
	return nil
}
