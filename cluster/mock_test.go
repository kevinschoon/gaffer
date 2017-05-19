package cluster

import (
	"github.com/stretchr/testify/assert"
	"github.com/vektorlab/gaffer/cluster/host"
	"testing"
	"time"
)

var (
	started = Cluster{
		ID: "test-cluster",
		Hosts: []*host.Host{
			&host.Host{
				ID:            "host-1",
				Registered:    true,
				LastContacted: time.Now(),
			},
			&host.Host{
				ID:            "host-2",
				Registered:    true,
				LastContacted: time.Now(),
			},
			&host.Host{
				ID:            "host-3",
				Registered:    true,
				LastContacted: time.Now(),
			},
		},
	}
)

func TestMockOperator(t *testing.T) {
	m := &Mock{10}
	services := m.Update(started)
	assert.Len(t, services, 3)
}
