package host

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Filter func(*Host) bool

type Hosts []*Host

func (hosts Hosts) Filter(filters ...Filter) Hosts {
	matched := Hosts{}
loop:
	for _, host := range hosts {
		for _, filter := range filters {
			if filter(host) {
				matched = append(matched, host)
				continue loop
			}
		}
	}
	return matched
}

func Any() Filter {
	return func(*Host) bool { return true }
}

func ByName(pattern string) Filter {
	return func(h *Host) bool {
		return regexp.MustCompile(pattern).MatchString(h.Name)
	}
}

func ByPort(p int) Filter {
	return func(h *Host) bool {
		return h.Port == int32(p)
	}
}

func New(pattern string) (*Host, error) {
	if !strings.Contains(pattern, "gaffer://") {
		return nil, fmt.Errorf("bad host pattern: %s", pattern)
	}
	split := strings.SplitN(strings.Replace(pattern, "gaffer://", "", -1), ":", 2)
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, err
	}
	return &Host{Name: split[0], Port: int32(port)}, nil
}
