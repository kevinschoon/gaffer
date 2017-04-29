package config

import (
	"math"
)

// State represents the state of a given cluster
type State int

func (s State) String() string {
	switch s {
	case ZK_CONVERGING:
		return "ZK_CONVERGING"
	case ZK_CONVERGED:
		return "ZK_CONVERGED"
	case ZK_READY:
		return "ZK_READY"
	case MASTER_CONVERGING:
		return "MASTER_CONVERGING"
	case MASTER_CONVERGED:
		return "MASTER_CONVERGED"
	case MASTER_READY:
		return "MASTER_READY"
	}
	return ""
}

const (
	// Zookeeper supervisors are still coming online and we are
	// resolving their IP addresses/hostnames
	ZK_CONVERGING State = iota
	// All Zookeeper supervisors are online and we are ready to begin
	// launching Zookeeper processes
	ZK_CONVERGED
	// All Zookeepers have converged and their processes are running
	ZK_READY
	// Master supervisors are still coming online and we are
	// resolving their IP addresses
	MASTER_CONVERGING
	// All master supervisors are online and we are ready to
	// begin launching the Mesos master process
	MASTER_CONVERGED
	// All Mesos master processes are running
	MASTER_READY
)

// Cluster represents the overall configuration of a Mesos cluster
type Cluster struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Size          int          `json:"size"`
	ZKOptions     []*Option    `json:"zk_options"`
	MasterOptions []*Option    `json:"master_options"`
	AgentOptions  []*Option    `json:"agent_options"`
	Masters       []*Master    `json:"masters"`
	Zookeepers    []*Zookeeper `json:"zookeepers"`
}

func (c Cluster) State() State {
	// ZK_CONVERGING
	state := State(0)
	// No Zookeepers recorded
	if c.Zookeepers == nil {
		// ZK_CONVERGING
		return state
	}
	// Not all Zookeepers have registered
	if len(c.Zookeepers) != c.Size {
		// ZK_CONVERGING
		return state
	}
	// Range each ZK and check if we have registered a hostname
	for i := 0; i < c.Size; i++ {
		// Not all Zookeepers have registered
		if c.Zookeepers[i].Hostname == "" {
			// ZK_CONVERGING
			return state
		}
	}
	// ZK_CONVERGED
	state++
	// Range each Zookeeper and check if it is running
	for i := 0; i < c.Size; i++ {
		// Not all Zookeepers are running
		if !c.Zookeepers[i].Running {
			// ZK_CONVERGED
			return state
		}
	}
	// ZK_READY
	state++
	// No masters recorded
	if c.Masters == nil {
		// ZK_READY
		return state
	}
	// MASTER_CONVERGING
	state++
	// Not all masters have registered
	if len(c.Masters) != c.Size {
		// MASTER_CONVERGING
		return state
	}
	// Range each master and check if we have registered a hostname
	for i := 0; i < c.Size; i++ {
		// Not all masters have registered
		if c.Masters[i].Hostname == "" {
			// MASTER_CONVERGING
			return state
		}
	}
	// MASTER_CONVERGED
	state++
	// Range each master and check if it is running
	for i := 0; i < c.Size; i++ {
		// Not all masters are running
		if !c.Masters[i].Running {
			// MASTER_CONVERGED
			return state
		}
	}
	// MASTER_READY
	state++
	return state
}

func (c Cluster) Quorum() int {
	return int(math.Floor(float64(c.Size) + .5))
}
