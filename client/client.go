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
	RegistrationKey      = "/hosts/%s"
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
		Endpoints:   cfg.Etcd.Endpoints,
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
	key := fmt.Sprintf(RegistrationKey, self.Mac)
	_, err = c.etcd.Put(context.TODO(), key, string(raw), etcd.WithLease(lease.ID))
	if err != nil {
		return err
	}
	log.Log.Debug(fmt.Sprintf("registered self: %s", key))
	return nil
}
