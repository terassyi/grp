package bgp

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terassyi/grp/pkg/log"
	"github.com/vishvananda/netlink"
)

func TestLocRib_Insert(t *testing.T) {
	loc, err := NewLocRib()
	require.NoError(t, err)
	tests := []struct {
		name    string
		network string
		expErr  error
	}{
		{name: "VALID 10.1.0.0/24", network: "10.0.1.0/24", expErr: nil},
		{name: "VALID 10.1.2.0/24", network: "10.0.1.2/24", expErr: nil},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := loc.Insert(tt.network)
			if tt.expErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestLocRib_InsertPath(t *testing.T) {
	loc, err := NewLocRib()
	require.NoError(t, err)
	eth0, err := netlink.LinkByName("eth0")
	require.NoError(t, err)
	eth1, err := netlink.LinkByName("eth1")
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    *Path
		wantErr bool
	}{
		{name: "VALID 1", path: &Path{link: eth0, nlri: PrefixFromString("10.0.0.0/24"), nextHop: net.ParseIP("10.0.0.2"), picked: true}, wantErr: false},
		{name: "VALID 2", path: &Path{link: eth1, nlri: PrefixFromString("10.0.1.0/24"), nextHop: net.ParseIP("10.0.1.3"), picked: true}, wantErr: false},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := loc.InsertPath(tt.path)
			require.NoError(t, err)
		})
	}

}
func TestLocRib_IsReachable(t *testing.T) {
	loc, err := NewLocRib()
	require.NoError(t, err)
	loc.Insert("10.0.0.0/24")
	loc.Insert("10.1.0.0/24")
	loc.Insert("10.2.0.0/24")
	loc.Insert("10.3.0.0/24")
	loc.Insert("10.4.0.0/24")

	tests := []struct {
		name string
		addr net.IP
		res  bool
	}{
		{name: "VALID 1", addr: net.ParseIP("10.0.0.100"), res: true},
		{name: "VALID 2", addr: net.ParseIP("10.0.0.200"), res: true},
		{name: "VALID 3", addr: net.ParseIP("10.1.0.100"), res: true},
		{name: "VALID 4", addr: net.ParseIP("10.3.0.100"), res: true},
		{name: "INVALID 1", addr: net.ParseIP("10.100.0.100"), res: false},
		{name: "INVALID 2", addr: net.ParseIP("10.30.0.100"), res: false},
		{name: "INVALID 3", addr: net.ParseIP("10.233.0.100"), res: false},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res := loc.IsReachable(tt.addr)
			assert.Equal(t, tt.res, res)
		})
	}
}

func TestAdjRibIn_Insert(t *testing.T) {
	r := &AdjRibIn{mutex: &sync.RWMutex{}, table: make(map[string][]*Path)}
	tests := []struct {
		name                string
		path                *Path
		expTableEntryLength int
	}{
		{name: "10.0.0.0/24 1", path: &Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}}, expTableEntryLength: 1},
		{name: "10.0.0.0/24 2", path: &Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}}, expTableEntryLength: 2},
		{name: "10.2.0.0/24 1", path: &Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}}, expTableEntryLength: 1},
		{name: "192.168.0.0/24 1", path: &Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("192.168.0.0")}}, expTableEntryLength: 1},
	}
	// no parallel
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Insert(tt.path)
			require.NoError(t, err)
		})
		assert.Equal(t, tt.expTableEntryLength, len(r.table[tt.path.nlri.String()]))
	}
}

func TestAdjRibIn_Lookup(t *testing.T) {
	r := &AdjRibIn{mutex: &sync.RWMutex{}, table: make(map[string][]*Path)}
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")}})
	tests := []struct {
		name    string
		prefix  *Prefix
		pathLen int
	}{
		{name: "VALID 1", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, pathLen: 3},
		{name: "VALID 2", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")}, pathLen: 1},
		{name: "VALID 3", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}, pathLen: 2},
		{name: "VALID 4", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.9.0.0")}, pathLen: 0},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := r.Lookup(tt.prefix)
			assert.Equal(t, tt.pathLen, len(p))
		})
	}
}

