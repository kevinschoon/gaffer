package config

import (
	"encoding/json"
	"fmt"
	"github.com/mesanine/gaffer/log"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Store      Store
	Runc       Runc
	Etcd       Etcd
	RPCServer  RPCServer
	HTTPServer HTTPServer
	User       User
}

type Store struct {
	BasePath   string
	ConfigPath string
	// Toggle if we should handle overlay
	// mounts ourself.
	Mount bool
	// Move lower --> rootfs
	MoveRoot bool
}

type Runc struct {
	Root string
}

type Etcd struct {
	Endpoints []string
}

type RPCServer struct {
	Port int
}

type HTTPServer struct {
	Port int
}

type User struct {
	User string
}

func (s Store) Envs() (map[string]map[string]string, error) {
	envs := map[string]map[string]string{}
	dirs, err := ioutil.ReadDir(s.ConfigPath)
	if err != nil {
		log.Log.Warn(fmt.Sprintf("could not load config from %s: %s", s.ConfigPath, err.Error()))
		return envs, nil
	}
	for _, dir := range dirs {
		raw, err := ioutil.ReadFile(filepath.Join(s.ConfigPath, dir.Name(), "envs.json"))
		if err != nil {
			continue
		}
		svcEnvs := map[string]string{}
		err = json.Unmarshal(raw, &svcEnvs)
		if err != nil {
			return nil, err
		}
		envs[dir.Name()] = svcEnvs
	}
	return envs, nil
}
