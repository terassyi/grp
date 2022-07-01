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
	p := newPeer(logger, nil, net.ParseIP("10.10.0.1"), net.ParseIP("10.0.0.2"), net.ParseIP("1.1.1.1"), 100, 200, r, &AdjRibIn{})
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
			peer:   &peer{peerInfo: &peerInfo{neighbor: neig}, fsm: newFSM(), logger: logger},
			events: []eventType{event_type_bgp_start, event_type_bgp_trans_conn_open, event_type_recv_open_msg, event_type_recv_keepalive_msg},
			result: ESTABLISHED,
		},
		{
			name: "IDLE to ESTAB 2",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, fsm: newFSM(), logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open_failed,
				event_type_bgp_trans_conn_open,
				event_type_recv_open_msg,
				event_type_recv_keepalive_msg},
			result: ESTABLISHED,
		},
		{
			name: "IDLE to ESTAB 3",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, fsm: newFSM(), logger: logger},
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
			peer:   &peer{peerInfo: &peerInfo{neighbor: neig}, fsm: newFSM(), logger: logger},
			events: []eventType{event_type_bgp_start, event_type_bgp_stop},
			result: IDLE,
		},
		{
			name: "IDLE to IDLE 2",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, fsm: newFSM(), logger: logger},
			events: []eventType{event_type_bgp_start,
				event_type_bgp_trans_conn_open,
				event_type_recv_keepalive_msg,
			},
			result: IDLE,
		},
		{
			name: "IDLE to ESTAB 4",
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, fsm: newFSM(), logger: logger},
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
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, fsm: newFSM(), logger: logger},
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
			peer: &peer{peerInfo: &peerInfo{neighbor: neig}, fsm: newFSM(), logger: logger},
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
				err := tt.peer.fsm.Change(e)
				require.NoError(t, err)
			}
			assert.Equal(t, tt.result, tt.peer.fsm.GetState())
		})
	}
}

func TestPeer_generateOutPath(t *testing.T) {
	logger, _ := log.New(log.NoLog, "")
	p := newPeer(logger, nil, net.ParseIP("10.0.0.2"), net.ParseIP("10.0.0.3"), net.ParseIP("1.1.1.1"), 100, 200, NewLocRib(), newAdjRibIn())
	tests := []struct {
		name         string
		incomingPath *Path
		outPath      *Path
		wantErr      bool
	}{
		{
			name: "LOCAL_PATH",
			incomingPath: &Path{
				id:    1,
				info:  nil,
				local: true,
			},
			outPath: &Path{
				id:    1,
				info:  p.peerInfo,
				local: true,
			},
			wantErr: false,
		},
		{
			name: "PEER_PATH",
			incomingPath: &Path{
				id:      2,
				info:    p.peerInfo,
				local:   false,
				origin:  *CreateOrigin(ORIGIN_IGP),
				as:      100,
				asPath:  *CreateASPath([]uint16{200}),
				nextHop: p.peerInfo.neighbor.addr,
				nlri:    PrefixFromString("10.0.1.0/24"),
			},
			outPath: &Path{
				id:      2,
				info:    p.peerInfo,
				local:   false,
				origin:  *CreateOrigin(ORIGIN_IGP),
				as:      100,
				asPath:  *CreateASPath([]uint16{100, 200}),
				nextHop: p.peerInfo.neighbor.addr,
				nlri:    PrefixFromString("10.0.1.0/24"),
			},
			wantErr: false,
		},
		{
			name: "NOT_PEER_PATH",
			incomingPath: &Path{
				id:      3,
				info:    p.peerInfo,
				local:   false,
				origin:  *CreateOrigin(ORIGIN_IGP),
				as:      100,
				asPath:  *CreateASPath([]uint16{200, 300}),
				nextHop: net.ParseIP("10.0.0.3"),
				nlri:    PrefixFromString("10.0.2.0/24"),
			},
			outPath: &Path{
				id:      3,
				info:    p.peerInfo,
				local:   false,
				origin:  *CreateOrigin(ORIGIN_IGP),
				as:      100,
				asPath:  *CreateASPath([]uint16{100, 200, 300}),
				nextHop: p.peerInfo.neighbor.addr,
				nlri:    PrefixFromString("10.0.2.0/24"),
			},
			wantErr: false,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			outPath, err := p.generateOutPath(tt.incomingPath)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.outPath, outPath)
			}
		})
	}
}

func TestPeer_buildUpdateMessage(t *testing.T) {
	logger, _ := log.New(log.NoLog, "")
	p := newPeer(logger, nil, net.ParseIP("10.0.0.2"), net.ParseIP("10.0.0.3"), net.ParseIP("1.1.1.1"), 100, 200, NewLocRib(), newAdjRibIn())
	tests := []struct {
		name   string
		pathes []*Path
		update *Update
	}{
		{
			name:   "CASE 1",
			pathes: []*Path{},
			update: &Update{
				WithdrawnRoutes:              []*Prefix{},
				PathAttrs:                    []PathAttr{},
				NetworkLayerReachabilityInfo: []*Prefix{},
			},
		},
		{
			name: "CASE 2",
			pathes: []*Path{
				{
					as:             100,
					asPath:         *CreateASPath([]uint16{100}),
					origin:         *CreateOrigin(ORIGIN_IGP),
					nextHop:        net.ParseIP("10.0.0.2"),
					med:            0,
					pathAttributes: []PathAttr{},
					nlri:           PrefixFromString("10.0.1.0/24"),
				},
			},
			update: &Update{
				WithdrawnRoutesLen: 0,
				WithdrawnRoutes:    []*Prefix{},
				TotalPathAttrLen:   25,
				PathAttrs: []PathAttr{
					CreateOrigin(ORIGIN_IGP),
					CreateASPath([]uint16{100}),
					CreateNextHop(net.ParseIP("10.0.0.2")),
					CreateMultiExitDisc(0),
				},
				NetworkLayerReachabilityInfo: []*Prefix{PrefixFromString("10.0.1.0/24")},
			},
		},
		{
			name: "CASE 2",
			pathes: []*Path{
				{
					as:             100,
					asPath:         *CreateASPath([]uint16{100, 200}),
					origin:         *CreateOrigin(ORIGIN_IGP),
					nextHop:        net.ParseIP("10.0.0.2"),
					med:            0,
					pathAttributes: []PathAttr{},
					nlri:           PrefixFromString("10.0.1.0/24"),
				},
				{
					as:             100,
					asPath:         *CreateASPath([]uint16{100, 200}),
					origin:         *CreateOrigin(ORIGIN_IGP),
					nextHop:        net.ParseIP("10.0.0.2"),
					med:            0,
					pathAttributes: []PathAttr{},
					nlri:           PrefixFromString("10.0.2.0/24"),
				},
			},
			update: &Update{
				WithdrawnRoutesLen: 0,
				WithdrawnRoutes:    []*Prefix{},
				TotalPathAttrLen:   27,
				PathAttrs: []PathAttr{
					CreateOrigin(ORIGIN_IGP),
					CreateASPath([]uint16{100, 200}),
					CreateNextHop(net.ParseIP("10.0.0.2")),
					CreateMultiExitDisc(0),
				},
				NetworkLayerReachabilityInfo: []*Prefix{PrefixFromString("10.0.1.0/24"), PrefixFromString("10.0.2.0/24")},
			},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			msg, err := p.buildUpdateMessage(tt.pathes)
			require.NoError(t, err)
			assert.Equal(t, tt.update, GetMessage[*Update](msg.Message))
		})
	}
}
