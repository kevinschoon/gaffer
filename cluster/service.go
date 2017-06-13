package cluster

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
)

type Services []*Service

func (services Services) Find(name string) *Service {
	for _, svc := range services {
		if svc.ID == name {
			return svc
		}
	}
	return nil
}

// Service is a configurable process
// that must remain running
type Service struct {
	ID          string   `json:"id"`
	Args        []string `json:"args"`
	Environment []*Env   `json:"environment"`
	Files       []*File  `json:"files"`
}

func (s Service) Env(name string) *Env {
	for _, env := range s.Environment {
		if env.Name == name {
			return env
		}
	}
	return nil
}

func (s *Service) Equal(o *Service) bool {
	s1, _ := json.Marshal(s)
	s2, _ := json.Marshal(o)
	return bytes.Compare(s1, s2) == 0
}

func (s *Service) Hash() string {
	r, _ := json.Marshal(s)
	return fmt.Sprintf("%x", md5.Sum(r))
}
