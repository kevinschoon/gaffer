package cluster

import (
	"regexp"
)

type Pattern struct {
	Regexp *regexp.Regexp
}

func (p Pattern) Match(host *Host) bool {
	if p.Regexp != nil {
		return p.Regexp.MatchString(host.Name)
	}
	return false
}

type Hosts []*Host

type Host struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func (hosts Hosts) Match(pattern Pattern) Hosts {
	matched := Hosts{}
	for _, host := range hosts {
		if pattern.Match(host) {
			matched = append(matched, host)
		}
	}
	return matched
}

func ByName(pattern string) Pattern {
	return Pattern{
		Regexp: regexp.MustCompile(pattern),
	}
}
