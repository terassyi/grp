package bgp

import (
	"fmt"
	"net"
	"reflect"
	"sync"

	"github.com/terassyi/grp/pkg/rib"
	"github.com/vishvananda/netlink"
)

// Routes: Advertisement and Storage
// For the purpose of this protocol, a route is defined as a unit of information
// that pairs a set of destinations with the attributes of a ptah to those destinations.
// The set of destinations are systems whose IP addresses are contained in one IP address prefix
// that is carried in the Network Layer Reachability Information(NLRI) field of an UPDATE message,
// and the path is the information reported in the path attributes field of the same UPDATE message.
//
// Routes are advertised between BGP speakers in UPDATE messages.
// Multiple routes that have the same path attributes can be advertised in a single UPDATE message
// by including multiple prefixes in the NLRI field of the UPDATE message.
//
// Routes are stored in the Routing Information Bases(RIBs):
// namely, the Adj-RIB-In, the Loc-RIB, and Adj-RIB-Out.

// Routing Information Base
type AdjRibIn struct {
	mutex *sync.RWMutex
	table map[string]map[int]*Path
}

func newAdjRibIn() *AdjRibIn {
	return &AdjRibIn{
		mutex: &sync.RWMutex{},
		table: make(map[string]map[int]*Path),
	}
}

func (r *AdjRibIn) Insert(path *Path) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	path.status = PathStatusInstalledIntoAdjRibIn
	_, ok := r.table[path.nlri.String()]
	if !ok {
		r.table[path.nlri.String()] = map[int]*Path{path.id: path}
	} else {
		// r.table[path.nlri.String()] = append(r.table[path.nlri.String()], path)
		r.table[path.nlri.String()][path.id] = path
	}
	return nil
}

func (r *AdjRibIn) Lookup(prefix *Prefix) []*Path {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	pathes, ok := r.table[prefix.String()]
	if !ok {
		return nil
	}
	res := make([]*Path, 0)
	for _, path := range pathes {
		res = append(res, path)
	}
	return res
}

func (r *AdjRibIn) LookupById(id int) *Path {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, network := range r.table {
		path, ok := network[id]
		if ok {
			return path
		}
	}
	return nil
}

func (r *AdjRibIn) Drop(prefix *Prefix, id int, next net.IP, asSequence []uint16) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for k, v := range r.table {
		if k == prefix.String() {
			if id > 0 {
				delete(r.table[prefix.String()], id)
				return nil
			}
			for i, vv := range v {
				var ases []uint16
				for _, a := range vv.asPath.Segments {
					if a.Type == SEG_TYPE_AS_SEQUENCE {
						ases = a.AS2
					}
				}
				if vv.nextHop.Equal(next) && reflect.DeepEqual(ases, asSequence) {
					delete(r.table[prefix.String()], i)
				}
			}
		}
	}
	return nil
}

type AdjRibOut struct {
	mutex *sync.RWMutex
	table map[string]*Path
}

func (r *AdjRibOut) Insert(path *Path) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	path.status = PathStatusInstalledIntoAdjRibOut
	r.table[path.nlri.String()] = path
	return nil
}

func (r *AdjRibOut) Lookup(prefix *Prefix) *Path {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	p, ok := r.table[prefix.String()]
	if !ok {
		return nil
	}
	return p
}

func (r *AdjRibOut) Drop(prefix *Prefix) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.table, prefix.String())
	return nil
}

func (r *AdjRibOut) Sync(path *Path) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.table[path.nlri.String()]
	if !ok {
	}
	return nil
}

type AdjRib struct {
	// Adj-RIB-In: The Adj-RIB-In store routing information that has been learned from inbound UPDATE messages.
	// 	           Their contents represent routes that are available as an input to the Decision Process.
	In *AdjRibIn
	// Adj-RIB-Out: The Adj-RIB-Out store the information that the local routing information that the BGP speaker has selected
	//              for advertisement to its peers.
	//              The routing information stored in the Adj-RIB-Out will be carried in the local BGP speaker's UPDATE messages and advertised to its peers.
	Out *AdjRibOut
}

func newAdjRib() (*AdjRib, error) {
	return &AdjRib{
		In:  &AdjRibIn{mutex: &sync.RWMutex{}, table: make(map[string]map[int]*Path)},
		Out: &AdjRibOut{mutex: &sync.RWMutex{}, table: make(map[string]*Path)},
	}, nil
}

// Loc-RIB: The Loc-RIB contains the local routing information that the BGP speaker has selected by applying its local policies
//          to the routing information contained in its Adj-RIB-In.
type LocRib struct {
	mutex *sync.RWMutex
	table map[string]*Path
	queue chan []*Path
}

func NewLocRib() *LocRib {
	return &LocRib{
		mutex: &sync.RWMutex{},
		table: make(map[string]*Path),
		queue: make(chan []*Path, 128),
	}
}

func (l *LocRib) enqueue(pathes []*Path) {
	l.queue <- pathes
}

func setupLocRib(family int) ([]netlink.Route, error) {
	routes := make([]netlink.Route, 0)
	interfaces, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("setupLocRib: failed to get interfaces: %w", err)
	}
	for _, iface := range interfaces {
		rs, err := rib.LookUp4(iface)
		if err != nil {
			return nil, fmt.Errorf("setupLocRib: failed to get route in %s: %w", iface.Attrs().Name, err)
		}
		routes = append(routes, rs...)
	}
	return routes, nil
}

func (l *LocRib) Insert(path *Path) error {
	l.mutex.Lock()
	l.table[path.nlri.String()] = path
	l.mutex.Unlock()
	routes, err := l.isntallToRib(path.link, path.nlri.Network(), path.nextHop)
	if err != nil {
		return fmt.Errorf("LocRib_InsertPath: %w", err)
	}
	path.routes = routes
	path.status = PathStatusInstalledIntoLocRib
	return nil
}

func (l *LocRib) isntallToRib(link netlink.Link, cidr *net.IPNet, next net.IP) ([]netlink.Route, error) {
	routes, err := rib.Get4(link, cidr.IP)
	if err != nil {
		return nil, fmt.Errorf("LocRib_installToRib: %w", err)
	}
	if len(routes) == 0 {
		if err := rib.Add4(link, cidr, next, rib.RT_PROTO_BGP); err != nil {
			return nil, fmt.Errorf("LocRib_installToRib: %w", err)
		}
	} else {
		if err := rib.Replace4(link, cidr, next, rib.RT_PROTO_BGP); err != nil {
			return nil, fmt.Errorf("LocRib_installToRib: %w", err)
		}
	}
	return rib.Get4(link, cidr.IP)
}

func (l *LocRib) IsReachable(addr net.IP) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	for _, v := range l.table {
		res := v.nlri.Network().Contains(addr)
		if res {
			return true
		}
	}
	return false
}

func (l *LocRib) GetNotSyncedPath() ([]*Path, error) {
	l.mutex.Lock()
	notSynced := make([]*Path, 0)
	defer l.mutex.Unlock()
	for _, path := range l.table {
		if path.status == PathStatusNotSynchronized || path.status == PathStatusInstalledIntoLocRib {
			notSynced = append(notSynced, path)
		}
	}
	return notSynced, nil
}
