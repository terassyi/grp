package bgp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terassyi/grp/pkg/log"
)

func TestBgpRegisterPeer(t *testing.T) {
	b, err := New(PORT, int(log.NoLog), "")
	require.NoError(t, err)
	peerAddrs := make([]string, 0, 2)
	for _, tt := range []struct {
		peerAddr net.IP
		routerId net.IP
		myAS     int
		peerAS   int
		force    bool
	}{
		{peerAddr: net.ParseIP("10.0.0.3"), routerId: net.ParseIP("10.0.0.2"), myAS: 100, peerAS: 200, force: false},
		{peerAddr: net.ParseIP("10.0.1.2"), routerId: net.ParseIP("10.0.0.2"), myAS: 100, peerAS: 300, force: false},
	} {
		_, err := b.registerPeer(tt.peerAddr, tt.routerId, tt.myAS, tt.peerAS, tt.force)
		require.NoError(t, err)
		peerAddrs = append(peerAddrs, tt.peerAddr.String())
	}
	for k, v := range b.peers {
		assert.Contains(t, peerAddrs, k)
		assert.Contains(t, peerAddrs, v.neighbor.addr.String())
	}
}

func TestBgpRequestHandle(t *testing.T) {

}
