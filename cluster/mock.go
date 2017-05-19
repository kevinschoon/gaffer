package cluster

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster/service"
)

type Mock struct {
	Duration int
}

func (m Mock) Update(other Cluster) map[string][]*service.Service {
	services := map[string][]*service.Service{}
	for _, host := range other.Hosts {
		services[host.ID] = []*service.Service{
			&service.Service{
				ID:   "CMD1",
				Args: []string{"sleep", fmt.Sprintf("%d", m.Duration)},
			},
			&service.Service{
				ID:   "CMD2",
				Args: []string{"sleep", fmt.Sprintf("%d", m.Duration)},
			},
		}
	}
	return services
}
