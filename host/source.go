package host

import (
	"fmt"
	"strings"
)

// Source gets and sets a host configuration.
type Source interface {
	Get() (*Config, error)
	Set(*Config) error
}

func NewSource(pattern string) (Source, error) {
	switch {
	case strings.Contains(pattern, "gaffer://"):
		ls := LocalSource{pattern: pattern}
		_, err := ls.Get()
		if err != nil {
			return nil, err
		}
		return ls, nil
	}
	return nil, fmt.Errorf("unknown host source: %s", pattern)
}
