package store

import (
	"github.com/vektorlab/gaffer/store/http"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/store/sql"
	"github.com/vektorlab/gaffer/user"
)

type Store interface {
	Query(*query.Query) (*query.Response, error)
	Close() error
}

var (
	_ Store = &sql.SQLStore{}
	_ Store = &http.Client{}
)

func NewSQLStore(name, connect string, init bool) (Store, error) {
	store, err := sql.New(name, connect, init)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func NewHTTPStore(endpoint string, u *user.User) Store {
	return http.New(endpoint, u)
}