func TestAdjRibIn_Drop(t *testing.T) {
	r := &AdjRibIn{mutex: &sync.RWMutex{}, table: make(map[string][]*Path)}
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, nextHop: net.ParseIP("10.0.0.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 300}}}}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, nextHop: net.ParseIP("10.0.1.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 400, 500}}}}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, nextHop: net.ParseIP("10.0.5.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 400, 500}}}}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")}, nextHop: net.ParseIP("10.0.2.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 400}}}}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}, nextHop: net.ParseIP("10.0.0.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 400}}}}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.9.0.0")}, nextHop: net.ParseIP("10.0.0.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 400}}}}})
	r.Insert(&Path{nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")}, nextHop: net.ParseIP("10.0.2.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 300, 400}}}}})
	tests := []struct {
		name       string
		prefix     *Prefix
		next       net.IP
		asSequence []uint16
		pathLen    int
	}{
		{name: "DROP 1", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, next: net.ParseIP("10.0.0.1"), asSequence: []uint16{100, 200, 300}, pathLen: 2},
		{name: "DROP 2", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")}, next: net.ParseIP("10.0.2.1"), asSequence: []uint16{100, 200, 400}, pathLen: 0},
		{name: "DROP 3", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}, next: net.ParseIP("10.0.0.1"), asSequence: []uint16{100, 200, 400}, pathLen: 0},
		{name: "DROP 4", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.9.0.0")}, next: net.ParseIP("10.0.0.1"), asSequence: []uint16{100, 200, 400}, pathLen: 0},
		{name: "DROP 5", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, next: net.ParseIP("10.0.5.1"), asSequence: []uint16{100, 200, 400}, pathLen: 2},
		{name: "DROP 6", prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, next: net.ParseIP("10.0.5.1"), asSequence: []uint16{100, 400, 500}, pathLen: 1},
	}
	// noparallel
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := r.Drop(tt.prefix, tt.next, tt.asSequence)
			require.NoError(t, err)
			p := r.Lookup(tt.prefix)
			assert.Equal(t, tt.pathLen, len(p))
		})
	}

}

func TestAdjRibOut_Lookup(t *testing.T) {
	r := &AdjRibOut{mutex: &sync.RWMutex{}, table: make(map[string]*Path)}
	r.Insert(&Path{
		as:              100,
		nextHop:         net.ParseIP("10.0.0.1"),
		nlri:            &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")},
		recognizedAttrs: []PathAttr{},
	})
	r.Insert(&Path{
		as:              200,
		nextHop:         net.ParseIP("10.1.0.1"),
		nlri:            &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")},
		recognizedAttrs: []PathAttr{},
	})
	r.Insert(&Path{
		as:              300,
		nextHop:         net.ParseIP("10.2.0.1"),
		nlri:            &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")},
		recognizedAttrs: []PathAttr{},
	})
	tests := []struct {
		name   string
		prefix *Prefix
		exist  bool
	}{
		{
			name:   "EXIST 1",
			prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")},
			exist:  true,
		},
		{
			name:   "EXIST 2",
			prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")},
			exist:  true,
		},
		{
			name:   "Not EXIST 1",
			prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.4.0.0")},
			exist:  false,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := r.Lookup(tt.prefix)
			if tt.exist && p == nil {
				t.Log(tt.prefix.String())
				t.Fatal("Path must be exist.")
			} else if !tt.exist && p != nil {
				t.Fatal("Path must not be exist.")
			}
		})
	}
}

func TestAdjRibOut_Drop(t *testing.T) {
	r := &AdjRibOut{mutex: &sync.RWMutex{}, table: make(map[string]*Path)}
	r.Insert(&Path{
		as:              100,
		nextHop:         net.ParseIP("10.0.0.1"),
		nlri:            &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")},
		recognizedAttrs: []PathAttr{},
	})
	r.Insert(&Path{
		as:              200,
		nextHop:         net.ParseIP("10.1.0.1"),
		nlri:            &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")},
		recognizedAttrs: []PathAttr{},
	})
	r.Insert(&Path{
		as:              300,
		nextHop:         net.ParseIP("10.2.0.1"),
		nlri:            &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")},
		recognizedAttrs: []PathAttr{},
	})
	tests := []struct {
		name   string
		prefix *Prefix
		exist  bool
	}{
		{
			name:   "DELETE 1",
			prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")},
			exist:  true,
		},
		{
			name:   "DELETE 2",
			prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")},
			exist:  true,
		},
		{
			name:   "Not EXIST 1",
			prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.4.0.0")},
			exist:  false,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Drop(tt.prefix)
			require.NoError(t, err)
			p := r.Lookup(tt.prefix)
			assert.Nil(t, p)
		})
	}
}

