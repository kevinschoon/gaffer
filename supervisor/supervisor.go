package supervisor

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type Token string

const PollTime = 200 * time.Millisecond

// Supervisor is an RPC service that wraps
// a service process. It allows checking
// the wrapped process status and pushing
// a new service configuration. If the new
// configuration fails to start after a few
// seconds it will return an error and revert
// to the previous configuration.
type Supervisor struct {
	token    Token
	db       *store.Store
	svc      *cluster.Service
	proc     *Process
	started  time.Time
	shutdown chan bool
	status   chan struct {
		proc chan *Process
	}
	restart chan struct {
		err chan error
	}
	update chan struct {
		service *cluster.Service
		err     chan error
	}
}

func NewSupervisor(token Token, db *store.Store) (*Supervisor, error) {
	seed, err := db.GetService()
	if err != nil {
		return nil, err
	}
	var proc *Process
	if seed != nil {
		p, err := NewProcess(seed)
		if err != nil {
			return nil, err
		}
		proc = p
	}
	return &Supervisor{
		db:       db,
		token:    token,
		svc:      seed,
		proc:     proc,
		shutdown: make(chan bool),
		status: make(chan struct {
			proc chan *Process
		}),
		restart: make(chan struct {
			err chan error
		}),
		update: make(chan struct {
			service *cluster.Service
			err     chan error
		}),
	}, nil
}

func (s *Supervisor) callStatus() *Process {
	proc := make(chan *Process)
	s.status <- struct {
		proc chan *Process
	}{proc}
	return <-proc
}

func (s *Supervisor) callUpdate(svc *cluster.Service) error {
	err := make(chan error)
	s.update <- struct {
		service *cluster.Service
		err     chan error
	}{svc, err}
	return <-err
}

func (s *Supervisor) callRestart() error {
	err := make(chan error)
	s.restart <- struct {
		err chan error
	}{err}
	return <-err
}

// monitor continuously checks the status of the
// underlying process and attempts to start it
// if it is not. Should only be called once.
func (s *Supervisor) monitor() {

	// Periodically monitor the process to ensure
	// it continues running.
	ticker := time.NewTicker(PollTime)
loop:
	for {
		select {
		case <-s.shutdown:
			break loop
		case <-ticker.C:
			s.ensureRunning()
		case e := <-s.restart:
			if s.proc == nil {
				e.err <- fmt.Errorf("no process configured")
			} else {
				e.err <- s.proc.Restart()
			}
		case u := <-s.update:
			u.err <- s.replace(u.service)
		case p := <-s.status:
			p.proc <- s.proc
		}
	}
}

func (s *Supervisor) ensureRunning() {
	// no process configured, noop
	if s.proc == nil {
		return
	}
	if !s.proc.Running() {
		log.Log.Info(
			"process is not running",
			zap.Any("service", s.svc),
			zap.Any("process", s.proc),
		)
		exp := backoff.NewExponentialBackOff()
		exp.MaxElapsedTime = 30 * time.Second
		err := backoff.RetryNotify(
			func() error {
				proc, err := NewProcess(s.svc)
				if err != nil {
					return err
				}
				s.proc = proc
				return s.proc.Start()
			},
			exp,
			backoff.Notify(func(err error, d time.Duration) {
				log.Log.Info(
					"process failed to start",
					zap.Duration("duration", d),
					zap.Any("service", s.svc),
					zap.Any("process", s.proc),
					zap.Error(err),
				)
			}))
		if err != nil {
			log.Log.Warn(
				"process has timed out",
				zap.Any("service", s.svc),
				zap.Any("process", s.proc),
				zap.Error(err),
			)
		}
	} else {
		// Process configured and running
		log.Log.Debug(
			s.svc.ID,
			zap.String("message", "process running normally"),
			zap.Any("service", s.svc),
			zap.Any("process", s.proc),
		)
	}
}

