package config

import (
	"encoding/json"
)

// Option represents a process configuration
type Option struct {
	Name  string          `json:"name"`
	Value string          `json:"value"`
	Data  json.RawMessage `json:"data"`
}

func findOpt(name string, opts []*Option) *Option {
	for _, opt := range opts {
		if opt.Name == name {
			return opt
		}
	}
	return nil
}

func merge(opts []*Option, other []*Option) {
	if other == nil {
		return
	}
	for _, opt := range other {
		if o := findOpt(opt.Name, opts); o != nil {
			o.Name = opt.Name
			o.Value = opt.Value
			o.Data = opt.Data
		} else {
			opts = append(opts, opt)
		}
	}
}
