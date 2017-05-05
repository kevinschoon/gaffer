package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/user"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const PollInterval time.Duration = 10 * time.Second

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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/1/cluster", c.endpoint), bytes.NewBuffer(raw))
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

func (c Client) Cluster(id string) (*cluster.Cluster, error) {
	q := &query.Query{Type: query.READ}
	q.Read.ID = id
	resp, err := c.Query(q)
	if err != nil {
		return nil, err
	}
	return resp.Clusters[0], nil
}

func (c Client) Update(o *cluster.Cluster) error {
	q := &query.Query{Type: query.UPDATE}
	q.Update.Clusters = []*cluster.Cluster{o}
	_, err := c.Query(q)
	return err
}
