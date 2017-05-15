package service

import (
	"bufio"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const TempPath = "/tmp"

// Service is a configurable process
// that must remain running
type Service struct {
	ID            string      `json:"id"`
	Args          []string    `json:"args"`
	Cmd           *exec.Cmd   `json:"-"`
	Process       *os.Process `json:"process"`
	Environment   []*Env      `json:"environment"`
	Files         []*File     `json:"files"`
	LastContacted time.Time   `json:"last_contacted"`
}

func (s Service) TimeSinceLastContacted() time.Duration {
	return time.Since(s.LastContacted)
}

func (s Service) Env(name string) *Env {
	for _, env := range s.Environment {
		if env.Name == name {
			return env
		}
	}
	return nil
}

func (s *Service) init() error {

	tmp, err := ioutil.TempDir(TempPath, "gaffer")
	if err != nil {
		return err
	}

	err = os.Chdir(tmp)
	if err != nil {
		return err
	}

	s.Cmd = exec.Command(s.Args[0], s.Args[1:]...)

	for _, env := range s.Environment {
		s.Cmd.Env = append(s.Cmd.Env, env.String())
	}

	for _, file := range s.Files {
		err = file.Write(tmp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Update() {
	s.LastContacted = time.Now()
	if s.Running() {
		s.Process = s.Cmd.Process
	}
}

// Start runs the command
func (s *Service) Start() error {

	err := s.init()

	if err != nil {
		return err
	}

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
		s.ID,
		zap.String("process", s.Cmd.Path),
		zap.Strings("args", s.Cmd.Args),
	)

	go func() {
		err := s.Cmd.Wait()
		if err != nil {
			log.Log.Error("process", zap.Error(err))
		}
		s.Update()
	}()

	return nil
}

func (s *Service) Stop() error {
	processGroup := 0 - s.Cmd.Process.Pid
	return syscall.Kill(processGroup, syscall.SIGKILL)
}

func (s *Service) Running() bool {
	pid := s.Pid()
	return pid != 0 && syscall.Kill(pid, syscall.Signal(0)) == nil
}

func (s *Service) Pid() int {
	if s.Cmd == nil || s.Cmd.Process == nil {
		return 0
	}
	return s.Cmd.Process.Pid
}
