package rip

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/terassyi/grp/pkg/log"
	"github.com/terassyi/grp/pkg/rib"
	"github.com/vishvananda/netlink"
)

var (
	RipRequestNoEntry error = errors.New("Request no entry")
)

// Routing Information Protocol
// https://datatracker.ietf.org/doc/html/rfc1058
type Rip struct {
	links   []netlink.Link
	addrs   []netlink.Addr
	port    int
	rx      chan message
	tx      chan message
	timer   time.Ticker
	gcTimer time.Ticker
	trigger chan struct{}
	timeout uint64
	gcTime  uint64
	table   *table
	logger  log.Logger
}

// RIP packet format
type Packet struct {
	Command RipCommand
	Version uint8
	Entries []*Entry
}

type Entry struct {
	Family  uint16
	Address net.IP
	Metric  uint32
}

type table struct {
	mutex  sync.Mutex
	routes []*route
}

type route struct {
	dst     *net.IPNet
	metric  uint32
	out     uint // out interface for split horizon
	gateway net.IP
	state   routeState
	timer   uint64
}

type message struct {
	ifindex int
	addr    net.IP
	port    int
	data    []byte
}

type RipCommand uint8

const (
	REQUEST   RipCommand = iota + 1 // A request for the responding system to send all or part of its routing table.
	RESPONSE  RipCommand = iota + 1 // A message containing all or part of the sender's routing table. This message may be sent in response to a request or poll, or it may be an update message generated by the sender.
	TRACEON   RipCommand = iota + 1 // Obsolete. Messages containing this command are to be ignored.
	TRACEOFF  RipCommand = iota + 1 // Obsolete. Messages containing this command are to be ignored.
	RESERVERD RipCommand = iota + 1 // This value is used by Syn Microsystems for its own purposes.
)

type routeState uint8

const (
	ROUTE_STATE_ALIVE   routeState = iota
	ROUTE_STATE_TIMEOUT routeState = iota
	ROUTE_STATE_GC      routeState = iota
	ROUTE_STATE_UPDATED routeState = iota
)

const (
	PORT            int    = 520
	INF             uint32 = 16
	DEFALUT_TIMEOUT uint64 = 180
	DEFALUT_GC_TIME uint64 = 120
)

var (
	DEFAULT_ROUTE  net.IP = net.IP([]byte{0x00, 0x00, 0x00, 0x00})
	BROADCAST      net.IP = net.IP([]byte{0xff, 0xff, 0xff, 0xff})
	addrFamilyIpv4 uint16 = 2
)

func New(names []string, port int, timeout, gcTime uint64, logLevel int, logOutput string) (*Rip, error) {
	links := make([]netlink.Link, 0, len(names))
	addrs := make([]netlink.Addr, 0, len(names))
	for _, name := range names {
		link, err := netlink.LinkByName(name)
		if err != nil {
			return nil, err
		}
		addr, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			return nil, err
		}
		if len(addr) == 0 {
			return nil, fmt.Errorf("Address is not allocated to %d", link.Attrs().Index)
		}
		addrs = append(addrs, addr[0])
		links = append(links, link)
	}
	table := &table{
		routes: make([]*route, 0),
		mutex:  sync.Mutex{},
	}
	logger, err := log.New(log.Level(logLevel), logOutput)
	if err != nil {
		return nil, err
	}
	return &Rip{
		// rx:      make(chan []byte, 16),
		// tx:      make(chan *Packet, 16),
		rx:      make(chan message, 128),
		tx:      make(chan message, 128),
		links:   links,
		addrs:   addrs,
		port:    port,
		timer:   *time.NewTicker(time.Second * 30),
		gcTimer: *time.NewTicker(time.Second * 1),
		trigger: make(chan struct{}, 1),
		timeout: timeout,
		gcTime:  gcTime,
		table:   table,
		logger:  logger,
	}, nil
}

func Parse(data []byte) (*Packet, error) {
	buf := bytes.NewBuffer(data)
	packet := &Packet{}
	if err := binary.Read(buf, binary.BigEndian, &packet.Command); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &packet.Version); err != nil {
		return nil, err
	}
	var zero16 uint16
	if err := binary.Read(buf, binary.BigEndian, &zero16); err != nil {
		return nil, err
	}
	packet.Entries = make([]*Entry, 0)
	for buf.Len() > 0 {
		entry := &Entry{}
		if err := binary.Read(buf, binary.BigEndian, &entry.Family); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &zero16); err != nil {
			return nil, err
		}
		addr := [4]byte{0, 0, 0, 0}
		if err := binary.Read(buf, binary.BigEndian, &addr); err != nil {
			return nil, err
		}
		entry.Address = net.IP(addr[:])
		for i := 0; i < 3; i++ {
			if err := binary.Read(buf, binary.BigEndian, &entry.Metric); err != nil {
				return nil, err
			}
		}
		packet.Entries = append(packet.Entries, entry)
	}
	return packet, nil
}

