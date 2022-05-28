package bgp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terassyi/grp/grp/log"
)

func TestGbpRegisterPeer(t *testing.T) {
	t.Skip()
	b, err := New(PORT, int(log.NoLog), "")
	require.NoError(t, err)
	addr := net.ParseIP("10.0.0.1")
	as := 0
	require.NoError(t, err)
	expPeer := newPeer(b.logger, nil, net.ParseIP("10.0.0.2"), addr, b.routerId, b.as, as)
	_, err = b.registerPeer(addr, b.routerId, b.as, as, false)
	require.NoError(t, err)
	p, ok := b.peers[addr.String()]
	assert.NotEqual(t, false, ok)
	assert.Equal(t, expPeer.state, p.state)
	// error already register
	_, err = b.registerPeer(addr, b.routerId, b.as, as, false)
	require.Error(t, ErrPeerAlreadyRegistered, err)
	// force register
	b.peers[addr.String()] = &peer{state: ACTIVE, as: as}
	_, err = b.registerPeer(addr, b.routerId, b.as, as, true)
	require.NoError(t, err)
	p2, ok := b.peers[addr.String()]
	assert.NotEqual(t, false, ok)
	assert.Equal(t, IDLE, p2.state)

}

func TestBgpRequestHandle(t *testing.T) {

}
