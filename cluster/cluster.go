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
	case CONVERGED:
		return "CONVERGED"
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
	// Management hosts are still coming online
	CONVERGING State = iota
	// All management hosts have joined
	CONVERGED
	// One or more host proceses are still starting
	STARTING
	// All essential host processes have started
	STARTED
	// One or more essential host process is not running
	DEGRADED
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

func (c Cluster) State() State { return STARTED }

// Quorum returns the optimal quorum size
// for the cluster
func (c Cluster) Quorum() int {
	return int(math.Floor(float64(c.Size) + .5))
}
