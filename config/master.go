package config

import (
	"fmt"
	"github.com/vektorlab/gaffer/process"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
)

var (
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
)

// Master represents a single Mesos Master process
type Master struct {
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	Running  bool      `json:"running"`
	Options  []*Option `json:"options"`
}

func (m *Master) Process(log *zap.Logger) (*process.Process, error) {
	tmpDir, err := ioutil.TempDir("", "gaffer")
	if err != nil {
		return nil, err
	}
	proc := process.NewProcess(
		log,
		"mesos-master",
	)
	for _, opt := range m.Options {
		proc.Env[opt.Name] = opt.Value
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
