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
	Create *Create `json:"create"`
	Read   *Read   `json:"read"`
	Update *Update `json:"update"`
	Delete *Delete `json:"delete"`
}

type Response struct {
	Create *CreateResponse `json:"create"`
	Read   *ReadResponse   `json:"read"`
	Update *UpdateResponse `json:"update"`
	Delete *DeleteResponse `json:"delete"`
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

type Update struct {
	Host    *host.Host       `json:"host"`
	Service *service.Service `json:"service"`
}

type UpdateResponse struct {
	Host    *host.Host       `json:"host"`
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
	if query.Update != nil {
		if query.Update.Service != nil {
			if query.Update.Host == nil {
				return ErrInvalidQuery{"must specify host with service"}
			}
		}
		reqs++
	}
	if query.Delete != nil {
		reqs++
	}
	if reqs != 1 {
		return ErrInvalidQuery{"must specify one change per request"}
	}
	return nil
}
