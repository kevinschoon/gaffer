package supervisor

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/runc"
	"github.com/mesanine/gaffer/service"
	"time"
)

const BackoffInterval = 1000 * time.Millisecond

type Supervisor struct {
	runcs    map[string]*runc.Runc
	cancelCh []chan bool
	config   config.Config
}

func New(services []*service.Service, cfg config.Config) (*Supervisor, error) {
	runcs := map[string]*runc.Runc{}
	for _, svc := range services {
		runcs[svc.Id] = runc.New(svc.Id, svc.Bundle, cfg)
	}
	return &Supervisor{runcs: runcs, config: cfg}, nil
}

func (s *Supervisor) Init() error {
	s.cancelCh = []chan bool{}
	for id, rc := range s.runcs {
		ch := make(chan bool)
		s.cancelCh = append(s.cancelCh, ch)
		go monitor(ch, MonitorFunc(id, rc))
	}
	return nil
}

func monitor(ch chan bool, fn func() error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		backoff.Retry(fn, backoff.WithContext(backoff.NewConstantBackOff(BackoffInterval), ctx))
	}()
	<-ch
}

func MonitorFunc(id string, rc *runc.Runc) func() error {
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
	}
}
