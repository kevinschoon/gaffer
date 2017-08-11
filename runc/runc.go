package runc

import (
	"context"
	"fmt"
	"github.com/containerd/go-runc"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"go.uber.org/zap"
	"syscall"
	"time"
)

type Runc struct {
	rc       *runc.Runc
	readOnly bool
	mount    bool
	bundle   string
	id       string
	io       *IO
	started  time.Time
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

func (rc *Runc) Mount() error {
	if rc.readOnly {
		log.Log.Info(fmt.Sprintf("mounting container bundle RO: %s", rc.bundle))
		return prepareRO(rc.bundle)
	}
	log.Log.Info(fmt.Sprintf("mounting container bundle RW: %s", rc.bundle))
	return prepareRW(rc.bundle)
}

func (rc *Runc) Unmount() error {
	log.Log.Info(fmt.Sprintf("unmounting container bundle: %s", rc.bundle))
	if rc.readOnly {
		return cleanupRO(rc.bundle)
	}
	return cleanupRW(rc.bundle)
}

func (rc *Runc) Run() (int, error) {
	if rc.mount {
		err := rc.Mount()
		if err != nil {
			return -1, err
		}
	}
	io, err := NewIO(rc.id)
	if err != nil {
		return -1, err
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
	if rc.mount {
		err := rc.Unmount()
		if err != nil {
			return err
		}
	}
	return rc.rc.Kill(
		context.Background(),
		rc.id,
		int(syscall.SIGKILL),
		// TODO: On my system running
		// "rootless" the --all flag
		// has the effect of bleeding
		// into all my other user processes
		// killing the entire desktop!
		// Unsure exactly what the cause is.
		&runc.KillOpts{All: false},
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

func New(id, bundle string, readOnly bool, cfg config.Config) *Runc {
	rc := &Runc{
		id:       id,
		bundle:   bundle,
		readOnly: readOnly,
		mount:    cfg.Runc.Mount,
		rc:       &runc.Runc{Root: cfg.Runc.Root},
	}
	return rc
}
