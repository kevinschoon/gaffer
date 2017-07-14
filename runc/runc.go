package runc

import (
	"context"
	"github.com/containerd/go-runc"
	"github.com/vektorlab/gaffer/config"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"syscall"
	"time"
)

type Runc struct {
	rc      *runc.Runc
	bundle  string
	id      string
	io      *IO
	started time.Time
}

func (rc *Runc) Container() (*runc.Container, error) {
	containers, err := rc.rc.List(context.Background())
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		if container.ID == rc.id {
			return container, nil
		}
	}
	return nil, nil
}

func (rc *Runc) Delete() error {
	return rc.rc.Delete(context.Background(), rc.id, &runc.DeleteOpts{Force: true})
}

func (rc *Runc) Run() (int, error) {
	io, err := NewIO(rc.id)
	if err != nil {
		return 0, err
	}
	rc.io = io
	defer func() {
		io.Close()
		//rc.io = nil
	}()
	rc.io.Start()
	rc.started = time.Now()
	return rc.rc.Run(context.Background(), rc.id, rc.bundle, &runc.CreateOpts{IO: io.rio})
}

func (rc *Runc) Stop() error {
	rc.io.Close()
	return rc.rc.Kill(
		context.Background(),
		rc.id,
		int(syscall.SIGKILL),
		&runc.KillOpts{All: true},
	)
}

func (rc *Runc) Running() bool {
	container, err := rc.rc.State(context.Background(), rc.id)
	if err != nil {
		log.Log.Error("couldn't get container state", zap.Error(err))
		return false
	}
	return container.Status == "running"
}

func (rc *Runc) Stats() (*runc.Stats, error) {
	return rc.rc.Stats(context.Background(), rc.id)
}

func (rc *Runc) Uptime() time.Duration {
	return time.Since(rc.started)
}

func New(id, bundle string, cfg config.Config) *Runc {
	rc := &Runc{
		id:     id,
		bundle: bundle,
		rc:     &runc.Runc{},
	}
	return rc
}
