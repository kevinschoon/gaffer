package host

import (
	"fmt"
	"net"
	"os"
	"regexp"
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

func ByIP(ip string) Filter {
	return func(h *Host) bool {
		return h.Address == ip
	}
}

func ByMAC(mac string) Filter {
	return func(h *Host) bool {
		return h.Mac == mac
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
	// Must have at least two interfaces
	// assuming the first host interface
	// is a loopback.
	if len(ifaces) > 1 {
		// Range all interfaces except loopback
		for _, iface := range ifaces[1:] {
			// Ignore any interfaces which are not "up"
			if !strings.Contains(iface.Flags.String(), "up") {
				continue
			}
			// List all addresses on the interface
			addrs, err := iface.Addrs()
			if err != nil {
				return nil, err
			}
			for _, addr := range addrs {
				ip, ok := addr.(*net.IPNet)
				if ok {
					// Assume the first interface with an IPv4
					// is the what we are bound to. TODO: This
					// should be configurable.
					if i := ip.IP.To4(); i != nil {
						host.Mac = iface.HardwareAddr.String()
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
		panic(fmt.Errorf("couldn't detect host from self: %s", err.Error()))
	}
	return host
}