// replace attempts to replace a possibly running
// process with a new service configuration.
// If the new configuration causes the process
// to fail it will revert to the old configuration.
func (s *Supervisor) replace(svc *cluster.Service) error {
	// The existing process configuration
	previous := s.proc
	revert := func() {
		if previous == nil {
			// No previous process
			return
		}
		if !previous.Running() {
			err := previous.Start()
			if err != nil {
				log.Log.Error(
					s.svc.ID,
					zap.String("message", "failed to recover previous service configuration"),
					zap.Error(err),
				)
			} else {
				s.proc = previous
				log.Log.Warn(
					s.svc.ID,
					zap.String("message", "recovered previous service configuration"),
					zap.Any("process", s.proc),
				)
			}
		}
	}
	var proc *Process
	// Create a new process and tmp path
	if len(svc.Args) > 0 {
		p, err := NewProcess(svc)
		if err != nil {
			return err
		}
		proc = p
	}
	// A Process was already configured
	if s.proc != nil {
		// Process is actively running
		if s.proc.Running() {
			// Attempt to stop the active process
			err := s.proc.Stop()
			if err != nil {
				return err
			}
		}
	}
	// Attempt to start the process
	if proc != nil {
		err := proc.Start()
		if err != nil {
			revert()
			return err
		}
		// TODO: Perhaps there is a better way
		// to wait for the process to have a
		// chance to launch. Poll the process
		// once per second for five seconds to
		// ensure it is started and is not is
		// not flapping.
		for i := 0; i < 5; i++ {
			time.Sleep(1 * time.Second)
			if !proc.Running() {
				revert()
				return fmt.Errorf("process failed to start with new configuration")
			}
		}
	}
	// Finally replace the old process and service configuration
	s.proc = proc
	s.svc = svc
	s.db.SetService(svc)
	log.Log.Info(
		svc.ID,
		zap.String("message", "process updated successfully"),
		zap.Any("process", s.proc),
		zap.Any("service", s.svc),
	)
	return nil
}

type StatusRequest struct{}
type StatusResponse struct {
	Pid     int              `json:"pid"`
	Uptime  time.Duration    `json:"uptime"`
	Service *cluster.Service `json:"service"`
}

func (s *Supervisor) Status(req StatusRequest, resp *StatusResponse) error {
	if proc := s.callStatus(); proc != nil {
		resp.Pid = proc.Pid()
		resp.Uptime = proc.Uptime()
		resp.Service = proc.svc
	}
	return nil
}

type RestartRequest struct{}
type RestartResponse struct {
	Pid    int           `json:"pid"`
	Uptime time.Duration `json:"uptime"`
}

func (s Supervisor) Restart(req RestartRequest, resp *RestartResponse) error {
	err := s.callRestart()
	if err != nil {
		return err
	}
	if proc := s.callStatus(); proc != nil {
		resp.Pid = proc.Pid()
		resp.Uptime = proc.Uptime()
	}
	return nil
}

type UpdateRequest struct{ Service *cluster.Service }
type UpdateResponse struct {
	Pid    int           `json:"pid"`
	Uptime time.Duration `json:"duration"`
}

func (s Supervisor) Update(req UpdateRequest, resp *UpdateResponse) error {
	err := s.callUpdate(req.Service)
	if err != nil {
		return err
	}
	if proc := s.callStatus(); proc != nil {
		resp.Pid = proc.Pid()
		resp.Uptime = proc.Uptime()
	}
	return nil
}

func Launch(supervisor *Supervisor, port int) error {
	go supervisor.monitor()
	server := rpc.NewServer()
	err := server.Register(supervisor)
	if err != nil {
		return err
	}
	server.HandleHTTP("/", "/debug")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	log.Log.Info(fmt.Sprintf("gaffer is listening @0.0.0.0:%d", port))
	err = http.Serve(listener, nil)
	supervisor.shutdown <- true
	return err
}
