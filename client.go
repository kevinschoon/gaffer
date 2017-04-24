package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const PollInterval time.Duration = 10 * time.Second

type Client struct {
	endpoint string
	token    string
	client   *http.Client
	log      *zap.Logger
}

func NewClient(endpoint, token string, logger *zap.Logger) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
		log:      logger,
		client:   http.DefaultClient,
	}
}

func (c Client) query(q *Query) (*Response, error) {
	raw, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/1/cluster", c.endpoint), bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("", c.token)
	c.log.Info(
		"client",
		zap.String("url", req.URL.String()),
		zap.Any("query", q),
	)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	r := &Response{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}
	c.log.Info(
		"client",
		zap.Int("status", resp.StatusCode),
		zap.Any("response", r),
	)
	return r, nil
}

func (c Client) Cluster(id string) (*Cluster, error) {
	resp, err := c.query(&Query{Type: READ})
	if err != nil {
		return nil, err
	}
	for _, cluster := range resp.Clusters {
		if cluster.ID == id {
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("%s not found", id)
}

func (c Client) Update(cluster *Cluster) error {
	_, err := c.query(&Query{Type: UPDATE, Cluster: cluster})
	return err
}

// UntilZKReady waits until a Zookeeper quorum is formed
func (c Client) UntilZKReady(cluster *Cluster) error {
	return backoff.Retry(func() error {
		rc, err := c.Cluster(cluster.ID)
		if err != nil {
			c.log.Warn("client", zap.String("error", err.Error()))
			return err
		}
		if !rc.ZKReady() {
			return fmt.Errorf("zookeepers still converging")
		}
		c.log.Info("client", zap.String("msg", "Zookeepers converged!"))
		return nil
	}, backoff.NewExponentialBackOff())
}

// UntilMasterReady waits until a Mesos Master quorum is formed
func (c Client) UntilMasterReady(cluster *Cluster) error {
	return backoff.Retry(func() error {
		rc, err := c.Cluster(cluster.ID)
		if err != nil {
			c.log.Warn("client", zap.String("error", err.Error()))
			return err
		}
		if !rc.MesosReady() {
			return fmt.Errorf("mesos still converging")
		}
		c.log.Info("client", zap.String("msg", "Mesos Masters converged!"))
		return nil
	}, backoff.NewExponentialBackOff())
}
