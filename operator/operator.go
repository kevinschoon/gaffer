package operator

import (
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/operator/mesos"
	"github.com/vektorlab/gaffer/operator/mock"
)

// An Operator emits a service configuration based on
// pre-defined options of a given cluster.
// Gaffer only supports configuring Mesos clusters for now
// but could be updated to support other systems in the future.
type Operator interface {
	// Update returns the desired service configuration
	// based on the cluster input. If no change is required
	// Update returns nil.
	Update(*cluster.Cluster) map[string][]*service.Service
}

var (
	_ Operator = mesos.Mesos{}
	_ Operator = mock.Mock{}
)
