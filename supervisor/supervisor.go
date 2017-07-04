package supervisor

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/vektorlab/gaffer/config"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/runc"
	"github.com/vektorlab/gaffer/store"
	"time"
)

const BackoffInterval = 1000 * time.Millisecond

type Supervisor struct {
	db       *store.FSStore
	runcs    map[string]*runc.Runc
	cancelCh []chan bool
	config   config.Config
}

func New(db *store.FSStore, cfg config.Config) (*Supervisor, error) {
	services, err := db.Services()
	if err != nil {
		return nil, err
	}
	runcs := map[string]*runc.Runc{}
	for _, service := range services {
		runcs[service.Id] = runc.New(service.Id, service.Bundle, cfg)
	}
	return &Supervisor{db: db, runcs: runcs, config: cfg}, nil
}

func (s *Supervisor) Init() error {
	bootSvcs, err := s.db.OnBoot()
	if err != nil {
		return err
	}
	for _, boot := range bootSvcs {
		_, err := runc.New(boot.Id, boot.Bundle, s.config).Run()
		if err != nil {
			return err
		}
	}
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