func (p *Packet) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 6*4))
	if err := binary.Write(buf, binary.BigEndian, p.Command); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.Version); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
		return nil, err
	}
	for _, e := range p.Entries {
		if err := binary.Write(buf, binary.BigEndian, e.Family); err != nil {
			return nil, err
		}
		if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
			return nil, err
		}
		if err := binary.Write(buf, binary.BigEndian, e.Address); err != nil {
			return nil, err
		}
		for i := 0; i < 2; i++ {
			if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
				return nil, err
			}
		}
		if err := binary.Write(buf, binary.BigEndian, e.Metric); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (p *Packet) Dump() string {
	str := ""
	str += fmt.Sprintf("protocol = RIPv%d command = %s routes = %d\n", p.Version, p.Command, len(p.Entries))
	for _, e := range p.Entries {
		str += fmt.Sprintf("  %s\n", e.dump())
	}
	return str
}

func (e *Entry) dump() string {
	return fmt.Sprintf("addr = %s metric = %d", e.Address, e.Metric)
}

func buildPacket(cmd RipCommand, routes []*route) (*Packet, error) {
	packet := &Packet{Command: cmd, Version: 1}
	entries := make([]*Entry, 0, len(routes))
	for _, rr := range routes {
		entry := &Entry{
			Family:  netlink.FAMILY_V4,
			Address: rr.dst.IP,
			Metric:  rr.metric,
		}
		entries = append(entries, entry)
	}
	packet.Entries = entries
	return packet, nil
}

func (r *Rip) init() error {
	for i, link := range r.links {
		routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
		if err != nil {
			return err
		}
		r.table.mutex.Lock()
		for _, rr := range routes {
			if rr.Dst == nil {
				continue
			}
			r.table.routes = append(r.table.routes, &route{
				dst:     rr.Dst,
				gateway: rr.Src,
				out:     uint(link.Attrs().Index),
				metric:  1,
			})
		}
		r.table.mutex.Unlock()
		// initial request
		def := &net.IPNet{IP: DEFAULT_ROUTE, Mask: net.IPMask{0, 0, 0, 0}}
		emptyRoutes := []*route{{dst: def, metric: INF}}
		p, err := buildPacket(REQUEST, emptyRoutes)
		if err != nil {
			return err
		}
		d, err := p.Decode()
		if err != nil {
			return err
		}
		r.tx <- message{
			addr: r.addrs[i].Broadcast,
			port: PORT,
			data: d,
		}
	}
	return nil
}

func (r *Rip) Poll() error {
	ctx, _ := context.WithCancel(context.Background())
	for idx := range r.links {
		go r.rxtxLoopLink(ctx, idx)
	}
	if err := r.init(); err != nil {
		return err
	}
	r.poll()
	return nil
}

func (r *Rip) poll() {
	r.logger.Info("GRP RIP polling start...")
	for {
		select {
		case <-r.timer.C:
			if err := r.regularUpdate(); err != nil {
				r.logger.Err("RIP ERROR: %v\n", err)
			}
		case <-r.gcTimer.C:
			if err := r.gc(); err != nil {
				r.logger.Err("RIP ERROR: %v\n", err)
			}
		case <-r.trigger:
			if err := r.triggerUpdate(); err != nil {
				r.logger.Err("RIP ERROR: %v\n", err)
			}
		case msg := <-r.rx:
			if err := r.rxHandle(&msg); err != nil {
				r.logger.Err("RIP ERROR: %v\n", err)
			}
		}
	}
}

func (r *Rip) rxHandle(msg *message) error {
	packet, err := Parse(msg.data)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if packet.Version == 0 || packet.Version > 1 {
		return fmt.Errorf("Invalid version")
	}
	for _, a := range r.addrs {
		if msg.addr.Equal(a.Broadcast) || msg.addr.Equal(a.IP) {
			return nil
		}
	}
	r.logger.Info("%s:%d\n", msg.addr, msg.port)
	switch packet.Command {
	case REQUEST:
		res, err := r.request(packet)
		if err != nil {
			return err
		}
		resData, err := res.Decode()
		if err != nil {
			return err
		}
		r.tx <- message{addr: msg.addr, port: msg.port, data: resData}
		return nil
	case RESPONSE:
		l, err := netlink.LinkByIndex(msg.ifindex)
		if err != nil {
			return err
		}
		return r.response(l, msg.addr, packet)
	default:
		return fmt.Errorf("Invalid command")
	}
}

