package user

import (
	"fmt"
	"strings"
)

type User struct {
	ID    string
	Token string
}

func FromString(str string) (*User, error) {
	split := strings.Split(str, ":")
	if len(split) != 2 {
		return nil, fmt.Errorf("bad auth %s", str)
	}
	return &User{split[0], split[1]}, nil
}
