package register

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/mesanine/gaffer/client"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	"time"
)

const RegistrationInterval = 25 * time.Second

type Server struct {
	err    chan error
	stop   chan bool
	config config.Config
}

func New() *Server {
	return &Server{
		err:  make(chan error, 1),
		stop: make(chan bool, 1),
	}
}

func (s Server) Name() string { return "register" }

func (s *Server) Configure(cfg config.Config) error {
	s.config = cfg
	return nil
}

func (s *Server) Run(eb *event.EventBus) error {
	go func() {
		s.err <- backoff.RetryNotify(func() error {
			var cli *client.Client
			defer func() {
				if cli != nil {
					cli.Close()
				}
			}()
			for {
				if cli == nil {
					c, err := client.New(s.config)
					if err != nil {
						return err
					}
					cli = c
				}
				err := cli.Register()
				if err != nil {
					return err
				}
				time.Sleep(RegistrationInterval)
			}
		}, backoff.NewConstantBackOff(1*time.Second),
			func(err error, d time.Duration) {
				log.Log.Warn(fmt.Sprintf("failed to register with etcd: %s", err.Error()))
			},
		)
	}()
	select {
	case err := <-s.err:
		return err
	case <-s.stop:
		return nil
	}
}

func (s *Server) Stop() error {
	s.stop <- true
	return nil
}