func (r *Rip) request(req *Packet) (*Packet, error) {
	return r.buildResponse(req)
}

func (r *Rip) buildResponse(req *Packet) (*Packet, error) {
	if len(req.Entries) == 0 {
		return nil, RipRequestNoEntry
	}
	if len(req.Entries) == 1 {
		if req.Entries[0].Family == 0 || req.Entries[0].Metric == INF {
			return buildPacket(RESPONSE, r.allRoutes())
		}
	}
	res := &Packet{Command: RESPONSE, Version: 1, Entries: make([]*Entry, 0)}
	for _, e := range req.Entries {
		route := r.lookupTable(&net.IPNet{IP: e.Address, Mask: e.Address.To4().DefaultMask()})
		if route == nil {
			res.Entries = append(res.Entries, &Entry{Family: netlink.FAMILY_V4, Address: e.Address, Metric: INF})
		} else {
			res.Entries = append(res.Entries, &Entry{Family: netlink.FAMILY_V4, Address: route.dst.IP, Metric: route.metric})
		}
	}
	return res, nil
}

func (r *Rip) response(link netlink.Link, addr net.IP, packet *Packet) error {
	r.logger.Info(r.dumpTable())
	// if src port is not 520, ignore
	// update neighbor information
	a := &net.IPNet{IP: addr, Mask: addr.To4().DefaultMask()}
	_, cidr, err := net.ParseCIDR(a.String())
	if err != nil {
		return err
	}
	neigh := r.lookupTable(cidr)
	if neigh == nil {
		// unregistered neighbor rip router
		if err := r.commit(link, cidr, addr, 1, false); err != nil {
			return err
		}
	} else {
		if neigh.metric > 1 {
			// if neighbor's metric is not minimum, update it.
			if err := r.commit(link, cidr, addr, 1, true); err != nil {
				return err
			}
		}
	}
	// restart timeout timer
	r.table.mutex.Lock()
	neigh.timer = 0
	r.table.mutex.Unlock()
	for _, e := range packet.Entries {
		if e.Metric >= INF {
			continue
		}
		if e.Family != addrFamilyIpv4 {
			continue
		}
		route := r.lookupTable(&net.IPNet{IP: e.Address, Mask: e.Address.To4().DefaultMask()})
		if route == nil {
			// route is not exist
			if err := r.commit(link, &net.IPNet{IP: e.Address, Mask: e.Address.To4().DefaultMask()}, addr, e.Metric+1, false); err != nil {
				return err
			}
		} else {
			// route exist
			metric := e.Metric + 1
			if metric > INF {
				metric = INF
			}
			// first check the gateway
			if route.gateway == nil {
				// if route.gateway is not set, this route is for neigh
				continue
			}
			if addr.Equal(route.gateway) {
				// restart timer
				r.table.mutex.Lock()
				route.timer = 0
				// check metric
				if metric != route.metric {
					route.metric = metric
				}
				r.table.mutex.Unlock()
				if route.metric > metric {
					if err := r.commit(link, route.dst, route.gateway, metric, true); err != nil {
						return err
					}
				}
			} else {
				// gateway is not matched
				if route.metric > metric {
					if err := r.commit(link, route.dst, addr, metric, true); err != nil {
						return err
					}
				}

			}
		}
	}
	return nil
}

func (r *Rip) rxtxLoopLink(ctx context.Context, index int) error {
	host := &net.UDPAddr{
		IP:   r.addrs[index].Broadcast,
		Port: r.port,
	}
	r.logger.Info("RIP Rx routine start... %s:%d\n", r.addrs[index].Broadcast, r.port)
	listener, err := net.ListenUDP("udp", host)
	if err != nil {
		return err
	}
	for {
		select {
		case txData := <-r.tx:
			listener.WriteToUDP(txData.data, &net.UDPAddr{IP: txData.addr, Port: txData.port})
		default:
			buf := make([]byte, 512)
			n, peer, err := listener.ReadFromUDP(buf)
			if err != nil {
				return err
			}
			r.rx <- message{ifindex: r.links[index].Attrs().Index, addr: peer.IP, port: peer.Port, data: buf[:n]}
		}
	}
}

