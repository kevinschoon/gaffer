package supervisor

import (
	"bufio"
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const TempPath = "/tmp"

type Process struct {
	Cmd     *exec.Cmd `json:"-"`
	svc     *cluster.Service
	started time.Time
}

func NewProcess(svc *cluster.Service) (*Process, error) {
	proc := &Process{Cmd: exec.Command(svc.Args[0], svc.Args[1:]...), svc: svc}

	tmp, err := ioutil.TempDir(TempPath, "gaffer")
	if err != nil {
		return nil, err
	}

	err = os.Chdir(tmp)
	if err != nil {
		return nil, err
	}

	for _, env := range svc.Environment {
		proc.Cmd.Env = append(proc.Cmd.Env, env.String())
	}

	for _, file := range svc.Files {
		err = file.Write(tmp)
		if err != nil {
			return nil, err
		}
	}

	return proc, nil
}

// Start runs the command
func (p *Process) Start() error {

	// Stdout
	ro, wo, err := os.Pipe()
	if err != nil {
		return err
	}
	p.Cmd.Stdout = wo
	go func() {
		scanner := bufio.NewScanner(ro)
		for scanner.Scan() {
			log.Log.Info(
				"stdout",
				zap.Int("pid", p.Pid()),
				zap.String("content", scanner.Text()),
			)
		}
	}()

	// Stderr
	re, we, err := os.Pipe()
	if err != nil {
		return err
	}
	p.Cmd.Stderr = we
	go func() {
		scanner := bufio.NewScanner(re)
		for scanner.Scan() {
			log.Log.Info(
				"stderr",
				zap.Int("pid", p.Pid()),
				zap.String("content", scanner.Text()),
			)
		}
	}()

	// Start the process
	if err := p.Cmd.Start(); err != nil {
		return err
	}
	p.started = time.Now()

	log.Log.Info(
		"creating new service process",
		zap.String("process", p.Cmd.Path),
		zap.Strings("args", p.Cmd.Args),
	)

	go func() {
		err := p.Cmd.Wait()
		if err != nil {
			log.Log.Info(fmt.Sprintf("process ended: %s", err.Error()))
		}
	}()

	return nil
}

func (p *Process) Stop() error {
	//processGroup := 0 - p.Cmd.Process.Pid
	if p.Running() {
		return syscall.Kill(p.Pid(), syscall.SIGKILL)
	}
	return nil
}

func (p *Process) Restart() error {
	err := p.Stop()
	if err != nil {
		return err
	}
	proc, err := NewProcess(p.svc)
	if err != nil {
		return err
	}
	*p = *proc
	return p.Start()
}

func (p *Process) Running() bool {
	pid := p.Pid()
	if pid == 0 {
		return false
	}
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}

func (p *Process) Pid() int {
	if p.Cmd == nil || p.Cmd.Process == nil {
		return 0
	}
	return p.Cmd.Process.Pid
}

func (p *Process) Uptime() time.Duration { return time.Since(p.started) }
