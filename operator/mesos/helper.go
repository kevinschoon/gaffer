package mesos

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"math"
)

func ZKString(c *cluster.Cluster, port int) string {
	zkStr := "zk://"
	for i, zk := range c.Hosts {
		zkStr += fmt.Sprintf("%s:%d", zk.Hostname, port)
		if i != len(c.Hosts)-1 {
			zkStr += ","
		} else {
			zkStr += "/mesos"
		}
	}
	return zkStr
}

func Quorum(size int) int {
	return int(math.Floor(float64(size)) + .5)
}
