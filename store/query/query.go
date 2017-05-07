package query

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
)

type ErrInvalidQuery struct {
	Message string
}

func (e ErrInvalidQuery) Error() string { return fmt.Sprintf("Invalid query: %s", e.Message) }

// Query is used to request an action against a store
type Query struct {
	Create        *Create        `json:"create"`
	Read          *Read          `json:"read"`
	UpdateHost    *UpdateHost    `json:"update_host"`
	UpdateService *UpdateService `json:"update_service"`
	Delete        *Delete        `json:"delete"`
}

type Response struct {
	Create        *CreateResponse        `json:"create"`
	Read          *ReadResponse          `json:"read"`
	UpdateHost    *UpdateHostResponse    `json:"update_host"`
	UpdateService *UpdateServiceResponse `json:"update_service"`
	Delete        *DeleteResponse        `json:"delete"`
}

type Create struct {
	Cluster *cluster.Cluster `json:"cluster"`
}

type CreateResponse struct {
	Cluster *cluster.Cluster `json:"cluster"`
}

type Read struct{}

type ReadResponse struct {
	Cluster *cluster.Cluster `json:"cluster"`
}

type UpdateHost struct {
	Host *host.Host `json:"host"`
}

type UpdateHostResponse struct {
	Host *host.Host `json:"host"`
}

type UpdateService struct {
	HostID  string           `json:"host_id"`
	Service *service.Service `json:"service"`
}
type UpdateServiceResponse struct {
	Service *service.Service `json:"service"`
}

type Delete struct {
	ServiceID string `json:"service_id"`
	HostID    string `json:"host_id"`
}

type DeleteResponse struct{}

func Validate(query *Query) error {
	if query == nil {
		return ErrInvalidQuery{}
	}
	var reqs int
	if query.Create != nil {
		reqs++
		if query.Create.Cluster == nil {
			return ErrInvalidQuery{"property cluster not specified"}
		}
	}
	if query.Read != nil {
		reqs++
	}
	if query.UpdateHost != nil {
		reqs++
	}
	if query.UpdateService != nil {
		reqs++
	}
	if query.Delete != nil {
		reqs++
	}
	if reqs != 1 {
		return ErrInvalidQuery{"must specify at least one request"}
	}
	return nil
}
