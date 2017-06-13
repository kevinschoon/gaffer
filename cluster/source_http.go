package cluster

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

type HTTPSource struct {
	endpoint *url.URL
}

func (hs HTTPSource) Get() (*Config, error) {
	resp, err := http.Get(hs.endpoint.String() + "/get")
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = json.NewDecoder(resp.Body).Decode(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (hs HTTPSource) Set(config *Config) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = http.Post(hs.endpoint.String()+"/set", "application/json", bytes.NewBuffer(raw))
	if err != nil {
		return err
	}
	return nil
}

func NewHTTPSource(pattern string) (*HTTPSource, error) {
	u, err := url.Parse(pattern)
	if err != nil {
		return nil, err
	}
	return &HTTPSource{endpoint: u}, nil
}
