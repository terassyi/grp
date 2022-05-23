package bgp

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terassyi/grp/grp/log"
)

func TestPeerHandleEvent(t *testing.T) {
	logger, err := log.New(log.NoLog, "")
	require.NoError(t, err)
	p := newPeer(logger, nil, net.ParseIP("10.10.0.1"), net.ParseIP("10.0.0.2"), 1)
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
	for _, d := range []struct {
		peer   *peer
		events []eventType
		result state
	}{
		{
			peer:   &peer{state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start, event_type_bgp_trans_conn_open, event_type_recv_open_msg, event_type_recv_keepalive_msg},
			result: ESTABLISHED,
		},
		{
			peer: &peer{state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open_failed,
				event_type_bgp_trans_conn_open,
				event_type_recv_open_msg,
				event_type_recv_keepalive_msg},
			result: ESTABLISHED,
		},
		{
			peer: &peer{state: IDLE, logger: logger},
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
			peer:   &peer{state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start, event_type_bgp_stop},
			result: IDLE,
		},
		{
			peer: &peer{state: IDLE, logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open,
				event_type_recv_keepalive_msg,
			},
			result: IDLE,
		},
		{
			peer: &peer{state: IDLE, logger: logger},
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
			peer: &peer{state: IDLE, logger: logger},
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
			peer: &peer{state: IDLE, logger: logger},
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
	} {
		for _, e := range d.events {
			err := d.peer.changeState(e)
			require.NoError(t, err)
		}
		assert.Equal(t, d.result, d.peer.state)
	}
}
