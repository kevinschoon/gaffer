package supervisor

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"net/rpc"
)

type Client struct {
	Hostname string
	Port     int
	conn     *rpc.Client
}

func (c *Client) client() (*rpc.Client, error) {
	if c.conn == nil {
		address := fmt.Sprintf("%s:%d", c.Hostname, c.Port)
		log.Log.Debug(fmt.Sprintf("dailing %s", address))
		conn, err := rpc.DialHTTP("tcp", address)
		if err != nil {
			return nil, err
		}
		c.conn = conn
	}
	return c.conn, nil
}

func (c Client) call(method string, req, resp interface{}) error {
	conn, err := c.client()
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Call(method, req, resp)
	log.Log.Debug(
		"rpc call",
		zap.String("method", method),
		zap.Any("request", req),
		zap.Any("response", resp),
		zap.Error(err),
	)
	return err
}

func (c Client) Status() (*StatusResponse, error) {
	resp := &StatusResponse{}
	err := c.call("Supervisor.Status", StatusRequest{}, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c Client) Restart() (*RestartResponse, error) {
	resp := &RestartResponse{}
	err := c.call("Supervisor.Restart", RestartRequest{}, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c Client) Update(svc *cluster.Service) (*UpdateResponse, error) {
	resp := &UpdateResponse{}
	err := c.call("Supervisor.Update", UpdateRequest{svc}, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
