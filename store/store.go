package store

import (
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/store/sql"
)

type Store interface {
	Query(*query.Query) (*query.Response, error)
	Close() error
}

var _ Store = &sql.SQLStore{}

func NewSQLStore(name, connect string, init bool) (Store, error) {
	store, err := sql.New(name, connect, init)
	if err != nil {
		return nil, err
	}
	return store, nil
}
