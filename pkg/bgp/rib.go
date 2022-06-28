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

// The Decision Process selects routes for subsequent advertisement by applying the policies in the local Policy Information Base(PIB) to the routes stored in its Adj-RIB-In.
// The output of the Decision Process is the set of routes that will be advertised to all peers;
// the selected routes will be stored in the local speaker's Adj-RIB-Out.
//
// The selection process is formalized by defining a function that takes the attribute of a given route as an argument
// and returns a non-negative integer denoting the degree of preference for the route.
// The function that calculates the degree of preference for a given route shall not use as its inputs any of the following:
// the existence of other routes, the non-existence of other routes,
// or the path attributes of other routes.
// Route selection then consists of individual application of the degree of preference function to each feasible route,
// followed by the choice of the one with the highest degree of preference.

// The Decision Process operates on routes contained in each Adj-RIB-In, and is responsible for:
// - selection of routes to be advertised to BGP speakers located in the local speaker's autonomous system
// - selection of routes to be advertised to BGP speakers located in neighboring autonomous systems
// - route aggregation and route information reduction

// Calculation of Degree of Preference
// This decision function is invoked whenever the local BGP speaker receives, from a peer,
// an UPDATE message that advertises a new route, a replacement route, or withdrawn routes.
// This decision function is a separate process, which completes when it has no further work to do.
// This decision function locks an Adj-RIB-In prior to operating on any route contained within it,
// and unlocks it after operating on all new or unfeasible routes contained within it.
// For each newly received or replacement feasible route, the local BGP speaker determines a degree of preference as follows:
//    If the route is learned an internal peer, either the value of the LOCAL_PREF attribute is taken an the degree of preference,
//    or the local system computes the degree of preference of the route based on preconfigured policy information.
//    Note that the latter may result information of persistent routing loops.
//
//    If the route is learned from an external peer, then the local BGP speaker computes the degree of preference based on preconfigured policy information.
//    If the return value indicates the route is ineligible, the route MAY NOT serve as an input to the next phase of route selection;
//    otherwise, the return value MUST be used as the LOCAL_PREF value in any IBGP readvertisement.
// TODO: implement bestpath selection
func (r *AdjRibIn) Calculate(nlri *Prefix, bestPathConfig *BestPathConfig) (BestPathSelectionReason, *Path, error) {
	pathes := r.Lookup(nlri)
	if pathes == nil {
		return REASON_INVALID, nil, fmt.Errorf("Peer_Calculate: path is not found for %s", nlri)
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	res, reason := sortPathes(pathes)
	if len(res) == 0 {
		return REASON_INVALID, nil, fmt.Errorf("Peer_Calculate: path is not found for %s", nlri)
	}
	// mark best
	res[0].best = true
	for i := 1; i < len(res); i++ {
		res[i].best = false
	}
	return reason, res[0], nil
}

// Route Selection
// This function is invoked on completion of Calculate().
// This function is a separate process, which completes when it has no further work to do.
// This process considers all routes that are eligible in the Adj-RIB-In.
// This function is blocked from running while the Phase 3 decision functions is in process.
// This locks all Adj-RIB-In prior to commencing its function, and unlocks then on completion.
// If the NEXT_HOP attribute of a BGP route depicts an address that is not resolvable,
// or if it would become unresolvable if the route was installed in the routing table, the BGP route MUST be excluded from this function.
// IF the AS_PATH attribute of a BGP route contians an AS loop, the BGP route should be excluded from this function.
// AS loop detection is done by scanning the full AS path(as specified in the AS_PATH attribute),
// and checking that the autonomous system number of the local system does not appear in the AS path.
// Operations of a BGP speaker that is configured to accept routes with its own autonomous system number in the AS path are outside the scope of this document.
// It is critical that BGP speakers within an AS do not make conflicting decisions regarding route selection that would cause forwarding loops to occur.
//
// For each set of destinations for which a feasible route exists in the Adj-RIB-In, the local BGP speaker identifies the route that has:
//   a) the highest degree of preference of any route to the same set of destinations, or
//   b) is the only route to that destination, or
//   c) is selected as a result of the Phase 2 tie breaking rules
//
// The local speaker SHALL then install that route in the Loc-RIB,
// replacing any route to the same destination that is currently being held in the Loc-RIB.
// When the new BGP route is installed in the Rouing Table,
// care must be taken to ensure that existing routes to the same destination that are now considered invalid are removed from the Routing Table.
// Wether the new BGP route replaces an existing non-BGP route in the Routing Table depends on the policy configured on the BGP speaker.
//
// The local speaker MUST determine the immediate next-hop address from the NEXT_HOP attribute of the selected route.
// If either the immediate next-hop or the IGP cost to the NEXT_HOP (wherer the NEXT_HOP is resolved throudh an IGP route) changes, Phase 2 Route selection MUST be performed again.
//
// Notice that even though BGP routes do not have to be installed in the Routing Table with the immediate next-hos(s),
// implementations MUST take care that, before any packets are forwarded along a BGP route,
// its associated NEXT_HOP address is resolved to the immediate (directly connected) next-hop address, and that this address (or multiple addresses) is finally used for actual packet forwarding.
func (r *AdjRibIn) Select(as int, path *Path, withdrawn bool, bestPathConfig *BestPathConfig) (*Path, error) {
	if !path.local && path.asPath.Contains(as) {
		return nil, nil
	}
	if path.asPath.CheckLoop() {
		return nil, fmt.Errorf("AdjRibIn_Select: detect AS loop")
	}
	// Insert into Adj-Rib-In
	if err := r.Insert(path); err != nil {
		return nil, fmt.Errorf("AdjRibIn_Select: %w", err)
	}

	_, bestPath, err := r.Calculate(path.nlri, bestPathConfig)
	if err != nil {
		return nil, fmt.Errorf("AdjRibIn_Select: %w", err)
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if bestPath.id != path.id {
		return nil, nil
	}
	bp := bestPath.DeepCopy()
	return &bp, nil
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

func (r *AdjRibOut) GroupByPathAttributes() [][]*Path {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
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
	if path.local {
		// if the given path is originated by local, don't insert into route table
		return nil
	}
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

func (l *LocRib) GetAll() []*Path {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	pathes := make([]*Path, 0)
	for _, path := range l.table {
		pathes = append(pathes, path)
	}
	return pathes
}

func (l *LocRib) GetByGroup() map[int][]*Path {
	pathMap := make(map[int][]*Path)
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	for _, path := range l.table {
		if _, ok := pathMap[path.group]; !ok {
			pathMap[path.group] = []*Path{path}
		} else {
			pathMap[path.group] = append(pathMap[path.group], path)
		}
	}
	return pathMap
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
