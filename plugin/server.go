package plugin

import (
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/ginit"
	"google.golang.org/grpc"
	"net"
	"net/url"
	"os"
)

type Server struct {
	grpc     *grpc.Server
	listener net.Listener
}

func NewServer(cfg config.Config) (*Server, error) {
	server := &Server{
		grpc: grpc.NewServer(),
	}
	u, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "tcp":
		listener, err := net.Listen(u.Scheme, fmt.Sprintf("%s:%s", u.Hostname(), u.Port()))
		if err != nil {
			return nil, err
		}
		server.listener = listener
	case "unix":
		listener, err := net.Listen(u.Scheme, u.Path)
		if err != nil {
			return nil, err
		}
		server.listener = listener
	default:
		return nil, fmt.Errorf("bad address: %s", u)
	}
	return server, nil
}

func (s *Server) Run(reg *Registry) error {
	for _, plugin := range reg.plugins {
		if rpc, ok := plugin.(RPC); ok {
			s.grpc.RegisterService(rpc.RPC(), plugin)
		}
	}
	return s.grpc.Serve(s.listener)
}

// Handle implements the ginit.Handler interface.
func (s Server) Handle(sig os.Signal) error {
	if ginit.Terminal(sig) {
		s.grpc.Stop()
		return s.listener.Close()
	}
	return nil
}
