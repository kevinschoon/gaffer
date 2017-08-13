package event

import (
	"time"
)

// EventType indicates the type of event
type EventType string

const (
	// Indicates recieving plugins
	// should be shutdown.
	REQUEST_SHUTDOWN = EventType("REQUEST_SHUTDOWN")
	// Service has started
	SERVICE_STARTED = EventType("SERVICE_STARTED")
	// Service has exited
	SERVICE_EXITED = EventType("SERVICE_EXITED")
	// Request service metrics
	REQUEST_METRICS = EventType("REQUEST_METRICS")
	// Broadcasted runtime metrics
	SERVICE_METRICS = EventType("SERVICE_METRICS")
)

func New(et EventType, opts ...Option) Event {
	evt := Event{
		Type: string(et),
		Time: time.Now().Unix(),
	}
	for _, opt := range opts {
		evt = opt(evt)
	}
	return evt
}