func TestPeer_Select(t *testing.T) {
	logger, err := log.New(log.NoLog, "")
	require.NoError(t, err)
	r, _ := NewLocRib()
	eth0, err := netlink.LinkByName("eth0")
	require.NoError(t, err)
	p := newPeer(logger, eth0, net.ParseIP("10.0.0.2"), net.ParseIP("10.0.0.3"), net.ParseIP("1.1.1.1"), 100, 200, r)
	p.rib.In.Insert(&Path{
		link:    eth0,
		nlri:    &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")},
		nextHop: net.ParseIP("10.0.0.3"),
		asPath: ASPath{
			Segments: []*ASPathSegment{
				{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300}},
			},
		},
		preference: 100,
		picked:     false,
	})
	p.rib.In.Insert(&Path{
		link:    eth0,
		nlri:    &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")},
		nextHop: net.ParseIP("10.0.0.3"),
		asPath: ASPath{
			Segments: []*ASPathSegment{
				{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 400}},
			},
		},
		preference: 100,
		picked:     false,
	})
	p.rib.In.Insert(&Path{
		link:    eth0,
		nlri:    &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")},
		nextHop: net.ParseIP("10.0.0.3"),
		asPath: ASPath{
			Segments: []*ASPathSegment{
				{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 400}},
			},
		},
		preference: 50,
		picked:     false,
	})
	p.rib.In.Insert(&Path{
		link:    eth0,
		nlri:    &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")},
		nextHop: net.ParseIP("10.0.0.3"),
		asPath: ASPath{
			Segments: []*ASPathSegment{
				{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200}},
			},
		},
		preference: 100,
		picked:     true,
	})

	tests := []struct {
		name            string
		path            *Path
		adjRibInPathLen int
		expectPick      bool
		wantErr         bool
	}{
		{
			name: "10.0.2.0/24 10.0.0.3 [200]",
			path: &Path{
				link:       eth0,
				nlri:       PrefixFromString("10.0.2.0/24"),
				nextHop:    net.ParseIP("10.0.0.3"),
				asPath:     ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200}}}},
				preference: 100,
				picked:     false,
			},
			adjRibInPathLen: 1,
			expectPick:      true,
			wantErr:         false,
		},
		{
			name: "10.0.2.0/24 10.0.0.3 [200, 300]",
			path: &Path{
				link:       eth0,
				nlri:       PrefixFromString("10.0.2.0/24"),
				nextHop:    net.ParseIP("10.0.0.3"),
				asPath:     ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300}}}},
				preference: 50,
				picked:     false,
			},
			adjRibInPathLen: 2,
			expectPick:      false,
			wantErr:         false,
		},
		{
			name: "10.0.2.0/24 10.0.0.3 [200, 300, 200] ASLoop",
			path: &Path{
				link:       eth0,
				nlri:       PrefixFromString("10.0.2.0/24"),
				nextHop:    net.ParseIP("10.0.0.3"),
				asPath:     ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 200}}}},
				preference: 10,
				picked:     false,
			},
			adjRibInPathLen: 2,
			expectPick:      false,
			wantErr:         true,
		},
	}
	t.Cleanup(func() {
		_, delRoute1, _ := net.ParseCIDR("10.0.2.0/24")
		if err := netlink.RouteDel(&netlink.Route{
			LinkIndex: eth0.Attrs().Index,
			Dst:       delRoute1,
		}); err != nil {
			t.Fatal(err)
		}
	})
	t.Parallel()
	for _, tt := range tests {
		ctx, cancel := context.WithCancel(context.Background())
		tt := tt
		go func(ctx context.Context, t *testing.T) {
			for {
				select {
				case event := <-p.eventQueue:
					if event.typ() != event_type_trigger_decision_process {
						t.Logf("Unexpected event %s", event.typ().String())
						cancel()
					}
					t.Logf("event received %s", event)
				case <-ctx.Done():
					return
				}
			}

		}(ctx, t)
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.path.nlri.Network())
			err := p.Select(tt.path)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectPick, tt.path.picked)
				if tt.expectPick {
					pickedRoute, ok := p.rib.In.Picked(tt.path.nlri)
					require.Equal(t, true, ok)
					assert.Equal(t, tt.path, pickedRoute)
				} else {
					pickedRoute, ok := p.rib.In.Picked(tt.path.nlri)
					require.Equal(t, true, ok)
					assert.NotEqual(t, tt.path, pickedRoute)
				}
			}
			assert.Equal(t, tt.adjRibInPathLen, len(p.rib.In.table[tt.path.nlri.String()]))
		})
		cancel()
	}
}
