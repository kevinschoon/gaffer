package event

import (
	"encoding/json"
	"github.com/containerd/go-runc"
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
	// Broadcasted runtime metrics
	SERVICE_METRICS = EventType("SERVICE_METRICS")
)

type Option func(Event) Event

func WithID(id string) Option {
	return func(e Event) Event {
		return Event{
			Id:    id,
			Type:  e.Type,
			Time:  e.Time,
			Stats: e.Stats,
			Spec:  e.Spec,
		}
	}
}

func WithStats(stats runc.Stats) Option {
	return func(e Event) Event {
		raw, _ := json.Marshal(stats)
		return Event{
			Stats: raw,
			Id:    e.Id,
			Type:  e.Type,
			Time:  e.Time,
			Spec:  e.Spec,
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
