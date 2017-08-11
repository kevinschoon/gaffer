package event

import (
	"github.com/mesanine/gaffer/log"
	"go.uber.org/zap"
)

const BufferSize = 128

type Subscriber struct {
	e chan Event
}

func NewSubscriber() Subscriber {
	return Subscriber{
		e: make(chan Event, BufferSize),
	}
}

// Next returns the next event in the
// subscriber event channel. If the
// channel was closed by the EventBus
// it will return nil.
func (s Subscriber) Next() *Event {
	evt, ok := <-s.e
	if !ok {
		return nil
	}
	return &evt
}

func (s Subscriber) Chan() <-chan Event { return s.e }

// EventBus is an event demultiplexer where
// events are submitted via a call to Push
// and sent to all subscribed listeners.
// The bus is only as fast as the slowest
// Subscriber.
type EventBus struct {
	running     bool
	shutdown    chan bool
	events      chan Event
	subscribe   chan Subscriber
	unsubscribe chan Subscriber
	subscribers map[Subscriber]bool
}

func NewEventBus() *EventBus {
	return &EventBus{
		events:      make(chan Event, BufferSize),
		shutdown:    make(chan bool),
		subscribe:   make(chan Subscriber),
		unsubscribe: make(chan Subscriber),
		subscribers: map[Subscriber]bool{},
	}
}

func (b *EventBus) broadcast(e Event) {
	// Range each subscriber and attempt
	// to publish the event to it.
	log.Log.Debug(
		"broadcasting event",
		zap.Int("subscribers", len(b.subscribers)),
		zap.Int("size", len(b.events)),
		zap.Any("event", e),
	)
	for sub, _ := range b.subscribers {
		// If a subscriber is not listening
		// or it's buffer is full the entire
		// eventbus will block. TODO: Consider
		// adding an optional Gaurentee flag
		// depending on the type of subscriber.
		sub.e <- e
	}
}

func (b *EventBus) run() {
loop:
	for {
		select {
		case <-b.shutdown:
			break loop
		case event := <-b.events:
			b.broadcast(event)
		case sub := <-b.subscribe:
			b.subscribers[sub] = true
		case sub := <-b.unsubscribe:
			delete(b.subscribers, sub)
		}
	}
}

// Subscribe adds a new subscriber to the EventBus
func (b *EventBus) Subscribe(sub Subscriber) {
	b.subscribe <- sub
}

// Unsubscribe removes a subscriber from the EventBus
func (b *EventBus) Unsubscribe(sub Subscriber) {
	b.unsubscribe <- sub
}

// Push a new event into the EventBus, it will be
// broadcasted to all listening subscribers.
func (b *EventBus) Push(event Event) {
	b.events <- event
}

func (b *EventBus) Start() {
	if b.running {
		return
	}
	b.running = true
	go b.run()
}

func (b *EventBus) Stop() {
	if !b.running {
		return
	}
	b.shutdown <- true
	for sub, _ := range b.subscribers {
		close(sub.e)
	}
	b.running = false
}
