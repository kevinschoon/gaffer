package host

import (
	"fmt"
	"net"
	"os"
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

func Self() (*Host, error) {
	host := &Host{}
	name, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	host.Name = name
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		// TODO: Make more robust
		for _, addr := range addrs {
			ip, ok := addr.(*net.IPAddr)
			if ok {
				if !ip.IP.IsLoopback() {
					if i := ip.IP.To4(); i != nil {
						host.Address = i.String()
						return host, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("cannot detect ip address")
}

func SelfMust() *Host {
	host, err := Self()
	if err != nil {
		panic(fmt.Errorf("couldn't detect host from self %s", err.Error()))
	}
	return host
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
