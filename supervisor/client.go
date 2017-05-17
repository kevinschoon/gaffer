package supervisor

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"net/rpc"
	"os"
)

type Client struct {
	Host    *host.Host
	Service *service.Service
	conn    *rpc.Client
}

func (c *Client) client() (*rpc.Client, error) {
	if c.conn == nil {
		address := fmt.Sprintf("%s:%d", c.Host.Hostname, c.Service.Port)
		log.Log.Info(
			"supervisor.client",
			zap.String("message", fmt.Sprintf("dailing %s", address)),
		)
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
	err = conn.Call(method, req, resp)
	log.Log.Debug(
		"supervisor.client",
		zap.String("message", "RPC call"),
		zap.String("method", method),
		zap.Any("request", req),
		zap.Any("response", resp),
		zap.Error(err),
	)
	return err
}

func (c Client) Status() (*os.Process, error) {
	resp := &StatusResponse{}
	err := c.call("Supervisor.Status", StatusRequest{}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Process, nil
}

func (c Client) Restart() (*os.Process, error) {
	resp := RestartResponse{}
	err := c.call("Supervisor.Restart", RestartRequest{}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Process, nil
}

func (c Client) Update(svc *service.Service) (*os.Process, error) {
	resp := &UpdateResponse{}
	err := c.call("Supervisor.Update", UpdateRequest{svc}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Process, nil
}
