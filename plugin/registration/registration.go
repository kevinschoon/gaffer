package register

import (
	"context"
	"encoding/json"
	"fmt"
	etcd "github.com/coreos/etcd/clientv3"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/host"
	"github.com/mesanine/gaffer/log"
	"time"
)

const (
	RegistrationKey      = "/hosts/%s"
	DailTimeout          = 5000 * time.Millisecond
	RegistrationInterval = 5000 * time.Millisecond
	RegistrationLeaseTTL = 25
)

type Server struct {
	stop chan bool
	etcd *etcd.Client
}

func (s Server) Name() string { return "gaffer.register" }

func (s *Server) Configure(cfg config.Config) error {
	client, err := etcd.New(etcd.Config{
		Endpoints:   cfg.RegistrationServer.EtcdEndpoints,
		DialTimeout: DailTimeout,
	})
	if err != nil {
		return err
	}
	s.etcd = client
	return nil
}

func (s *Server) Run(eb *event.EventBus) error {
	ticker := time.NewTicker(RegistrationInterval)
	self, err := host.Self()
	if err != nil {
		return err
	}
	rawSelf, err := json.Marshal(self)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ticker.C:
			lease, err := s.etcd.Grant(context.TODO(), RegistrationLeaseTTL)
			if err != nil {
				log.Log.Warn(fmt.Sprintf("failed to aquire registration lease: %s", err.Error()))
				continue
			}
			_, err = s.etcd.Put(context.TODO(), fmt.Sprintf(RegistrationKey, self.Mac), string(rawSelf), etcd.WithLease(lease.ID))
			if err != nil {
				log.Log.Warn(fmt.Sprintf("failed to register: %s", err.Error()))
			}
		case <-s.stop:
			return nil
		}
	}
}

func (s *Server) Stop() error {
	s.stop <- true
	return s.etcd.Close()
}
