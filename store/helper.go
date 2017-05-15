package store

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/store/query"
)

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
	resp, err = store.Query(&query.Query{Update: &query.Update{Host: self}})
	if err != nil {
		return nil, nil, err
	}
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
	return self, svc, nil
}

func Update(store Store, h *host.Host, svc *service.Service) error {
	h.Update()
	svc.Update()
	_, err := store.Query(&query.Query{Update: &query.Update{h, svc}})
	return err
}
