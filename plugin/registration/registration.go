package register

import (
	"fmt"
	"github.com/mesanine/gaffer/client"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	"time"
)

const RegistrationInterval = 25 * time.Second

type Server struct {
	stop   chan bool
	client *client.Client
}

func (s Server) Name() string { return "gaffer.register" }

func (s *Server) Configure(cfg config.Config) error {
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	s.stop = make(chan bool, 1)
	s.client = cli
	return nil
}

func (s *Server) Run(eb *event.EventBus) error {
	ticker := time.NewTicker(RegistrationInterval)
	for {
		select {
		case <-ticker.C:
			err := s.client.Register()
			if err != nil {
				log.Log.Error(fmt.Sprintf("failed to register self: %s", err.Error()))
			}
		case <-s.stop:
			return nil
		}
	}
}

func (s *Server) Stop() error {
	s.stop <- true
	return s.client.Close()
}
