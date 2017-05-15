package http

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Endpoint string
	User     *user.User
	client   *http.Client
}

func New(endpoint string, u *user.User) *Client {
	return &Client{
		Endpoint: endpoint,
		User:     u,
		client:   http.DefaultClient,
	}
}

func (c Client) Query(q *query.Query) (*query.Response, error) {
	raw, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/1/query", c.Endpoint), bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}
	if c.User != nil {
		req.SetBasicAuth(c.User.ID, c.User.Token)
	}
	log.Log.Debug(
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
	log.Log.Debug(
		"client",
		zap.Int("status", resp.StatusCode),
		zap.Any("response", r),
	)
	return r, nil
}

func (c Client) Close() error { return nil }
