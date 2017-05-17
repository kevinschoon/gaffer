package store

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/store/query"
	"math/rand"
	"time"
)

const (
	MAX_PORT int = 65535
	MIN_PORT int = 49152
)

func assignPort(services []*service.Service) int {
	rand.Seed(time.Now().Unix())
	port := rand.Intn(MAX_PORT-MIN_PORT) + MIN_PORT
	for _, svc := range services {
		if svc.Port == port {
			return assignPort(services)
		}
	}
	return port
}

func Register(store Store, id string) (*host.Host, *service.Service, error) {
	var (
		config *cluster.Cluster
		svc    *service.Service
		self   *host.Host
	)
	resp, err := store.Query(&query.Query{Read: &query.Read{}})
	if err != nil {
		return nil, nil, err
	}
	config = resp.Read.Cluster
	for _, h := range config.Hosts {
		if err := h.Register(); err == nil {
			self = h
			break
		}
	}
	if self == nil {
		return nil, nil, fmt.Errorf("could not register with gaffer API")
	}
	self.Update()
	services, ok := config.Services[self.ID]
	if !ok {
		return nil, nil, fmt.Errorf("no services configured for this host")
	}
	for _, s := range services {
		if s.ID == id {
			svc = s
		}
	}
	if svc == nil {
		return nil, nil, fmt.Errorf("could not register service %s", id)
	}
	// Assign random port if not already configured
	if svc.Port == 0 {
		svc.Port = assignPort(config.ServicesFlat())
	}
	svc.Registered = true
	_, err = store.Query(&query.Query{Update: &query.Update{Host: self, Service: svc}})
	return self, svc, err
}