func (r *Rip) gc() error {
	flag := false
	r.table.mutex.Lock()
	newTable := make([]*route, 0, len(r.table.routes))
	for _, rr := range r.table.routes {
		rr.timer++
		if rr.timer >= r.timeout+r.gcTime && rr.state == ROUTE_STATE_TIMEOUT {
			link, err := netlink.LinkByIndex(int(rr.out))
			if err != nil {
				return err
			}
			if err := rib.Delete4(link, rr.dst, nil, nil); err != nil {
				return err
			}
			continue
		} else if rr.timer >= r.timeout {
			rr.metric = INF
			if rr.state == ROUTE_STATE_ALIVE {
				flag = true
			}
			rr.state = ROUTE_STATE_TIMEOUT
		}
		newTable = append(newTable, rr)
	}
	r.table.routes = newTable
	r.table.mutex.Unlock()
	if flag {
		r.trigger <- struct{}{}
	}
	return nil
}

func (r *Rip) regularUpdate() error {
	for _, link := range r.links {
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			return err
		}
		if len(addrs) == 0 {
			return fmt.Errorf("Address is not allocated to %d", link.Attrs().Index)
		}
		p, err := buildPacket(RESPONSE, r.allRoutes())
		if err != nil {
			return err
		}
		data, err := p.Decode()
		if err != nil {
			return err
		}
		r.tx <- message{addr: addrs[0].Broadcast, port: PORT, data: data}
	}
	return nil
}

func (r *Rip) triggerUpdate() error {
	r.logger.Info("Trigger update.")
	rt := make([]*route, 0)
	for _, rr := range r.table.routes {
		if rr.state == ROUTE_STATE_TIMEOUT || rr.state == ROUTE_STATE_UPDATED {
			rt = append(rt, rr)
			if rr.state == ROUTE_STATE_UPDATED {
				rr.state = ROUTE_STATE_ALIVE
			}
		}
	}
	p, err := buildPacket(RESPONSE, rt)
	if err != nil {
		return err
	}
	data, err := p.Decode()
	if err != nil {
		return err
	}
	for _, addr := range r.addrs {
		r.tx <- message{addr: addr.Broadcast, port: PORT, data: data}
	}
	return nil
}

func (r *Rip) allRoutes() []*route {
	r.table.mutex.Lock()
	defer r.table.mutex.Unlock()
	routes := make([]*route, 0, len(r.table.routes))
	for _, v := range r.table.routes {
		if v.dst == nil || v.gateway == nil {
			continue
		}
		routes = append(routes, v)
	}
	return routes
}

func (r *Rip) commit(link netlink.Link, dst *net.IPNet, gateway net.IP, metric uint32, replace bool) error {
	// update RIP routing database
	if replace {
		route := r.lookupTable(dst)
		if route == nil {
			return fmt.Errorf("Route is not exist.")
		}
		route.gateway = gateway
		route.metric = metric
		route.state = ROUTE_STATE_UPDATED // set route update flag
		return rib.Replace4(link, dst, gateway, rib.RT_PROTO_RIP)
	}
	r.table.mutex.Lock()
	defer r.table.mutex.Unlock()
	r.table.routes = append(r.table.routes, &route{
		dst:     dst,
		gateway: gateway,
		out:     uint(link.Attrs().Index),
		metric:  uint32(metric),
	})
	return rib.Add4(link, dst, gateway, rib.RT_PROTO_RIP)
}

func (r *Rip) lookupTable(dst *net.IPNet) *route {
	r.table.mutex.Lock()
	defer r.table.mutex.Unlock()
	for _, route := range r.table.routes {
		if route.dst.String() == dst.String() { // I'm looking forward to good way to compare IP network objects
			return route
		}
	}
	return nil
}

func (r *Rip) findRibRoutes(dst netip.Addr) ([]*rib.Route, error) {
	routes := make([]*rib.Route, 0)
	for _, link := range r.links {
		rs, err := rib.Get4(link, net.IP(dst.AsSlice()))
		if err != nil {
			return nil, err
		}
		for _, r := range rs {
			routes = append(routes, &rib.Route{Link: link, Route: &r})
		}
	}
	if len(routes) == 0 {
		return nil, nil
	}
	return routes, nil
}

func (r *Rip) dumpTable() string {
	str := "GRP RIP Routing Database\n"
	r.table.mutex.Lock()
	defer r.table.mutex.Unlock()
	for _, route := range r.table.routes {
		str += fmt.Sprintf("  dst=%s gateway=%s metric=%d ifindex=%d timer=%d\n", route.dst, route.gateway, route.metric, route.out, route.timer)
	}
	return str
}

func (cmd RipCommand) String() string {
	switch cmd {
	case REQUEST:
		return "Request"
	case RESPONSE:
		return "Response"
	default:
		return "Unknown"
	}
}
