package cluster

import (
	"math"
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
)

// Cluster represents the overall configuration of a Mesos cluster
type Cluster struct {
	ID    string  `json:"id"`
	Size  int     `json:"size"`
	Hosts []*Host `json:"hosts"`
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
	for _, host := range c.Hosts {
		for _, service := range host.Services {
			if service.Process == nil {
				// All services are not running yet
				return state
			}
		}
	}
	// All services are running
	// STARTED
	state++
	return state
}

// Quorum returns the optimal quorum size
// for the cluster
func (c Cluster) Quorum() int {
	return int(math.Floor(float64(c.Size) + .5))
}
