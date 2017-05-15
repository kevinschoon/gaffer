package mesos

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/service"
)

type Mesos struct {
	ZookeeperPort int            `json:"zookeeper_port"`
	ZookeeperConf *service.File  `json:"zookeeper_conf"`
	MasterEnv     []*service.Env `json:"master_env"`
}

func (m Mesos) Master(c *cluster.Cluster) *service.Service {
	svc := &service.Service{
		ID:          "mesos-master",
		Args:        []string{"mesos-master"},
		Environment: []*service.Env{},
	}
	for _, env := range m.MasterEnv {
		svc.Environment = append(svc.Environment, &service.Env{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	if env := svc.Env("MESOS_ZK"); env != nil {
		env.Value = ZKString(c, m.ZookeeperPort)
	} else {
		svc.Environment = append(svc.Environment, &service.Env{"MESOS_ZK", ZKString(c, m.ZookeeperPort)})
	}
	if env := svc.Env("MESOS_QUORUM"); env != nil {
		env.Value = fmt.Sprintf("%d", Quorum(len(c.Hosts)))
	} else {
		svc.Environment = append(svc.Environment, &service.Env{"MESOS_QUORUM", fmt.Sprintf("%d", Quorum(len(c.Hosts)))})
	}
	return svc
}

func (m Mesos) Zookeeper(c *cluster.Cluster, id int) *service.Service {
	svc := &service.Service{
		ID:    "zookeeper",
		Args:  []string{"zkServer.sh", "start-foreground", "zoo.cfg"},
		Files: []*service.File{},
	}
	svc.Files = append(svc.Files, &service.File{
		Path:    "./zookeeper/myid",
		Content: []string{fmt.Sprintf("%d", id+1)},
	})
	content := []string{
		"tickTime=2000",
		"initLimit=10",
		"syncLimit=5",
		fmt.Sprintf("clientPort=%d", m.ZookeeperPort),
		"dataDir=./zookeeper",
	}
	for i, host := range c.Hosts {
		content = append(content, fmt.Sprintf("server.%d=%s:2888:3888", i+1, host.Hostname))
	}
	svc.Files = append(svc.Files, &service.File{
		Path:    "zookeeper.conf",
		Content: content,
	})
	return svc
}

func (m Mesos) Update(c *cluster.Cluster) map[string][]*service.Service {
	services := map[string][]*service.Service{}
	for i, host := range c.Hosts {
		services[host.ID] = []*service.Service{
			m.Master(c),
			m.Zookeeper(c, i),
		}
	}
	return services
}
