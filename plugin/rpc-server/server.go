package server

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/service"
	"golang.org/x/net/context"
	"net"

	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	db   *store.FSStore
	eb   *event.EventBus
	port int
	stop chan bool
}

func (s *Server) Name() string { return "gaffer.rpc-server" }

func (s *Server) Configure(cfg config.Config) error {
	s.db = store.NewFSStore(cfg)
	s.eb = nil
	s.port = cfg.RPCServer.Port
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

func (s *Server) Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	resp := &StatusResponse{
		Services: map[string]*service.Service{},
		Stats:    map[string]*any.Any{},
	}
	return resp, nil
}

func (s *Server) Restart(ctx context.Context, req *RestartRequest) (*RestartResponse, error) {
	return &RestartResponse{}, nil
}
