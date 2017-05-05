package query

import (
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/user"
)

type Type string

const (
	CREATE    Type = "CREATE"
	UPDATE    Type = "WRITE"
	READ      Type = "READ"
	DELETE    Type = "DELETE"
	READ_USER Type = "READ_USER"
)

// Query is used to request an action against a store
type Query struct {
	Type   Type       `json:"type"`
	User   *user.User `json:"user"`
	Create struct {
		Clusters []*cluster.Cluster
	} `json:"create"`
	Read struct {
		ID string `json:"id"`
	} `json:"read"`
	ReadUser struct {
		User *user.User `json:"user"`
	} `json:"read_user"`
	Update struct {
		Clusters []*cluster.Cluster
	} `json:"write"`
	Delete struct {
		ID string `json:"id"`
	} `json:"delete"`
}
type Response struct {
	Clusters []*cluster.Cluster `json:"clusters"`
	User     *user.User         `json:"-"`
}
