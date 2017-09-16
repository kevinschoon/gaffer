package metrics

import (
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	"go.uber.org/zap"
)

type Metrics struct {
	err  chan error
	stop chan bool
}

func New() *Metrics {
	return &Metrics{
		err:  make(chan error, 1),
		stop: make(chan bool, 1),
	}
}

func (m Metrics) Name() string { return "metrics" }

func (m Metrics) Configure(cfg config.Config) error { return nil }

func (m Metrics) Run(e *event.EventBus) error {
	sub := event.NewSubscriber()
	e.Subscribe(sub)
	ec := sub.Chan()
	for {
		select {
		case evt := <-ec:
			if event.Is(event.SERVICE_METRICS)(evt) {
				// TODO TODO
				log.Log.Info("processing metric event", zap.Any("event", evt))
			}
		case err := <-m.err:
			return err
		case <-m.stop:
			return nil
		}
	}
}

func (m Metrics) Stop() error {
	m.stop <- true
	return nil
}
