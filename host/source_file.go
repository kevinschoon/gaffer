package host

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type FileSource struct {
	Path string
	mu   sync.RWMutex
}

func (fs FileSource) Get() (*Config, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	raw, err := ioutil.ReadFile(fs.Path)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(raw, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (fs FileSource) Set(config *Config) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fs.Path, raw, 0644)
}
