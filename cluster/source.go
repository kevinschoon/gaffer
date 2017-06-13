package cluster

import (
	"fmt"
	"strings"
)

// Source gets and sets a configuration
type Source interface {
	Get() (*Config, error)
	Set(*Config) error
}

func NewSource(pattern string) (Source, error) {
	switch {
	case strings.Contains(pattern, "file://"):
		return FileSource{Path: strings.Replace(pattern, "file://", "", -1)}, nil
	case strings.Contains(pattern, "http"):
		return NewHTTPSource(pattern)
	}
	return nil, fmt.Errorf("unknown file source: %s", pattern)
}
