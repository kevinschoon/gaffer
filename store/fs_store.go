package store

import (
	"encoding/json"
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/service"
	"github.com/opencontainers/runtime-spec/specs-go"
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
	BasePath   string
	ConfigPath string
	mu         sync.RWMutex
}

func (s *FSStore) services(path string) ([]service.Service, error) {
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	svcs := []service.Service{}
	for _, dir := range dirs {
		bundle := filepath.Join(path, dir.Name())
		cfgPath := filepath.Join(bundle, "config.json")
		log.Log.Debug(fmt.Sprintf("loading service from dir %s", bundle))
		raw, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			return nil, err
		}
		svc := service.Service{Id: dir.Name(), Bundle: bundle}
		spec := &specs.Spec{}
		err = json.Unmarshal(raw, spec)
		if err != nil {
			return nil, err
		}
		var modified bool
		envs, err := loadEnvs(filepath.Join(s.ConfigPath, svc.Id, "envs.json"))
		if err == nil {
			modified = true
			for key, value := range envs {
				log.Log.Debug(fmt.Sprintf("updating environment variable %s=%s", key, value))
				spec.Process.Env = append(spec.Process.Env, fmt.Sprintf("%s=%s", key, value))
			}
			log.Log.Debug(fmt.Sprintf("environment for service %s updated from local config", svc.Id))
		} else {
			log.Log.Debug(fmt.Sprintf("could not load environment from local config: %s", err.Error()))
		}
		if modified {
			// Write out updated configuration
			log.Log.Debug(fmt.Sprintf("re-writing updated bundle config %s", cfgPath))
			err = ioutil.WriteFile(cfgPath, raw, 0644)
			if err != nil {
				return nil, err
			}
		}
		log.Log.Debug("loaded service bundle", zap.Any("spec", spec))
		svcs = append(svcs, service.WithSpec(*spec)(svc))
	}
	return svcs, nil
}

func (s *FSStore) Services() ([]service.Service, error) {
	return s.services(s.BasePath)
}

func NewFSStore(cfg config.Config) *FSStore {
	return &FSStore{
		BasePath:   cfg.Store.BasePath,
		ConfigPath: cfg.Store.ConfigPath,
	}
}

func loadEnvs(path string) (map[string]string, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	envs := map[string]string{}
	return envs, json.Unmarshal(raw, &envs)
}
