package main

import (
	"encoding/json"
	//"github.com/satori/go.uuid"
)

const ZKClass string = "org.apache.zookeeper.server.quorum.QuorumPeerMain"

var (
	// TODO infer all of this God forsaken shit
	DefaultZKClassPath = []string{
		"/usr/share/zookeeper/lib/jline-0.9.94.jar",
		"/usr/share/zookeeper/lilog4j-1.2.16.jar",
		"/usr/share/zookeeper/lib/netty-3.10.5.Final.jar",
		"/usr/share/zookeeper/lib/slf4j-api-1.6.1.jar",
		"/usr/share/zookeeper/lib/slf4j-log4j12-1.6.1.jar",
		"/usr/share/zookeeper/lib/zookeeper-3.4.9.jar",
	}
)

// Option represents a process configuration
type Option struct {
	Name  string          `json:"name"`
	Value string          `json:"value"`
	Data  json.RawMessage `json:"data"`
}

// Cluster represents the overall configuration of a Mesos cluster
type Cluster struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Size             int          `json:"size"`
	ZookeeperOptions []*Option    `json:"zookeeper_options"`
	MasterOptions    []*Option    `json:"master_options"`
	AgentOptions     []*Option    `json:"agent_options"`
	Masters          []*Master    `json:"masters"`
	Zookeepers       []*Zookeeper `json:"zookeepers"`
}

func (c Cluster) ZKReady() bool {
	if c.Zookeepers == nil {
		return false
	}
	for i := 0; i < c.Size; i++ {
		if !c.Zookeepers[i].Running {
			return false
		}
	}
	return true
}

func (c Cluster) MesosReady() bool {
	if c.Masters == nil {
		return false
	}
	for i := 0; i < c.Size; i++ {
		if !c.Masters[i].Running {
			return false
		}
	}
	return true
}

// Master represents a single Mesos Master process
type Master struct {
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	Running  bool      `json:"running"`
	Options  []*Option `json:"options"`
}

// Agent represents a single Mesos Agent process
type Agent struct {
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	Running  bool      `json:"running"`
	Options  []*Option `json:"options"`
}

// Zookeeper represents a single Zookeeper process
type Zookeeper struct {
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	Running   bool      `json:"running"`
	Java      string    `json:"java"`
	Classpath []string  `json:"classpath"`
	Options   []*Option `json:"options"`
}
