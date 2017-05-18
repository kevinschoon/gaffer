package client

import (
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/supervisor"
	"go.uber.org/zap"
	"os"
)

// Client handles all communication
// across a Gaffer cluster.
type Client struct {
	db    store.Store
	conns map[string]map[string]*supervisor.Client
}

func (c *Client) rpc(host *host.Host, svc *service.Service) *supervisor.Client {
	if _, ok := c.conns[host.ID]; !ok {
		c.conns[host.ID] = map[string]*supervisor.Client{}
	}
	if _, ok := c.conns[host.ID][svc.ID]; !ok {
		c.conns[host.ID][svc.ID] = &supervisor.Client{Host: host, Service: svc}
	}
	return c.conns[host.ID][svc.ID]
}

func (c Client) Services() (map[string][]*service.Service, error) {
	resp, err := c.db.Query(&query.Query{Read: &query.Read{}})
	if err != nil {
		return nil, err
	}
	return resp.Read.Cluster.Services, nil
}

func (c Client) Hosts() ([]*host.Host, error) {
	resp, err := c.db.Query(&query.Query{Read: &query.Read{}})
	if err != nil {
		return nil, err
	}
	return resp.Read.Cluster.Hosts, nil
}

func (c Client) Processes() (cluster.ProcessList, error) {
	hosts, err := c.Hosts()
	if err != nil {
		return nil, err
	}
	services, err := c.Services()
	if err != nil {
		return nil, err
	}
	pl := cluster.ProcessList{}
	for _, host := range hosts {
		if !host.Registered {
			continue
		}
		if _, ok := services[host.ID]; !ok {
			continue
		}
		// TODO concurrency
		pl[host.ID] = map[string]*os.Process{}
		for _, svc := range services[host.ID] {
			p, err := c.rpc(host, svc).Status()
			if err != nil {
				log.Log.Info(
					"could not establish client RPC connection",
					zap.Error(err),
				)
			} else {
				pl[host.ID][svc.ID] = p
			}
		}
	}
	return pl, nil
}

func NewClient(db store.Store) *Client {
	return &Client{
		db:    db,
		conns: map[string]map[string]*supervisor.Client{},
	}
}
