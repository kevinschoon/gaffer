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

type State struct {
	Date    time.Time   `json:"date"`
	Process *os.Process `json:"process"`
	Message string      `json:"message"`
}

// Service is a configurable process
// that must remain running
type Service struct {
	ID          string      `json:"id"`
	Args        []string    `json:"args"`
	Cmd         *exec.Cmd   `json:"-"`
	Process     *os.Process `json:"process"`
	History     []*State    `json:"history"`
	Environment []*Env      `json:"environment"`
	Files       []*File     `json:"files"`
}

//TODO
//func (s *Service) Flapping() bool {}

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
	if s.Running() {
		s.Process = s.Cmd.Process
	} else {
		if s.History == nil {
			s.History = []*State{}
		}
		s.History = append(s.History, &State{
			Date:    time.Now(),
			Process: s.Process,
			Message: s.Cmd.ProcessState.String(),
		})
		s.Process = nil
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
		"process",
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
