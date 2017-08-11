package event

type EventType string

const (
	// Indicates recieving plugins
	// should be shutdown.
	SHUTDOWN = EventType("SHUTDOWN")
)

// Type returns the type of event
func (e Event) Type() EventType { return EventType(e.Id) }
