package supervisor

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/runc"
	"github.com/mesanine/gaffer/store"
	"go.uber.org/zap"
	"time"
)

const (
	BackoffInterval = 1000 * time.Millisecond
	StatsInterval   = 2000 * time.Millisecond
)

type Supervisor struct {
	runcs  map[string]*runc.Runc
	cancel map[string]context.CancelFunc
	config config.Config
	err    chan error
	stop   chan bool
}

func (s *Supervisor) Name() string { return "gaffer.supervisor" }

func (s *Supervisor) Configure(cfg config.Config) error {
	services, err := store.New(cfg).Services()
	if err != nil {
		return err
	}
	s.runcs = map[string]*runc.Runc{}
	for _, svc := range services {
		s.runcs[svc.Id] = runc.New(svc.Id, svc.Bundle, cfg)
	}
	s.cancel = map[string]context.CancelFunc{}
	s.err = make(chan error, 1)
	s.stop = make(chan bool, 1)
	s.config = cfg
	return nil
}

func (s *Supervisor) Run(eb *event.EventBus) error {
	// Launch all registered containers
	s.init(eb)
	sub := event.NewSubscriber()
	eb.Subscribe(sub)
	defer eb.Unsubscribe(sub)
	evtCh := sub.Chan()
	for {
		select {
		case <-s.stop:
			return nil
		case evt := <-evtCh:
			switch {
			case event.Is(event.REQUEST_METRICS)(evt):
				for name, rc := range s.runcs {
					stats, err := rc.Stats()
					if err != nil {
						log.Log.Warn(fmt.Sprintf("failed to collect stats from %s: %s", name, err.Error()))
						continue
					}
					eb.Push(event.New(
						event.SERVICE_METRICS,
						event.WithID(name),
						event.WithStats(*stats),
					))
				}
			}
		}
	}
}

func (s *Supervisor) Stop() error {
	for name, cancelFn := range s.cancel {
		// Cancel each runc backoff context
		// causing each container to not be
		// restarted when killed.
		cancelFn()
		if err := s.runcs[name].Stop(); err != nil {
			// If we can't stop a container we will log it but continue
			// trying since the entire process is being shutdown.
			log.Log.Error(fmt.Sprintf("failed to cancel service %s: %s", name, err.Error()))
		} else {
			log.Log.Warn(fmt.Sprintf("killed service %s", name))
		}
	}
	// Signial stop to the Run() function
	s.stop <- true
	return nil
}

// Runc returns the underlying runc instance with id
func (s *Supervisor) Runc(id string) (*runc.Runc, error) {
	if rc, ok := s.runcs[id]; ok {
		return rc, nil
	}
	return nil, fmt.Errorf("no service with id %s exists", id)
}

func (s *Supervisor) init(eb *event.EventBus) {
	for name, rc := range s.runcs {
		if _, ok := s.cancel[name]; ok {
			panic(fmt.Sprintf("container %s was already registered", name))
		}
		ctx, cancelFn := context.WithCancel(context.Background())
		s.cancel[name] = cancelFn
		go func(ctx context.Context, rc *runc.Runc, name string) {
			s.err <- backoff.RetryNotify(
				func() error {
					log.Log.Info(fmt.Sprintf("launching runc container %s", name))
					eb.Push(
						event.New(
							event.SERVICE_STARTED,
							event.WithID(name),
						),
					)
					code, err := rc.Run()
					var msg string
					if err != nil {
						msg = err.Error()
					}
					return fmt.Errorf("container %s exited with code %d: %s", name, code, msg)
				},
				backoff.WithContext(backoff.NewConstantBackOff(BackoffInterval), ctx),
				func(err error, d time.Duration) {
					eb.Push(event.New(
						event.SERVICE_EXITED,
						event.WithID(name),
					))
					log.Log.Warn(err.Error(), zap.Duration("runtime", d))
				},
			)
		}(ctx, rc, name)
	}

}

// MonitorFuncs returns backoff.Operation and backoff.Notify
// functions. The operation function is re-ran each time the
// underlying runc container exits.
func MonitorFuncs(id string, rc *runc.Runc) (backoff.Operation, backoff.Notify) {
	return func() error {
		log.Log.Info(fmt.Sprintf("Launching %s", id))
		rc.Delete()
		code, err := rc.Run()
		var msg string
		if err != nil {
			msg = err.Error()
		}
		log.Log.Info(fmt.Sprintf("Service %s exited with code %d: %s", id, code, msg))
		return fmt.Errorf("container exited")
	}, func(err error, d time.Duration) {}
}
