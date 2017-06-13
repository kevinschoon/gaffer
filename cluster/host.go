package cluster

import (
	"regexp"
)

type Filter func(*Host) bool

type Hosts []*Host

type Host struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

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
		return h.Port == p
	}
}
