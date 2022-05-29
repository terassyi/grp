package bgp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitAddrAndPort(t *testing.T) {
	for _, d := range []struct {
		s    string
		addr net.IP
		port int
	}{
		{s: "[2001:db8::1]:80", addr: net.ParseIP("2001:db8::1"), port: 80},
		{s: "10.0.0.1:8080", addr: net.ParseIP("10.0.0.1"), port: 8080},
	} {
		a, p, err := SplitAddrAndPort(d.s)
		require.NoError(t, err)
		assert.Equal(t, d.addr, a)
		assert.Equal(t, d.port, p)
	}
}
