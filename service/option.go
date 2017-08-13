package service

import (
	"encoding/json"
	"github.com/containerd/go-runc"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type Option func(Service) Service

func WithStats(stats runc.Stats) Option {
	return func(svc Service) Service {
		raw, _ := json.Marshal(stats)
		return Service{
			Id:     svc.Id,
			Bundle: svc.Bundle,
			Spec:   svc.Spec,
			Stats:  raw,
		}
	}
}

func WithSpec(spec specs.Spec) Option {
	return func(svc Service) Service {
		raw, _ := json.Marshal(spec)
		return Service{
			Id:     svc.Id,
			Bundle: svc.Bundle,
			Stats:  svc.Stats,
			Spec:   raw,
		}
	}
}
