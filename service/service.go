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
	json.Unmarshal(svc.Spec, spec)
	return spec
}

func Stats(svc Service) *runc.Stats {
	stats := &runc.Stats{}
	json.Unmarshal(svc.Stats, stats)
	return stats
}
