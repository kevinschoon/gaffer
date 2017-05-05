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
	User   *user.User `json:"-"`
	Create *Create    `json:"create"`
	Read   *Read      `json:"read"`
	Update *Update    `json:"update"`
	Delete *Delete    `json:"delete"`
}

type Create struct {
	Clusters []*cluster.Cluster `json:"clusters"`
}

type Read struct {
	ID string `json"id"`
}

type Update struct {
	Clusters []*cluster.Cluster `json:"clusters"`
}

type Delete struct {
	ID string `json:"id"`
}

type Response struct {
	Clusters []*cluster.Cluster `json:"clusters"`
	User     *user.User         `json:"-"`
}
