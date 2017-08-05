package supervisor

import (
	"fmt"
	"github.com/mesanine/gaffer/host"
	"github.com/mesanine/gaffer/log"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
)

func NewClient(h host.Host) (*grpc.ClientConn, error) {
	address := fmt.Sprintf("%s:%d", h.Name, h.Port)
	log.Log.Debug(fmt.Sprintf("dailing %s", address))
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type ClientMux struct {
	db      host.Source
	filters []host.Filter
}

func (cm *ClientMux) Status(req *StatusRequest) (chan *StatusResponse, error) {
	hosts, err := cm.db.Get()
	if err != nil {
		return nil, err
	}
	ch := make(chan *StatusResponse)
	var wg sync.WaitGroup
	for _, h := range hosts.Hosts.Filter(cm.filters...) {
		wg.Add(1)
		go func(h *host.Host) {
			defer wg.Done()
			conn, err := NewClient(*h)
			if err != nil {
				log.Log.Error(fmt.Sprintf("cannot connect to %s: %s", h.String(), err.Error()))
				return
			}
			defer conn.Close()
			resp, err := NewSupervisorClient(conn).Status(context.Background(), req)
			if err != nil {
				log.Log.Error("bad RPC request", zap.Error(err))
				return
			}
			ch <- resp
		}(h)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch, nil
}

func (cm *ClientMux) Restart(req *RestartRequest) (chan *RestartResponse, error) {
	hosts, err := cm.db.Get()
	if err != nil {
		return nil, err
	}
	ch := make(chan *RestartResponse)
	var wg sync.WaitGroup
	for _, h := range hosts.Hosts.Filter(cm.filters...) {
		wg.Add(1)
		go func(h *host.Host) {
			defer wg.Done()
			conn, err := NewClient(*h)
			if err != nil {
				log.Log.Error(fmt.Sprintf("cannot connect to %s: %s", h.String(), err.Error()))
				return
			}
			defer conn.Close()
			resp, err := NewSupervisorClient(conn).Restart(context.Background(), req)
			if err != nil {
				log.Log.Error(fmt.Sprintf("bad request %s: %s", h.String(), err.Error()))
				return
			}
			ch <- resp
		}(h)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch, nil
}

func NewClientMux(db host.Source, filters ...host.Filter) *ClientMux {
	mux := &ClientMux{
		db:      db,
		filters: filters,
	}
	return mux
}
