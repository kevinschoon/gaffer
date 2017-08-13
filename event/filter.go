package event

// Filter filters an event based
// on a specific property
type Filter func(Event) bool

func ID(id string) Filter {
	return func(evt Event) bool {
		return evt.Id == id
	}
}

func Is(et EventType) Filter {
	return func(evt Event) bool {
		return EventType(evt.Type) == et
	}
}
