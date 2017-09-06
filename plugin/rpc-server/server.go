package server

import (
	"fmt"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/runc"
	"github.com/mesanine/gaffer/service"
	"github.com/mesanine/gaffer/store"
	"golang.org/x/net/context"
	"net"

	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// RuncFn is a function to lookup a runc.Runc
// presumably running on this host.
type RuncFn func(string) (*runc.Runc, error)

type Server struct {
	db   *store.FSStore
	eb   *event.EventBus
	port int
	stop chan bool
	runc RuncFn
}

func (s *Server) Name() string { return "gaffer.rpc-server" }

func (s *Server) Configure(cfg config.Config) error {
	s.db = store.New(cfg, "services")
	s.eb = nil
	s.port = cfg.Plugins.RPCServer.Port
	s.stop = make(chan bool, 1)
	return nil
}

func (s *Server) Run(eb *event.EventBus) error {
	s.eb = eb
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		return err
	}
	gs := grpc.NewServer()
	RegisterRPCServer(gs, s)
	reflection.Register(gs)
	log.Log.Info(fmt.Sprintf("launching rpc server @ 0.0.0.0:%d", s.port))
	errCh := make(chan error, 1)
	go func(listener net.Listener, gs *grpc.Server) {
		defer listener.Close()
		errCh <- gs.Serve(listener)
	}(listener, gs)
	select {
	case err := <-errCh:
		return err
	case <-s.stop:
	}
	return nil
}

func (s *Server) Stop() error {
	s.stop <- true
	return nil
}

// SetRuncFn sets the function to get a
// runc.Runc instance.
func (s *Server) SetRuncFn(fn RuncFn) {
	s.runc = fn
}

func (s *Server) Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	resp := &StatusResponse{
		Services: []*service.Service{},
	}
	services, err := s.db.Services()
	if err != nil {
		return nil, err
	}
	for _, svc := range services {
		rc, err := s.runc(svc.Id)
		if err != nil {
			return nil, err
		}
		stats, err := rc.Stats()
		if err != nil {
			return nil, err
		}
		svc = service.WithStats(*stats)(svc)
		resp.Services = append(resp.Services, &svc)
	}
	return resp, nil
}

func (s *Server) Restart(ctx context.Context, req *RestartRequest) (*RestartResponse, error) {
	rc, err := s.runc(req.Id)
	if err != nil {
		return nil, err
	}
	// Kill the underlying runc app
	// causing the supervisor to start it again.
	err = rc.Stop()
	if err != nil {
		return nil, err
	}
	return &RestartResponse{}, nil
}
