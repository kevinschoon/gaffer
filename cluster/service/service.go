package service

import (
	"bufio"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"syscall"
)

// Service is a configurable process
// that must remain running
type Service struct {
	Cmd     *exec.Cmd `json:"cmd"`
	Running bool      `json:"running"`
}

// Start runs the command
func (s *Service) Start() error {

	// Stdout
	ro, wo, err := os.Pipe()
	if err != nil {
		return err
	}
	s.Cmd.Stdout = wo
	go func() {
		scanner := bufio.NewScanner(ro)
		for scanner.Scan() {
			log.Log.Info(
				"stdout",
				zap.Int("pid", s.Pid()),
				zap.String("content", scanner.Text()),
			)
		}
	}()

	// Stderr
	re, we, err := os.Pipe()
	if err != nil {
		return err
	}
	s.Cmd.Stderr = we
	go func() {
		scanner := bufio.NewScanner(re)
		for scanner.Scan() {
			log.Log.Info(
				"stderr",
				zap.Int("pid", s.Pid()),
				zap.String("content", scanner.Text()),
			)
		}
	}()

	// Start the process
	if err := s.Cmd.Start(); err != nil {
		return err
	}

	log.Log.Info(
		"process",
		zap.String("process", s.Cmd.Path),
		zap.Strings("args", s.Cmd.Args),
	)

	go func() {
		err := s.Cmd.Wait()
		if err != nil {
			log.Log.Error("process", zap.Error(err))
		}
		//close(s.quit)
	}()

	return nil
}

func (s *Service) Stop() error {
	processGroup := 0 - s.Cmd.Process.Pid
	return syscall.Kill(processGroup, syscall.SIGKILL)
}

/*
func (s *Service) Running() bool {
	pid := s.Pid()
	return pid != 0 && syscall.Kill(pid, syscall.Signal(0)) == nil
}
*/

// Pid return Process PID
func (s *Service) Pid() int {
	if s.Cmd == nil || s.Cmd.Process == nil {
		return 0
	}
	return s.Cmd.Process.Pid
}

// Signal sends a signal to the Process
func (s *Service) Signal(sig syscall.Signal) error {
	return syscall.Kill(s.Cmd.Process.Pid, sig)
}
