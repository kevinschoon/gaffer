package client

import (
	"github.com/vektorlab/gaffer/cluster/service"
	"math/rand"
	"time"
)

const (
	MAX_PORT int = 65535
	MIN_PORT int = 49152
)

func assignPort(services []*service.Service) int {
	rand.Seed(time.Now().Unix())
	port := rand.Intn(MAX_PORT-MIN_PORT) + MIN_PORT
	for _, svc := range services {
		if svc.Port == port {
			return assignPort(services)
		}
	}
	return port
}
