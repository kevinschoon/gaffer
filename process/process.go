package process

import (
	"bufio"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

/*
type Process interface {
	Start() error
	Kill() error
	Pid() int
	Signal(syscall.Signal) error
}
*/

type Process struct {
	cmd  *exec.Cmd
	err  chan error
	quit chan struct{}
	Env  map[string]string
	log  *zap.Logger
}

func NewProcess(logger *zap.Logger, args ...string) *Process {
	return &Process{
		log:  logger,
		cmd:  exec.Command(args[0], args[1:]...),
		err:  make(chan error),
		quit: make(chan struct{}),
		Env:  map[string]string{},
	}
}

// Start runs the command
func (p *Process) Start() error {

	// Append any local envs
	p.cmd.Env = os.Environ()
	for k, v := range p.Env {
		p.cmd.Env = append(p.cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Stdout
	ro, wo, err := os.Pipe()
	if err != nil {
		return err
	}
	p.cmd.Stdout = wo
	go func() {
		scanner := bufio.NewScanner(ro)
		for scanner.Scan() {
			p.log.Info(
				p.Name(),
				zap.Int("PID", p.Pid()),
				zap.String("STDOUT", scanner.Text()),
			)
		}
	}()

	// Stderr
	re, we, err := os.Pipe()
	if err != nil {
		return err
	}
	p.cmd.Stderr = we
	go func() {
		scanner := bufio.NewScanner(re)
		for scanner.Scan() {
			p.log.Info(
				p.Name(),
				zap.Int("PID", p.Pid()),
				zap.String("STDERR", scanner.Text()),
			)
		}
	}()

	// Start the process
	if err := p.cmd.Start(); err != nil {
		return err
	}

	p.log.Info(
		"process",
		zap.String("process", p.cmd.Path),
		zap.Strings("args", p.cmd.Args),
	)

	go func() {
		err := p.cmd.Wait()
		if err != nil {
			p.err <- err
		}
		close(p.quit)
	}()

	return nil
}

func (p *Process) Name() string {
	if p.cmd != nil {
		split := strings.Split(p.cmd.Path, "/")
		return split[len(split)-1]
	}
	return ""
}

// Kill the entire Process group.
func (p *Process) Kill() error {
	processGroup := 0 - p.cmd.Process.Pid
	return syscall.Kill(processGroup, syscall.SIGKILL)
}

// Pid return Process PID
func (p *Process) Pid() int {
	if p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

// Signal sends a signal to the Process
func (p *Process) Signal(sig syscall.Signal) error {
	return syscall.Kill(p.cmd.Process.Pid, sig)
}
