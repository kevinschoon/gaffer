package store

import (
	"fmt"
	"github.com/vektorlab/gaffer/store/http"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/store/sql"
	"github.com/vektorlab/gaffer/user"
	"net/url"
	"strings"
)

type Store interface {
	Query(*query.Query) (*query.Response, error)
	Close() error
}

var (
	_ Store = &sql.SQLStore{}
	_ Store = &http.Client{}
)

// NewStore will return a new store by evaluating the pattern.
// We support the following formats:
// HTTP store:
// http://[user:pass]@hostname[:port]
// SQL Store:
// sqlite:///gaffer.db
func NewStore(pattern string) (Store, error) {
	switch {
	case strings.Contains(pattern, "http://") || strings.Contains(pattern, "https://"):
		return fromHTTP(pattern)
	case strings.Contains(pattern, "sqlite://"):
		return fromSQLite(strings.Replace(pattern, "sqlite://", "", -1))
	}
	return nil, fmt.Errorf("cannot read pattern %s", pattern)
}

func fromHTTP(pattern string) (Store, error) {
	parsed, err := url.Parse(pattern)
	if err != nil {
		return nil, err
	}
	var u *user.User
	if parsed.User != nil {
		name := parsed.User.Username()
		token, _ := parsed.User.Password()
		u = &user.User{name, token}
	}
	return http.New(parsed.String(), u), nil
}

func fromSQLite(pattern string) (Store, error) {
	return sql.New(pattern)
}
