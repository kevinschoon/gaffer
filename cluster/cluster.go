package cluster

import (
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"math"
	"time"
)

// State represents the state of a given cluster
type State int

func (s State) String() string {
	switch s {
	case CONVERGING:
		return "CONVERGING"
	case STARTING:
		return "STARTING"
	case STARTED:
		return "STARTED"
	case DEGRADED:
		return "DEGRADED"
	}
	return ""
}

const (
	_ State = iota
	// Management hosts are still coming online
	CONVERGING
	// One or more host services are still starting
	STARTING
	// All essential host services have started
	STARTED
	// One or more services have not reported back
	DEGRADED
)

// Cluster represents the overall configuration of a Mesos cluster
type Cluster struct {
	ID       string                        `json:"id"`
	Hosts    []*host.Host                  `json:"hosts"`
	Services map[string][]*service.Service `json:"services"`
}

func New(id string, size int) *Cluster {
	cluster := &Cluster{
		ID:       id,
		Hosts:    []*host.Host{},
		Services: map[string][]*service.Service{},
	}
	for i := 0; i < size; i++ {
		cluster.Hosts = append(cluster.Hosts, host.NewHost())
		cluster.Services[cluster.Hosts[i].ID] = []*service.Service{}
	}
	return cluster
}

func (c Cluster) State() State {
	state := CONVERGING
	if c.Hosts == nil {
		// No hosts registered
		// CONVERGING
		return state
	}
	for _, host := range c.Hosts {
		if !host.Registered {
			// Not all hosts registered
			// CONVERGING
			return state
		}
	}
	// All hosts registered
	// STARTING
	state++
	for _, services := range c.Services {
		for _, service := range services {
			// Process is not registered
			if service.Process == nil {
				return state
			}
		}
	}
	// All services are running
	// STARTED
	state++
	for _, host := range c.Hosts {
		// Host has not been contacted recently
		if host.TimeSinceLastContacted() > 1*time.Minute {
			// DEGRADED
			state++
			break
		}
	}

	for _, services := range c.Services {
		for _, service := range services {
			if service.TimeSinceLastContacted() > 1*time.Minute || service.LastContacted.IsZero() {
				state++
				break
			}
		}
	}
	if state > DEGRADED {
		return DEGRADED
	}
	return state
}

func (c Cluster) Service(host, id string) *service.Service {
	if services, ok := c.Services[host]; ok {
		for _, svc := range services {
			if svc.ID == id {
				return svc
			}
		}
	}
	return nil
}

// Quorum returns the optimal quorum size
// for the cluster
func (c Cluster) Quorum() int {
	return int(math.Floor(float64(len(c.Hosts)) + .5))
}
