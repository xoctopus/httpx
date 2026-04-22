package confhttp_test

import (
	"net"
	"testing"

	"github.com/xoctopus/httpx/pkg/confhttp"
)

func TestHosts(t *testing.T) {
	ha := confhttp.HostAlias{
		IP:        net.IPv4(1, 1, 1, 1),
		Hostnames: []string{"testV4_1", "testV4_2"},
	}
	t.Log(ha.String())

	ha = confhttp.HostAlias{
		IP:        net.ParseIP("2001:db8::1"),
		Hostnames: []string{"testV6_1", "testV6_2"},
	}
	t.Log(ha.String())
}
