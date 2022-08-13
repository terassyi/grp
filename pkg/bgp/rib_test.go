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
	loc := NewLocRib()
	eth0, err := netlink.LinkByName("eth0")
	require.NoError(t, err)
	eth1, err := netlink.LinkByName("eth1")
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    *Path
		wantErr bool
	}{
		{name: "VALID 1", path: &Path{link: eth0, nlri: PrefixFromString("10.0.0.0/24"), nextHop: net.ParseIP("10.0.0.2")}, wantErr: false},
		{name: "VALID 2", path: &Path{link: eth1, nlri: PrefixFromString("10.0.1.0/24"), nextHop: net.ParseIP("10.0.1.3")}, wantErr: false},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := loc.Insert(tt.path)
			require.NoError(t, err)
		})
	}

}
func TestLocRib_IsReachable(t *testing.T) {
	loc := NewLocRib()

	nlri1 := PrefixFromString("10.0.0.0/24")
	nlri2 := PrefixFromString("10.1.0.0/24")
	nlri3 := PrefixFromString("10.2.0.0/24")
	nlri4 := PrefixFromString("10.3.0.0/24")
	nlri5 := PrefixFromString("10.4.0.0/24")

	loc.table[nlri1.String()] = &Path{nlri: nlri1}
	loc.table[nlri2.String()] = &Path{nlri: nlri2}
	loc.table[nlri3.String()] = &Path{nlri: nlri3}
	loc.table[nlri4.String()] = &Path{nlri: nlri4}
	loc.table[nlri5.String()] = &Path{nlri: nlri5}

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

func TestLocRib_GetByGroup(t *testing.T) {
	loc := NewLocRib()
	loc.table["10.0.0.0/24"] = &Path{id: 1, group: 0}
	loc.table["10.1.0.0/24"] = &Path{id: 2, group: 1}
	loc.table["10.1.1.0/24"] = &Path{id: 3, group: 1}
	loc.table["10.1.20.0/24"] = &Path{id: 4, group: 1}
	loc.table["10.2.1.0/24"] = &Path{id: 5, group: 2}
	loc.table["10.2.10.0/24"] = &Path{id: 6, group: 2}
	loc.table["10.20.0.0/24"] = &Path{id: 7, group: 3}
	wantMap := map[int][]int{
		0: {1},
		1: {2, 3, 4},
		2: {5, 6},
		3: {7},
	}
	t.Run("Group", func(t *testing.T) {
		pathes := loc.GetByGroup()
		pathIds := map[int][]int{
			0: {},
			1: {},
			2: {},
			3: {},
		}
		for g, pathes := range pathes {
			for _, p := range pathes {
				pathIds[g] = append(pathIds[g], p.id)
			}
		}
		for g, _ := range wantMap {
			assert.Equal(t, len(wantMap[g]), len(pathIds[g]))
		}
	})
}

func TestAdjRibIn_Insert(t *testing.T) {
	r := &AdjRibIn{mutex: &sync.RWMutex{}, table: make(map[string]map[int]*Path)}
	tests := []struct {
		name     string
		path     *Path
		inserted bool
	}{
		{name: "10.0.0.0/24 1", path: &Path{id: 1, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}}, inserted: true},
		{name: "10.0.0.0/24 2", path: &Path{id: 2, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}}, inserted: true},
		{name: "10.2.0.0/24 1", path: &Path{id: 3, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}}, inserted: true},
		{name: "192.168.0.0/24 1", path: &Path{id: 4, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("192.168.0.0")}}, inserted: true},
	}
	// no parallel
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Insert(tt.path)
			require.NoError(t, err)
			_, ok := r.table[tt.path.nlri.String()][tt.path.id]
			assert.Equal(t, tt.inserted, ok)
		})
	}
}

func TestAdjRibIn_Lookup(t *testing.T) {
	r := &AdjRibIn{mutex: &sync.RWMutex{}, table: make(map[string]map[int]*Path)}
	r.Insert(&Path{id: 1, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}})
	r.Insert(&Path{id: 2, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}})
	r.Insert(&Path{id: 3, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}})
	r.Insert(&Path{id: 4, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")}})
	r.Insert(&Path{id: 5, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}})
	r.Insert(&Path{id: 6, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}})
	r.Insert(&Path{id: 7, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")}})
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
			if tt.pathLen == 0 {
				if p != nil {
					t.Fatal("Adj-Rib-In: expected some value")
				}
			}
			assert.Equal(t, tt.pathLen, len(p))
		})
	}
}

