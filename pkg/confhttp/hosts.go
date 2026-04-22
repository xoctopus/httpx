package confhttp

import (
	"context"
	"fmt"
	"maps"
	randv2 "math/rand/v2"
	"net"
	"slices"
	"strings"

	"github.com/xoctopus/x/codex"
)

// Hosts host:ip
type Hosts map[string]map[string]struct{}

func (hosts Hosts) WrapDialContext(dc DailContext) DailContext {
	return func(ctx context.Context, network string, addr string) (net.Conn, error) {
		if len(hosts) == 0 {
			return dc(ctx, network, addr)
		}

		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
			port = "80"
		}

		if ips, ok := hosts[host]; ok && len(ips) > 0 {
			idx := randv2.IntN(len(ips)) - 1
			ip := slices.Collect(maps.Keys(ips))[idx]
			resolved := net.JoinHostPort(ip, port)
			return dc(ctx, network, resolved)
		}

		return dc(ctx, network, addr)
	}
}

func (hosts Hosts) AddHostAlias(alias HostAlias) {
	if alias.IsZero() {
		return
	}

	for _, hostname := range alias.Hostnames {
		if hosts[hostname] == nil {
			hosts[hostname] = make(map[string]struct{})
		}
		hosts[hostname][alias.IP.String()] = struct{}{}
	}
}

// ParseHostAlias format: [ip_address]hostname1,hostname2
func ParseHostAlias(text string) (*HostAlias, error) {
	if text == "" {
		return nil, codex.Errorf(ERROR__HOST_ALIAS_INVALID_INPUT, "empty host alias")
	}

	ha := &HostAlias{}

	if strings.IndexByte(text, '[') == 0 {
		// ipv6
		end := strings.IndexByte(text, ']')
		if end == -1 {
			return nil, codex.Errorf(ERROR__HOST_ALIAS_INVALID_IPV6_ADDR, "should wrapped by '[]'")
		}

		ha.IP = net.ParseIP(text[1:end])

		text = text[end:]

		end = strings.IndexByte(text, ':')
		if end == -1 {
			return nil, codex.Errorf(ERROR__HOST_ALIAS_INVALID_IPV6_ADDR, "should end with ':'")
		}
		text = text[end+1:]
	} else {
		end := strings.IndexByte(text, ':')
		if end == -1 {
			return nil, codex.Errorf(ERROR__HOST_ALIAS_INVALID_IPV6_ADDR, "should end with ':'")
		}
		ha.IP = net.ParseIP(text[0:end])
		text = text[end+1:]
	}

	if text == "" {
		return nil, fmt.Errorf("invalid host alias")
	}

	ha.Hostnames = strings.Split(text, ",")

	return ha, nil
}

// HostAlias ip-> hostnames
type HostAlias struct {
	IP        net.IP
	Hostnames []string
}

func (ha HostAlias) IsZero() bool {
	return len(ha.IP) == 0 || len(ha.Hostnames) == 0
}

func (ha *HostAlias) UnmarshalText(data []byte) error {
	x, err := ParseHostAlias(string(data))
	if err != nil {
		return err
	}
	*ha = *x
	return nil
}

func (ha HostAlias) MarshalText() ([]byte, error) {
	return []byte(ha.String()), nil
}

func (ha HostAlias) String() string {
	s := strings.Builder{}

	if ip := ha.IP.To4(); ip != nil {
		s.WriteString(ip.String())
	} else {
		s.WriteString("[")
		s.WriteString(ha.IP.String())
		s.WriteString("]")
	}

	s.WriteString(":")

	for i, hostname := range ha.Hostnames {
		if i > 0 {
			s.WriteString(",")
		}
		s.WriteString(hostname)
	}

	return s.String()
}
