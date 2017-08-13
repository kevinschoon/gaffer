package event

import (
	"encoding/json"
	"github.com/containerd/go-runc"
)

// Option modifies an event
// with specific properties
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
