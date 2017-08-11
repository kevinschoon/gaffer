package event

import "time"

// EventType indicates the type of event
type EventType string

const (
	// Indicates recieving plugins
	// should be shutdown.
	SHUTDOWN = EventType("SHUTDOWN")
	// Service has started
	SERVICE_STARTED = EventType("SERVICE_STARTED")
	// Service has exited
	SERVICE_EXITED = EventType("SERVICE_EXITED")
)

type Option func(Event) Event

func WithID(id string) Option {
	return func(e Event) Event {
		return Event{
			Id:   id,
			Type: e.Type,
			Time: e.Time,
		}
	}
}

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

func (e Event) Is(et EventType) bool { return EventType(e.Type) == et }
