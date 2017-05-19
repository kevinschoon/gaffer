package cluster

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/vektorlab/gaffer/cluster/host"
	"os"
	"testing"
	"time"
)

func newCluster() Cluster {
	return Cluster{
		ID: "test-cluster",
		Hosts: []*host.Host{
			&host.Host{
				ID:            "host-1",
				Hostname:      "host-1",
				Registered:    true,
				LastContacted: time.Now(),
			},
			&host.Host{
				ID:            "host-2",
				Hostname:      "host-2",
				Registered:    true,
				LastContacted: time.Now(),
			},
			&host.Host{
				ID:            "host-3",
				Hostname:      "host-3",
				Registered:    true,
				LastContacted: time.Now(),
			},
		},
	}
}

func TestMesosOperator(t *testing.T) {
	m := &Mesos{ZookeeperPort: 2181}
	c := newCluster()
	c.Services = m.Update(c)
	assert.Len(t, c.Services, 3)
	assert.Equal(t, c.Service("host-1", "mesos-master").Env("MESOS_QUORUM").Value, "3")
	assert.Equal(t, c.Service("host-1", "mesos-master").Env("MESOS_ZK").Value, "zk://host-1:2181,host-2:2181,host-3:2181/mesos")
	assert.Equal(t, c.Service("host-2", "mesos-master").Env("MESOS_QUORUM").Value, "3")
	assert.Equal(t, c.Service("host-2", "mesos-master").Env("MESOS_ZK").Value, "zk://host-1:2181,host-2:2181,host-3:2181/mesos")
	assert.Equal(t, c.Service("host-3", "mesos-master").Env("MESOS_QUORUM").Value, "3")
	assert.Equal(t, c.Service("host-3", "mesos-master").Env("MESOS_ZK").Value, "zk://host-1:2181,host-2:2181,host-3:2181/mesos")
	json.NewEncoder(os.Stdout).Encode(c)
	c.Hosts = c.Hosts[:1]
	c.Services = m.Update(c)
	assert.Len(t, c.Services, 1)
	assert.Equal(t, c.Service("host-1", "mesos-master").Env("MESOS_QUORUM").Value, "1")
	assert.Equal(t, c.Service("host-1", "mesos-master").Env("MESOS_ZK").Value, "zk://host-1:2181/mesos")
	json.NewEncoder(os.Stdout).Encode(c)
}
