package supervisor

/*
import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/mesanine/gaffer/service"
	"golang.org/x/net/context"
	"net"

	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	spv  *Supervisor
	db   *store.FSStore
	port int
}

func NewServer(spv *Supervisor, cfg config.Config) *Server {
	return &Server{spv: spv, db: store.NewFSStore(cfg), port: cfg.Supervisor.Port}
}

func (s *Server) Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	resp := &StatusResponse{
		Services: map[string]*service.Service{},
		Stats:    map[string]*any.Any{},
	}
	services, err := s.db.Services()
	if err != nil {
		log.Log.Error("could not list services", zap.Error(err))
		return nil, err
	}
	for _, svc := range services {
		resp.Services[svc.Id] = svc
		stats, err := s.spv.runcs[svc.Id].Stats()
		if err != nil {
			log.Log.Error("could not get stats", zap.Error(err))
			return nil, err
		}
		raw, _ := json.Marshal(stats)
		resp.Stats[svc.Id] = &any.Any{Value: raw}
	}
	return resp, nil
}

func (s *Server) Restart(ctx context.Context, req *RestartRequest) (*RestartResponse, error) {
	if _, ok := s.spv.runcs[req.Id]; !ok {
		return nil, fmt.Errorf("unknown service %s", req.Id)
	}
	err := s.spv.runcs[req.Id].Stop()
	if err != nil {
		return nil, err
	}
	return &RestartResponse{}, nil
}

func (s *Server) Listen() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		return err
	}
	gs := grpc.NewServer()
	RegisterSupervisorServer(gs, s)
	reflection.Register(gs)
	log.Log.Info(fmt.Sprintf("launching rpc server @ 0.0.0.0:%d", s.port))
	return gs.Serve(listener)
}
*/
