package cluster

type Cluster struct {
	Services []*Service `json:"services"`
}
