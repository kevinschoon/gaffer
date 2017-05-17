package supervisor

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

const PollTime = 200 * time.Millisecond

// Supervisor is an RPC service that wraps
// a service process. It allows checking
// the wrapped process status and pushing
// a new service configuration. If the new
// configuration fails to start after a few
// seconds it will return an error and revert
// to the previous configuration.
type Supervisor struct {
	svc      *service.Service
	proc     *Process
	started  time.Time
	shutdown chan bool
	restart  chan struct {
		err chan error
	}
	update chan struct {
		service *service.Service
		err     chan error
	}
}

func NewSupervisor(seed *service.Service) (*Supervisor, error) {
	proc, err := NewProcess(seed)
	if err != nil {
		return nil, err
	}
	return &Supervisor{
		svc:      seed,
		proc:     proc,
		shutdown: make(chan bool),
		restart: make(chan struct {
			err chan error
		}),
		update: make(chan struct {
			service *service.Service
			err     chan error
		}),
	}, nil
}

func (s *Supervisor) callUpdate(svc *service.Service) error {
	err := make(chan error)
	s.update <- struct {
		service *service.Service
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
			}
			e.err <- s.proc.Restart()
		case u := <-s.update:
			u.err <- s.replace(u.service)
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
			s.svc.ID,
			zap.String("message", "process is not running"),
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
					s.svc.ID,
					zap.String("message", "process failed to start"),
					zap.Duration("duration", d),
					zap.Any("service", s.svc),
					zap.Any("process", s.proc),
					zap.Error(err),
				)
			}))
		if err != nil {
			log.Log.Warn(
				s.svc.ID,
				zap.String("message", "process has timed out"),
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
func (s *Supervisor) replace(svc *service.Service) error {
	// The existing process configuration
	previous := s.proc
	revert := func() {
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
	// Create a new process and tmp path
	proc, err := NewProcess(svc)
	if err != nil {
		return err
	}
	// A Process was already configured
	if s.proc != nil {
		// Process is actively running
		if s.proc.Running() {
			// Attempt to stop the active process
			err = s.proc.Stop()
			if err != nil {
				return err
			}
		}
		// No process currently configured
	} else {
		s.proc = proc
		return s.proc.Start()
	}
	// Attempt to start the process
	err = proc.Start()
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
	// Finally replace the old process and service configuration
	s.proc = proc
	s.svc = svc
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
	Uptime time.Duration `json:"uptime"`
	Pid    int           `json:"pid"`
}

func (s Supervisor) Status(req StatusRequest, resp *StatusResponse) error {
	if s.proc == nil {
		return fmt.Errorf("process is not running")
	}
	resp.Pid = s.proc.Pid()
	resp.Uptime = time.Since(s.started)
	return nil
}

type RestartRequest struct{}
type RestartResponse struct{ Restarted bool }

func (s Supervisor) Restart(req RestartRequest, resp *RestartResponse) error {
	err := s.callRestart()
	if err != nil {
		return err
	}
	resp.Restarted = true
	return nil
}

type UpdateRequest struct{ Service *service.Service }
type UpdateResponse struct{ Process *os.Process }

func (s Supervisor) Update(req UpdateRequest, resp *UpdateResponse) error {
	err := s.callUpdate(req.Service)
	if err != nil {
		return err
	}
	resp.Process = s.proc.Cmd.Process
	return nil
}

type Command struct {
	update  *service.Service
	restart bool
}

func Launch(db store.Store, id string) error {
	var (
		self       *host.Host
		svc        *service.Service
		supervisor *Supervisor
	)
	register := func() error {
		h, s, err := store.Register(db, id)
		if err != nil {
			return err
		}
		self = h
		svc = s
		// Attempt to configure the initial service
		// based on the configuration in the database.
		spv, err := NewSupervisor(svc)
		if err != nil {
			return err
		}
		supervisor = spv
		// Successfully registered
		return nil
	}
	// Attempt to register continuously
	err := backoff.RetryNotify(
		register,
		backoff.NewConstantBackOff(5000*time.Millisecond),
		func(err error, d time.Duration) {
			log.Log.Info(
				id,
				zap.String("message", "service registration failed"),
				zap.Duration("duration", d),
				zap.Error(err),
			)
		},
	)
	if err != nil {
		return err
	}
	log.Log.Info(
		id,
		zap.String("message", "registration complete"),
		zap.Any("host", self),
		zap.Any("service", svc),
	)
	go supervisor.monitor()
	server := rpc.NewServer()
	err = server.Register(supervisor)
	if err != nil {
		return err
	}
	server.HandleHTTP("/", "/debug")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", svc.Port))
	if err != nil {
		return err
	}
	log.Log.Info(
		svc.ID,
		zap.String("message", fmt.Sprintf("supervisor listening @0.0.0.0:%d", svc.Port)),
	)
	err = http.Serve(listener, nil)
	supervisor.shutdown <- true
	return err
}
