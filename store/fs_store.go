package store

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/vektorlab/gaffer/config"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/service"
	"go.uber.org/zap"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// FSStore reads runc Service
// configuration from a base
// path. It is is compatible
// with LinuxKit's /containers/{service,onboot}
// paths.
type FSStore struct {
	BasePath string
	mu       sync.RWMutex
}

func (s FSStore) services(path string) ([]*service.Service, error) {
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	svcs := []*service.Service{}
	for _, dir := range dirs {
		bundle := filepath.Join(path, dir.Name())
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

func (s FSStore) Services() ([]*service.Service, error) {
	return s.services(s.BasePath)
}

func NewFSStore(cfg config.Config) *FSStore {
	return &FSStore{BasePath: cfg.Store.BasePath}
}
