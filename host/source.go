package host

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
	case strings.Contains(pattern, "s3://"):
		split := strings.Split(strings.Replace(pattern, "s3://", "", -1), "/")
		var (
			bucket string
			key    string
		)
		if len(split) < 2 {
			return nil, fmt.Errorf("bad s3 url: %s", pattern)
		}
		bucket = split[0]
		key = strings.Join(split[1:], "/")
		return NewS3Source(bucket, key)
	}
	return nil, fmt.Errorf("unknown file source: %s", pattern)
}
