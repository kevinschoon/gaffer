package store

import (
	"encoding/json"
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/ginit/mount"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/service"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FSStore reads runc Service
// configuration from a base
// path. It is is compatible
// with LinuxKit's /containers/{service,onboot}
// paths.
type FSStore struct {
	BasePath string
	Mount    bool
	MoveRoot bool
	config   config.Config
}

func (s FSStore) Services() ([]service.Service, error) {
	dirs, err := ioutil.ReadDir(s.BasePath)
	if err != nil {
		return nil, err
	}
	svcs := []service.Service{}
	for _, dir := range dirs {
		bundle := filepath.Join(s.BasePath, dir.Name())
		log.Log.Debug(fmt.Sprintf("loading service from dir %s", bundle))
		// Load the runc spec
		raw, err := ioutil.ReadFile(filepath.Join(bundle, "config.json"))
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, service.Service{Id: dir.Name(), Bundle: bundle, Spec: raw})
	}
	return svcs, nil
}

// Clean up rootfs mounts if they
// are being handled by us.
func (s FSStore) Close() error {
	if s.Mount {
		services, err := s.Services()
		if err != nil {
			return err
		}
		for _, svc := range services {
			log.Log.Info(fmt.Sprintf("unmounting rootfs @ %s", svc.Bundle))
			err := mount.Unmount(filepath.Join(svc.Bundle, "rootfs"))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Init makes several modifications to the
// container store path.
// BUG: This function does not clean up mount
// paths in all cases.
func (s FSStore) Init() error {
	services, err := s.Services()
	if err != nil {
		return err
	}
	for _, svc := range services {
		if s.MoveRoot {
			// Moby now creates a lower/upper directory
			// with the assumption that Linuxkit will
			// mount it in a particular way. Gaffer will
			// have to support both configurations for
			// as it needs to run without privileges
			// for testing and development.
			old := filepath.Join(svc.Bundle, "rootfs")
			// Path should be empty but rename to be safe
			err := os.Rename(old, fmt.Sprintf("%s__old", old))
			if err != nil {
				return err
			}
			// move /containers/<base>/<id>/lower --> /containers/<base>/<id>/rootfs
			err = os.Rename(filepath.Join(svc.Bundle, "lower"), old)
			if err != nil {
				return err
			}
		}
		// If we were given environment variables
		// to configure the service with we modify
		// it's config.json file.
		updates, err := s.config.Store.Envs()
		if err != nil {
			return err
		}
		if envs, ok := updates[svc.Id]; ok {
			updated := service.Spec(svc)
			// Append any existing environment variables
			// in the config.json file
			for key, value := range envs {
				updated.Process.Env = append(updated.Process.Env, fmt.Sprintf("%s=%s", key, value))
			}
			raw, err := json.Marshal(updated)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(filepath.Join(svc.Bundle, "config.json"), raw, 0644)
			if err != nil {
				return err
			}
		}
	}
	if s.Mount {
		// Range through the services again
		// mounting their rootfs path as RW
		// or RO depending on it's configuration.
		// (/containers/<base>/<service/rootfs)
		// BUG: if rootfs was moved and we are
		// still handling mounts this would fail
		// but it doesn't fit our use case for now.
		for _, svc := range services {
			if service.ReadOnly(svc) {
				// mount --bind -o ro ...
				path := filepath.Join(svc.Bundle, "lower")
				log.Log.Info(fmt.Sprintf("re-binding mount (RO) @ %s", path))
				err = mount.Mount(mount.Bind(path, true))
				if err != nil {
					return err
				}
			} else {
				// mount -t overlay ...
				lower := filepath.Join(svc.Bundle, "lower")
				target := filepath.Join(svc.Bundle, "rootfs")
				log.Log.Info(fmt.Sprintf("mounting overlayfs (RW) @ %s --> %s", lower, target))
				err = mount.Mount(mount.Overlay(lower, target))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func New(cfg config.Config) *FSStore {
	return &FSStore{
		config:   cfg,
		BasePath: cfg.Store.BasePath,
		Mount:    cfg.Store.Mount,
		MoveRoot: cfg.Store.MoveRoot,
	}
}
