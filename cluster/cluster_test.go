package cluster

import (
	"github.com/stretchr/testify/assert"
	"github.com/vektorlab/gaffer/cluster/service"
	"os"
	"testing"
	"time"
)

var (
	converging = &Cluster{
		ID: "test-cluster",
		Hosts: []*Host{
			&Host{ID: "host-1"},
			&Host{ID: "host-2"},
			&Host{ID: "host-3"},
		},
	}
	starting = &Cluster{
		ID: "test-cluster",
		Hosts: []*Host{
			&Host{ID: "host-1", Registered: true},
			&Host{ID: "host-2", Registered: true},
			&Host{ID: "host-3", Registered: true},
		},
		Services: map[string]*service.Service{
			"host-1": &service.Service{Process: &os.Process{Pid: 1234}},
			"host-2": &service.Service{Process: &os.Process{Pid: 1234}},
			"host-3": &service.Service{},
		},
	}
	started = &Cluster{
		ID: "test-cluster",
		Hosts: []*Host{
			&Host{
				ID:            "host-1",
				Registered:    true,
				LastContacted: time.Now(),
			},
			&Host{
				ID:            "host-2",
				Registered:    true,
				LastContacted: time.Now(),
			},
			&Host{
				ID:            "host-3",
				Registered:    true,
				LastContacted: time.Now(),
			},
		},
		Services: map[string]*service.Service{
			"host-1": &service.Service{Process: &os.Process{Pid: 1234}},
			"host-2": &service.Service{Process: &os.Process{Pid: 1234}},
			"host-3": &service.Service{Process: &os.Process{Pid: 1234}},
		},
	}
	degradedHost = &Cluster{
		ID: "test-cluster",
		Hosts: []*Host{
			&Host{
				ID:         "host-1",
				Registered: true,
			},
			&Host{
				ID:            "host-2",
				Registered:    true,
				LastContacted: time.Now(),
			},
			&Host{
				ID:            "host-3",
				Registered:    true,
				LastContacted: time.Now(),
			},
		},
		Services: map[string]*service.Service{
			"host-1": &service.Service{Process: &os.Process{Pid: 1234}},
			"host-2": &service.Service{Process: &os.Process{Pid: 1234}},
			"host-3": &service.Service{Process: &os.Process{Pid: 1234}},
		},
	}
)

func TestCluster(t *testing.T) {
	assert.Equal(t, CONVERGING, converging.State())
	assert.Equal(t, STARTING, starting.State())
	assert.Equal(t, STARTED, started.State())
	assert.Equal(t, DEGRADED, degradedHost.State())
}