func TestAdjRibIn_Drop(t *testing.T) {
	r := &AdjRibIn{mutex: &sync.RWMutex{}, table: make(map[string]map[int]*Path)}
	r.Insert(&Path{id: 1, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, nextHop: net.ParseIP("10.0.0.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 300}}}}})
	r.Insert(&Path{id: 2, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, nextHop: net.ParseIP("10.0.1.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 400, 500}}}}})
	r.Insert(&Path{id: 3, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, nextHop: net.ParseIP("10.0.5.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 400, 500}}}}})
	r.Insert(&Path{id: 4, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")}, nextHop: net.ParseIP("10.0.2.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 400}}}}})
	r.Insert(&Path{id: 5, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}, nextHop: net.ParseIP("10.0.0.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 400}}}}})
	r.Insert(&Path{id: 6, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.9.0.0")}, nextHop: net.ParseIP("10.0.0.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 400}}}}})
	r.Insert(&Path{id: 7, nlri: &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")}, nextHop: net.ParseIP("10.0.2.1"), asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 300, 400}}}}})
	tests := []struct {
		name       string
		id         int
		prefix     *Prefix
		next       net.IP
		asSequence []uint16
		pathLen    int
	}{
		{name: "DROP 1", id: 0, prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, next: net.ParseIP("10.0.0.1"), asSequence: []uint16{100, 200, 300}, pathLen: 2},
		{name: "DROP 2", id: 0, prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")}, next: net.ParseIP("10.0.2.1"), asSequence: []uint16{100, 200, 400}, pathLen: 0},
		{name: "DROP 3", id: 0, prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")}, next: net.ParseIP("10.0.0.1"), asSequence: []uint16{100, 200, 400}, pathLen: 0},
		{name: "DROP 4", id: 0, prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.9.0.0")}, next: net.ParseIP("10.0.0.1"), asSequence: []uint16{100, 200, 400}, pathLen: 0},
		{name: "DROP 5", id: 0, prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, next: net.ParseIP("10.0.5.1"), asSequence: []uint16{100, 200, 400}, pathLen: 2},
		{name: "DROP 6", id: 0, prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.0.0.0")}, next: net.ParseIP("10.0.5.1"), asSequence: []uint16{100, 400, 500}, pathLen: 1},
		{name: "DROP 7", id: 7, prefix: &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")}, pathLen: 0},
	}
	// noparallel
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := r.Drop(tt.prefix, tt.id, tt.next, tt.asSequence)
			require.NoError(t, err)
			p := r.Lookup(tt.prefix)
			assert.Equal(t, tt.pathLen, len(p))
		})
	}

}

func TestAdjRibOut_Lookup(t *testing.T) {
	r := &AdjRibOut{mutex: &sync.RWMutex{}, table: make(map[string]*Path)}
	r.Insert(&Path{
		as:             100,
		nextHop:        net.ParseIP("10.0.0.1"),
		nlri:           &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")},
		pathAttributes: []PathAttr{},
	})
	r.Insert(&Path{
		as:             200,
		nextHop:        net.ParseIP("10.1.0.1"),
		nlri:           &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")},
		pathAttributes: []PathAttr{},
	})
	r.Insert(&Path{
		as:             300,
		nextHop:        net.ParseIP("10.2.0.1"),
		nlri:           &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")},
		pathAttributes: []PathAttr{},
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
		as:             100,
		nextHop:        net.ParseIP("10.0.0.1"),
		nlri:           &Prefix{Length: 24, Prefix: net.ParseIP("10.1.0.0")},
		pathAttributes: []PathAttr{},
	})
	r.Insert(&Path{
		as:             200,
		nextHop:        net.ParseIP("10.1.0.1"),
		nlri:           &Prefix{Length: 24, Prefix: net.ParseIP("10.2.0.0")},
		pathAttributes: []PathAttr{},
	})
	r.Insert(&Path{
		as:             300,
		nextHop:        net.ParseIP("10.2.0.1"),
		nlri:           &Prefix{Length: 24, Prefix: net.ParseIP("10.3.0.0")},
		pathAttributes: []PathAttr{},
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
	r := NewLocRib()
	ar := newAdjRibIn()
	eth0, err := netlink.LinkByName("eth0")
	require.NoError(t, err)
	p := newPeer(logger, eth0, net.ParseIP("10.0.0.2"), net.ParseIP("10.0.0.3"), net.ParseIP("1.1.1.1"), 100, 200, r, ar, make(chan Packet, 0))

	tests := []struct {
		name            string
		path            *Path
		adjRibInPathLen int
		withdraw        bool
		expectedBest    bool
		expectedReason  BestPathSelectionReason
		wantErr         bool
	}{
		{
			name: "10.0.2.0/24 10.0.0.3 [200]",
			path: &Path{
				id:      1,
				link:    eth0,
				nlri:    PrefixFromString("10.0.2.0/24"),
				nextHop: net.ParseIP("10.0.0.3"),
				asPath:  ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200}}}},
			},
			adjRibInPathLen: 1,
			expectedBest:    true,
			expectedReason:  REASON_ONLY_PATH,
			wantErr:         false,
		},
		{
			name: "10.0.2.0/24 10.0.0.3 [200, 300]",
			path: &Path{
				id:      2,
				link:    eth0,
				nlri:    PrefixFromString("10.0.2.0/24"),
				nextHop: net.ParseIP("10.0.0.3"),
				asPath:  ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300}}}},
			},
			adjRibInPathLen: 2,
			expectedBest:    false,
			expectedReason:  REASON_AS_PATH_ATTR,
			wantErr:         false,
		},
		{
			name: "10.0.2.0/24 10.0.0.3 [200, 300, 200] Err ASLoop",
			path: &Path{
				id:      3,
				link:    eth0,
				nlri:    PrefixFromString("10.0.2.0/24"),
				nextHop: net.ParseIP("10.0.0.3"),
				asPath:  ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 200}}}},
			},
			adjRibInPathLen: 2,
			expectedBest:    false,
			wantErr:         true,
		},
		{
			name: "10.0.2.0/24 10.0.0.3 [] Local",
			path: &Path{
				id:   3,
				link: eth0,
				nlri: PrefixFromString("10.0.2.0/24"),
				// nextHop: net.ParseIP("10.0.0.3"),
				asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{}}}},
				local:  true,
			},
			adjRibInPathLen: 3,
			expectedBest:    true,
			expectedReason:  REASON_LOCAL_ORIGINATED,
			wantErr:         false,
		},
		{
			name: "10.0.2.0/24 10.0.0.3 [200, 400]",
			path: &Path{
				id:        4,
				link:      eth0,
				nlri:      PrefixFromString("10.0.2.0/24"),
				nextHop:   net.ParseIP("10.0.0.3"),
				asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 400}}}},
				localPref: 110,
			},
			adjRibInPathLen: 4,
			expectedBest:    true,
			expectedReason:  REASON_LOCAL_PREF_ATTR,
			wantErr:         false,
		},
	}
	t.Cleanup(func() {
		_, delRoute1, _ := net.ParseCIDR("10.0.2.0/24")
		if err := netlink.RouteDel(&netlink.Route{
			LinkIndex: eth0.Attrs().Index,
			Dst:       delRoute1,
		}); err != nil {
			t.Log(err)
		}
		t.Log("Clean up")
	})
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
			tt := tt
			t.Log(tt.path.nlri.Network())
			_, _, err := p.rib.In.Select(p.as, tt.path, tt.withdraw, p.bestPathConfig)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBest, tt.path.best)
				if tt.expectedBest {
					assert.Equal(t, tt.expectedReason, tt.path.reason)
				}
			}
			t.Logf("Adj-Rib-In[%s]", tt.path.nlri)
			for _, path := range p.rib.In.table[tt.path.nlri.String()] {
				t.Log(path)
			}
			// assert.Equal(t, tt.adjRibInPathLen, len(p.rib.In.table[tt.path.nlri.String()]))
		})
		cancel()
	}
}
