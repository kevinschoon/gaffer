package store

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/service"
	"go.uber.org/zap"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// FSStore is a read-only "database"
// for storing service configuration.
// The base path is the LinuxKit
// default of /containers/services/<svc>
type FSStore struct {
	BasePath string
	mu       sync.RWMutex
}

func (s FSStore) Service(id string) (*service.Service, error) {
	services, err := s.Services()
	if err != nil {
		return nil, err
	}
	for _, service := range services {
		if service.Id == id {
			return service, nil
		}
	}
	return nil, fmt.Errorf("%s not found", id)
}

func (s FSStore) Services() ([]*service.Service, error) {
	dirs, err := ioutil.ReadDir(s.BasePath)
	if err != nil {
		return nil, err
	}
	svcs := []*service.Service{}
	for _, dir := range dirs {
		bundle := filepath.Join(s.BasePath, dir.Name())
		log.Log.Debug(fmt.Sprintf("loading service from dir %s", bundle))
		raw, err := ioutil.ReadFile(filepath.Join(bundle, "config.json"))
		if err != nil {
			return nil, err
		}
		svc := &service.Service{Id: dir.Name(), Bundle: bundle}
		spec := &specs.Spec{}
		err = json.Unmarshal(raw, spec)
		if err != nil {
			return nil, err
		}
		log.Log.Debug("loaded service bundle", zap.Any("spec", spec))
		svc.Spec = &any.Any{Value: raw}
		svcs = append(svcs, svc)
	}
	return svcs, nil
}

func NewFSStore(path string) *FSStore {
	return &FSStore{BasePath: path}
}
