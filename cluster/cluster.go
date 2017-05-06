package cluster

import (
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
	ID       string                      `json:"id"`
	Size     int                         `json:"size"`
	Hosts    []*Host                     `json:"hosts"`
	Services map[string]*service.Service `json:"services"`
}

func New(id string, size int) *Cluster {
	cluster := &Cluster{
		ID:    id,
		Size:  size,
		Hosts: []*Host{},
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
	for _, service := range c.Services {
		// Process is not registered
		if service.Process == nil {
			return state
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
	return state
}

// Quorum returns the optimal quorum size
// for the cluster
func (c Cluster) Quorum() int {
	return int(math.Floor(float64(len(c.Hosts)) + .5))
}
