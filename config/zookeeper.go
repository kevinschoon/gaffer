package config

import (
	"fmt"
	"github.com/vektorlab/gaffer/process"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"strings"
)

const ZKClass string = "org.apache.zookeeper.server.quorum.QuorumPeerMain"

var (
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
)

// Zookeeper represents a single Zookeeper process
type Zookeeper struct {
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	Running   bool      `json:"running"`
	Java      string    `json:"java"`
	Classpath []string  `json:"classpath"`
	Options   []*Option `json:"options"`
}

func (zk *Zookeeper) Process(log *zap.Logger) (*process.Process, error) {
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
	return process.NewProcess(
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
