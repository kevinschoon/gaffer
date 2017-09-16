package client

import (
	"context"
	"encoding/json"
	"fmt"
	etcd "github.com/coreos/etcd/clientv3"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/host"
	"github.com/mesanine/gaffer/log"
	"time"
)

const (
	RegistrationKey      = "gaffer_host_"
	DailTimeout          = 5 * time.Second
	RegistrationLeaseTTL = 60
)

// Client is an HTTP client for
// interacting with a Gaffer cluster.
type Client struct {
	etcd *etcd.Client
}

func New(cfg config.Config) (*Client, error) {
	cli, err := etcd.New(etcd.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: DailTimeout,
	})
	if err != nil {
		return nil, err
	}
	return &Client{etcd: cli}, nil
}

func (s Client) Close() error { return s.etcd.Close() }

func (c Client) Register() error {
	self, err := host.Self()
	if err != nil {
		return err
	}
	raw, err := json.Marshal(self)
	if err != nil {
		return err
	}
	lease, err := c.etcd.Grant(context.TODO(), RegistrationLeaseTTL)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s_%s", RegistrationKey, self.Mac)
	_, err = c.etcd.Put(context.TODO(), key, string(raw), etcd.WithLease(lease.ID))
	if err != nil {
		return err
	}
	log.Log.Debug(fmt.Sprintf("registered self: %s", key))
	return nil
}

func (c Client) Hosts() ([]*host.Host, error) {
	resp, err := c.etcd.Get(context.TODO(), RegistrationKey, etcd.WithPrefix(), etcd.WithSort(etcd.SortByKey, etcd.SortDescend))
	if err != nil {
		return nil, err
	}
	hosts := []*host.Host{}
	for _, kv := range resp.Kvs {
		host := &host.Host{}
		err = json.Unmarshal(kv.Value, host)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}
