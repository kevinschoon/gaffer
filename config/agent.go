package config

import (
	"fmt"
	"github.com/vektorlab/gaffer/process"
	"go.uber.org/zap"
	"io/ioutil"
)

var (
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
			Name:  "MESOS_ISOLATION",
			Value: "cgroups/cpu,cgroups/mem,cgroups/pids,namespaces/pid,filesystem/shared,filesystem/linux,docker/runtime,volume/sandbox_path",
		},
		&Option{
			Name:  "MESOS_WORK_DIR",
			Value: "/tmp/mesos",
		},
		&Option{
			Name:  "MESOS_MASTER",
			Value: "",
		},
	}
)

// Agent represents a single Mesos Agent process
// TODO Consider monitoring the state of each agent
// the same way Master/Zookeeper is
type Agent struct {
	Options []*Option `json:"options"`
}

func (m *Agent) Process(log *zap.Logger) (*process.Process, error) {
	tmpDir, err := ioutil.TempDir("", "gaffer")
	if err != nil {
		return nil, err
	}
	proc := process.NewProcess(
		log,
		"mesos-agent",
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

func NewAgent(cluster *Cluster) *Agent {
	agent := &Agent{
		Options: make([]*Option, len(DefaultAgentOptions)),
	}
	copy(agent.Options, DefaultAgentOptions)
	merge(agent.Options, cluster.AgentOptions)
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
	findOpt("MESOS_MASTER", agent.Options).Value = zkStr
	return agent
}
