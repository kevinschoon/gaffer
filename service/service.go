package service

import (
	"encoding/json"
	"github.com/containerd/go-runc"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func ReadOnly(svc Service) bool {
	return Spec(svc).Root.Readonly
}

func Spec(svc Service) *specs.Spec {
	spec := &specs.Spec{}
	err := json.Unmarshal(svc.Spec, spec)
	if err != nil {
		panic(err)
	}
	return spec
}

func Stats(svc Service) *runc.Stats {
	stats := &runc.Stats{}
	err := json.Unmarshal(svc.Stats, stats)
	if err != nil {
		panic(err)
	}
	return stats
}
