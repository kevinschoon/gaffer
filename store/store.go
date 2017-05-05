package store

import (
	"fmt"
	"github.com/vektorlab/gaffer/store/query"
)

type ErrUserNotFound struct {
	id string
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("User %s not found", e.id)
}

type Store interface {
	Query(*query.Query) (*query.Response, error)
	Close() error
}

var _ Store = &SQLStore{}
