package bgp

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terassyi/grp/pkg/log"
)

func TestPeerHandleEvent(t *testing.T) {
	logger, err := log.New(log.NoLog, "")
	r := &LocRib{}
	require.NoError(t, err)
	p := newPeer(logger, nil, net.ParseIP("10.10.0.1"), net.ParseIP("10.0.0.2"), net.ParseIP("1.1.1.1"), 100, 200, r, &AdjRib{})
	for _, d := range []struct {
		evt event
	}{
		{evt: &bgpStart{}},
		{evt: &bgpStop{}},
	} {
		err := p.handleEvent(context.Background(), d.evt)
		require.NoError(t, err)
	}
}

func TestPeerChangeState(t *testing.T) {
	logger, _ := log.New(log.NoLog, "")
	neig := &neighbor{
		addr: net.ParseIP("10.0.0.1"),
		port: PORT,
		as:   100,
	}
	tests := []struct {
		name   string
		peer   *peer
		events []eventType
		result state
	}{
		{
			name:   "IDLE to ESTAB 1",
			peer:   &peer{peerInfo: &peerInfo{neighbor: neig}, state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start, event_type_bgp_trans_conn_open, event_type_recv_open_msg, event_type_recv_keepalive_msg},
			result: ESTABLISHED,
		},
		{
			name: "IDLE to ESTAB 2",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open_failed,
				event_type_bgp_trans_conn_open,
				event_type_recv_open_msg,
				event_type_recv_keepalive_msg},
			result: ESTABLISHED,
		},
		{
			name: "IDLE to ESTAB 3",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open_failed,
				event_type_bgp_start,
				event_type_bgp_start,
				event_type_bgp_start,
				event_type_bgp_trans_conn_open,
				event_type_recv_open_msg,
				event_type_recv_keepalive_msg},
			result: ESTABLISHED,
		},
		{
			name:   "IDLE to IDLE 1",
			peer:   &peer{peerInfo: &peerInfo{neighbor: neig}, state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start, event_type_bgp_stop},
			result: IDLE,
		},
		{
			name: "IDLE to IDLE 2",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open,
				event_type_recv_keepalive_msg,
			},
			result: IDLE,
		},
		{
			name: "IDLE to ESTAB 4",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open_failed,
				event_type_bgp_trans_conn_open,
				event_type_recv_open_msg,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_keepalive_timer_expired,
				event_type_keepalive_timer_expired,
				event_type_keepalive_timer_expired,
			},
			result: ESTABLISHED,
		},
		{
			name: "IDLE to ESTAB 5",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open_failed,
				event_type_bgp_trans_conn_open,
				event_type_recv_open_msg,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_recv_keepalive_msg,
			},
			result: ESTABLISHED,
		},
		{
			name: "IDLE to IDLE 3",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open_failed,
				event_type_bgp_trans_conn_open,
				event_type_recv_open_msg,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_recv_keepalive_msg,
				event_type_keepalive_timer_expired,
				event_type_recv_keepalive_msg,
				event_type_recv_notification_msg,
			},
			result: IDLE,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for _, e := range tt.events {
				err := tt.peer.changeState(e)
				require.NoError(t, err)
			}
			assert.Equal(t, tt.result, tt.peer.state)
		})
	}
}
