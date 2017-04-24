package main

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	//"net"
	"os"
	"strings"
	//"github.com/satori/go.uuid"
)

const ZKClass string = "org.apache.zookeeper.server.quorum.QuorumPeerMain"

var (
	ZKLog4J []byte = []byte(`
zookeeper.root.logger=INFO, CONSOLE
zookeeper.console.threshold=INFO
log4j.rootLogger=${zookeeper.root.logger}
log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender
log4j.appender.CONSOLE.Threshold=${zookeeper.console.threshold}
log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout
log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} [myid:%X{myid}] - %-5p [%t:%C{1}@%L] - %m%n`)
)

var (
	// TODO infer all of this God forsaken shit
	DefaultZKClasspath = []string{
		"/usr/share/zookeeper/lib/jline-0.9.94.jar",
		"/usr/share/zookeeper/lib/log4j-1.2.16.jar",
		"/usr/share/zookeeper/lib/netty-3.10.5.Final.jar",
		"/usr/share/zookeeper/lib/slf4j-api-1.6.1.jar",
		"/usr/share/zookeeper/lib/slf4j-log4j12-1.6.1.jar",
		"/usr/share/zookeeper/lib/zookeeper-3.4.9.jar",
	}
	DefaultZKOptions = []*Option{
		&Option{
			Name:  "tickTime",
			Value: "2000",
		},
		&Option{
			Name:  "initLimit",
			Value: "10",
		},
		&Option{
			Name:  "syncLimit",
			Value: "5",
		},
		&Option{
			Name:  "clientPort",
			Value: "2181",
		},
		&Option{
			Name:  "dataDir",
			Value: "/tmp/zookeeper",
		},
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
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Size          int          `json:"size"`
	ZKOptions     []*Option    `json:"zk_options"`
	MasterOptions []*Option    `json:"master_options"`
	AgentOptions  []*Option    `json:"agent_options"`
	Masters       []*Master    `json:"masters"`
	Zookeepers    []*Zookeeper `json:"zookeepers"`
}

func (c Cluster) ZKReady() bool {
	if c.Zookeepers == nil {
		return false
	}
	for i := 0; i < c.Size; i++ {
		if len(c.Zookeepers) >= i+1 {
			if !c.Zookeepers[i].Running {
				return false
			}
		}
	}
	return true
}

func (c Cluster) MesosReady() bool {
	if c.Masters == nil {
		return false
	}
	for i := 0; i < c.Size; i++ {
		if len(c.Masters) >= i+1 {
			if !c.Masters[i].Running {
				return false
			}
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

func (zk *Zookeeper) Process(log *zap.Logger) *Process {
	d, _ := ioutil.TempDir("", "g")
	cfg, _ := ioutil.TempFile(d, "g")
	for _, opt := range zk.Options {
		cfg.WriteString(fmt.Sprintf("%s=%s\n", opt.Name, opt.Value))
	}
	l, _ := ioutil.TempFile(d, "g")
	l.Write(ZKLog4J)
	//cp := []string{}
	//for _, s := range zk.Classpath {
	//	cp = append(cp, s)
	//}
	//cp = append(cp, l.Name())
	return NewProcess(
		log,
		zk.Java,
		"-cp",
		strings.Join(zk.Classpath, ":"),
		fmt.Sprintf("-Dlog4j.configuration=file://%s", l.Name()),
		ZKClass,
		cfg.Name(),
	)
}

func NewZookeeper(cluster *Cluster) (*Zookeeper, error) {
	// Lookup the hostname to identify this host
	// TODO Consider comparing interface addresses
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	// Range any pre-defined Zookeepers
	for _, zk := range cluster.Zookeepers {
		// Check if Zookeeper has already been configured
		// to run on this host.
		if zk.Hostname == hostname {
			// Return the Zookeeper configuration for this host
			return zk, nil
		}
	}
	// Create a new Zookeeper configuration for this host
	zk := &Zookeeper{
		Hostname:  hostname,
		Java:      "java",
		Classpath: DefaultZKClasspath,
		Options:   DefaultZKOptions,
	}
	// Append this Zookeeper to the cluster configuration
	// TODO Potential race condition, consider allowing only
	// pre-defined Zookeepers
	cluster.Zookeepers = append(cluster.Zookeepers, zk)
	// Override all of the default options if any where specified
	// TODO Should likely merge instead
	if cluster.ZKOptions != nil {
		zk.Options = cluster.ZKOptions
	}
	return zk, nil
}
