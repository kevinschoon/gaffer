package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/user"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const PollInterval time.Duration = 10 * time.Second

type ErrClient struct {
	msg string
}

func (e ErrClient) Error() string { return fmt.Sprintf("client error: %s", e.msg) }

type Client struct {
	endpoint string
	user     *user.User
	client   *http.Client
}

func New(endpoint string, u *user.User) *Client {
	return &Client{
		endpoint: endpoint,
		user:     u,
		client:   http.DefaultClient,
	}
}

func (c Client) Query(q *query.Query) (*query.Response, error) {
	raw, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/1/query", c.endpoint), bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}
	if c.user != nil {
		req.SetBasicAuth(c.user.ID, c.user.Token)
	}
	log.Log.Info(
		"client",
		zap.String("url", req.URL.String()),
		zap.Any("query", q),
	)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	r := &query.Response{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}
	log.Log.Info(
		"client",
		zap.Int("status", resp.StatusCode),
		zap.Any("response", r),
	)
	return r, nil
}

func (c Client) Register(svcID string) (*host.Host, *service.Service, error) {
	var (
		config *cluster.Cluster
		svc    *service.Service
		self   *host.Host
	)
	resp, err := c.Query(&query.Query{Read: &query.Read{}})
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
		return nil, nil, ErrClient{"could not register with gaffer API"}
	}
	self.Update()
	resp, err = c.Query(&query.Query{Update: &query.Update{Host: self}})
	if err != nil {
		return nil, nil, err
	}
	services, ok := config.Services[self.ID]
	if !ok {
		return nil, nil, ErrClient{"no services configured for this host"}
	}
	for _, s := range services {
		if s.ID == svcID {
			svc = s
		}
	}
	if svc == nil {
		return nil, nil, ErrClient{fmt.Sprintf("could not register service %s", svcID)}
	}
	return self, svc, nil
}

func (c Client) Update(h *host.Host, svc *service.Service) error {
	h.Update()
	svc.Update()
	_, err := c.Query(&query.Query{Update: &query.Update{h, svc}})
	return err
}
