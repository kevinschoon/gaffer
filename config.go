package main

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	//"net"
	"math"
	"os"
	"strings"
	//"github.com/satori/go.uuid"
)

const ZKClass string = "org.apache.zookeeper.server.quorum.QuorumPeerMain"

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
	DefaultZKLog4j []byte = []byte(`
zookeeper.root.logger=INFO, CONSOLE
zookeeper.console.threshold=INFO
log4j.rootLogger=${zookeeper.root.logger}
log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender
log4j.appender.CONSOLE.Threshold=${zookeeper.console.threshold}
log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout
log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} [myid:%X{myid}] - %-5p [%t:%C{1}@%L] - %m%n`)

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

	DefaultMasterOptions = []*Option{
		&Option{
			Name:  "MESOS_ZK",
			Value: "",
		},
		&Option{
			Name:  "MESOS_QUORUM",
			Value: "",
		},
		&Option{
			Name:  "MESOS_LOGGING_LEVEL",
			Value: "INFO",
		},
		&Option{
			Name:  "MESOS_WORK_DIR",
			Value: "/tmp/mesos",
		},
	}

	DefaultAgentOptions = []*Option{
		&Option{
			Name:  "MESOS_CONTAINERIZERS",
			Value: "mesos,docker",
		},
		&Option{
			Name:  "MESOS_EXECUTOR_REGISTRATION_TIMEOUT",
			Value: "5mins",
		},
		&Option{
			Name:  "MESOS_IMAGE_PROVIDERS",
			Value: "DOCKER,APPC",
		},
		&Option{
			Name:  "MESOS_MASTER",
			Value: "",
		},
	}
)

// Option represents a process configuration
type Option struct {
	Name  string          `json:"name"`
	Value string          `json:"value"`
	Data  json.RawMessage `json:"data"`
}

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

// Master represents a single Mesos Master process
type Master struct {
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	Running  bool      `json:"running"`
	Options  []*Option `json:"options"`
}

func (m *Master) Process(log *zap.Logger) (*Process, error) {
	tmpDir, err := ioutil.TempDir("", "gaffer")
	if err != nil {
		return nil, err
	}
	proc := NewProcess(
		log,
		"mesos-master",
	)
	for _, opt := range m.Options {
		fmt.Println(opt)
		proc.env[opt.Name] = opt.Value
		if opt.Data != nil {
			// TODO Do we need flags or are env opts with file:///... good enough?
			tmpCfg, err := ioutil.TempFile(tmpDir, "gaffer")
			if err != nil {
				return nil, err
			}
			// Writes out any raw JSON configuration
			_, err = tmpCfg.Write(opt.Data)
			if err != nil {
				return nil, err
			}
		}
	}
	return proc, nil
}

func NewMaster(cluster *Cluster) (*Master, error) {
	// TODO Consider comparing interface addresses
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	var master *Master
	// Range any pre-defined Masters
	for _, other := range cluster.Masters {
		// Check if Mesos Master was already configured
		if other.Hostname == hostname {
			// return the master
			master = other
		}
	}
	if master == nil {
		master = &Master{
			Hostname: hostname,
			Options:  make([]*Option, len(DefaultMasterOptions)),
		}
		copy(master.Options, DefaultMasterOptions)
		cluster.Masters = append(cluster.Masters, master)
	}
	// Config below is loaded dynamically
	merge(master.Options, cluster.MasterOptions)
	zkStr := "zk://"
	for i, zk := range cluster.Zookeepers {
		// TODO support ZK port numbers
		zkStr += fmt.Sprintf("%s:2181", zk.Hostname)
		if i != cluster.Size-1 {
			zkStr += ","
		} else {
			zkStr += "/mesos"
		}
	}
	findOpt("MESOS_ZK", master.Options).Value = zkStr
	findOpt("MESOS_QUORUM", master.Options).Value = fmt.Sprintf("%d", cluster.Quorum())
	return master, nil
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

func (zk *Zookeeper) Process(log *zap.Logger) (*Process, error) {
	tmpDir, err := ioutil.TempDir("", "gaffer")
	if err != nil {
		return nil, err
	}
	tmpCfg, err := ioutil.TempFile(tmpDir, "gaffer")
	if err != nil {
		return nil, err
	}
	for _, opt := range zk.Options {
		_, err = tmpCfg.WriteString(fmt.Sprintf("%s=%s\n", opt.Name, opt.Value))
		if err != nil {
			return nil, err
		}
	}
	tmpCfgL4j, err := ioutil.TempFile(tmpDir, "gaffer")
	if err != nil {
		return nil, err
	}
	_, err = tmpCfgL4j.Write(DefaultZKLog4j)
	if err != nil {
		return nil, err
	}
	return NewProcess(
		log,
		zk.Java,
		"-cp",
		strings.Join(zk.Classpath, ":"),
		fmt.Sprintf("-Dlog4j.configuration=file://%s", tmpCfgL4j.Name()),
		ZKClass,
		tmpCfg.Name(),
	), nil
}

// TODO format server.x= configuration with multiple
func NewZookeeper(cluster *Cluster) (*Zookeeper, error) {
	// Lookup the hostname to identify this host
	// TODO Consider comparing interface addresses
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	var zk *Zookeeper
	// Range any pre-defined Zookeepers
	for _, other := range cluster.Zookeepers {
		// Check if Zookeeper has already been configured
		// to run on this host.
		if other.Hostname == hostname {
			// Return the Zookeeper configuration for this host
			zk = other
		}
	}
	if zk == nil {
		zk = &Zookeeper{
			Hostname:  hostname,
			Java:      "java",
			Classpath: DefaultZKClasspath,
			Options:   make([]*Option, len(DefaultZKOptions)),
		}
		// Create a new Zookeeper configuration for this host
		copy(zk.Options, DefaultZKOptions)
		// Append this Zookeeper to the cluster configuration
		// TODO Potential race condition, consider allowing only
		// pre-defined Zookeepers
		cluster.Zookeepers = append(cluster.Zookeepers, zk)
	}
	// Config below is loaded dynamically
	// Purge any server.n options
	opts := []*Option{}
	for _, opt := range zk.Options {
		if !strings.Contains(opt.Name, "server.") {
			opts = append(opts, opt)
		}
	}
	zk.Options = opts
	for i, other := range cluster.Zookeepers {
		zk.Options = append(zk.Options, &Option{
			Name: fmt.Sprintf("server.%d", i+1),
			// TODO handle ports
			Value: fmt.Sprintf("%s:2888:3888", other.Hostname),
		})
	}
	merge(zk.Options, cluster.ZKOptions)
	return zk, nil
}

func findOpt(name string, opts []*Option) *Option {
	for _, opt := range opts {
		if opt.Name == name {
			return opt
		}
	}
	return nil
}

func merge(opts []*Option, other []*Option) {
	if other == nil {
		return
	}
	for _, opt := range other {
		if o := findOpt(opt.Name, opts); o != nil {
			o.Name = opt.Name
			o.Value = opt.Value
			o.Data = opt.Data
		} else {
			opts = append(opts, opt)
		}
	}
}
