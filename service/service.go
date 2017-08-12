package service

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func (s *Service) Equal(o *Service) bool {
	s1, _ := json.Marshal(s)
	s2, _ := json.Marshal(o)
	return bytes.Compare(s1, s2) == 0
}

func (s *Service) Hash() string {
	r, _ := json.Marshal(s)
	return fmt.Sprintf("%x", md5.Sum(r))
}

func (s *Service) ReadOnly() (bool, error) {
	spec, err := s.UnmarshalSpec()
	if err != nil {
		return false, err
	}
	return spec.Root.Readonly, nil
}

func (s *Service) UnmarshalSpec() (*specs.Spec, error) {
	spec := &specs.Spec{}
	err := json.Unmarshal(s.Spec, spec)
	if err != nil {
		return nil, err
	}
	return spec, nil
}
