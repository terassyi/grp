package rip

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/terassyi/grp/pb"
	"github.com/terassyi/grp/pkg/log"
	routeManager "github.com/terassyi/grp/pkg/route"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

var (
	RipRequestNoEntry error = errors.New("Request no entry")
)

var mask24 net.IPMask = net.IPv4Mask(0xff, 0xff, 0xff, 0x00)

// Routing Information Protocol
// https://datatracker.ietf.org/doc/html/rfc1058
type Rip struct {
	links   []netlink.Link
	addrs   []*broadcastAddress
	port    int
	rx      chan message
	tx      chan message
	timer   time.Ticker
	gcTimer time.Ticker
	trigger chan struct{}
	timeout uint64
	gcTime  uint64

	mutex                sync.RWMutex
	table                *table
	logger               log.Logger
	routeManagerEndpoint string
	routeManager         pb.RouteApiClient
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

type broadcastAddress struct {
	addr      net.IP
	broadcast net.IP
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

func New(networks []string, port int, timeout, gcTime uint64, logLevel int, logOutput string) (*Rip, error) {
	links := make([]netlink.Link, 0, len(networks))
	addrs := make([]*broadcastAddress, 0, len(networks))
	table := newTable()
	for _, network := range networks {
		_, cidr, err := net.ParseCIDR(network)
		if err != nil {
			return nil, err
		}
		link, err := linkFromNetwork(network)
		if err != nil {
			return nil, err
		}
		links = append(links, link)

		addrss, err := netlink.AddrList(link, nl.FAMILY_V4)
		if err != nil {
			return nil, err
		}
		var b *broadcastAddress
		for _, a := range addrss {
			if cidr.Contains(a.IP) {
				b = &broadcastAddress{
					addr:      a.IP,
					broadcast: broadcast(cidr),
				}
			}
		}
		if b == nil {
			return nil, fmt.Errorf("gateway addres  is not found for %s", network)
		}
		addrs = append(addrs, b)
		if err := table.addLocalRoute(cidr, link, b.addr); err != nil {
			return nil, err
		}
	}
	logger, err := log.New(log.Level(logLevel), logOutput)
	if err != nil {
		return nil, err
	}
	logger.SetProtocol("rip")
	if timeout == 0 {
		timeout = DEFALUT_TIMEOUT
	}
	if gcTime == 0 {
		gcTime = DEFALUT_GC_TIME
	}
	return &Rip{
		rx:                   make(chan message, 128),
		tx:                   make(chan message, 128),
		links:                links,
		addrs:                addrs,
		port:                 port,
		timer:                *time.NewTicker(time.Second * 30),
		gcTimer:              *time.NewTicker(time.Second * 1),
		trigger:              make(chan struct{}, 1),
		timeout:              timeout,
		gcTime:               gcTime,
		table:                table,
		logger:               logger,
		routeManagerEndpoint: fmt.Sprintf("%s:%d", routeManager.DefaultRouteManagerHost, routeManager.DefaultRouteManagerPort),
	}, nil
}

func FromConfig(config *Config, logLevel int, logOut string) (*Rip, error) {
	port := PORT
	timeout := DEFALUT_TIMEOUT
	gc := DEFALUT_GC_TIME
	if config.Port != 0 {
		port = config.Port
	}
	if config.Timeout != 0 {
		timeout = uint64(config.Timeout)
	}
	if config.Gc != 0 {
		gc = uint64(config.Gc)
	}
	return New(config.Networks, port, timeout, gc, logLevel, logOut)
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
	for i := range r.links {
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
			addr: r.addrs[i].broadcast,
			port: PORT,
			data: d,
		}
	}
	return nil
}

func (r *Rip) PollWithContext(ctx context.Context) error {
	r.logger.Info("RIPv1 daemon start")
	client, err := routeManager.NewRouteManagerClient(r.routeManagerEndpoint)
	if err != nil {
		r.logger.Warn("Route manager is not running")
	}
	r.routeManager = client
	for idx, link := range r.links {
		r.logger.Info("associate link=%s", link.Attrs().Name)
		go r.rxtxLoopLink(ctx, idx)
	}
	if err := r.init(); err != nil {
		return err
	}
	r.poll(ctx)
	return nil
}

func (r *Rip) poll(ctx context.Context) {
	r.logger.Info("GRP RIP polling start...")
	for {
		select {
		case <-r.timer.C:
			if err := r.regularUpdate(); err != nil {
				r.logger.Err("%v", err)
			}
		case <-r.gcTimer.C:
			if err := r.gc(); err != nil {
				r.logger.Err("%v", err)
			}
		case <-r.trigger:
			if err := r.triggerUpdate(); err != nil {
				r.logger.Err("%v", err)
			}
		case msg := <-r.rx:
			if err := r.rxHandle(&msg); err != nil {
				r.logger.Err("%v", err)
			}
		case <-ctx.Done():
			return
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
		if msg.addr.Equal(a.broadcast) || msg.addr.Equal(a.addr) {
			return nil
		}
	}
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
		route := r.lookupTable(&net.IPNet{IP: e.Address, Mask: mask24})
		if route == nil {
			res.Entries = append(res.Entries, &Entry{Family: netlink.FAMILY_V4, Address: e.Address, Metric: INF})
		} else {
			res.Entries = append(res.Entries, &Entry{Family: netlink.FAMILY_V4, Address: route.dst.IP, Metric: route.metric})
		}
	}
	return res, nil
}

func (r *Rip) response(link netlink.Link, addr net.IP, packet *Packet) error {
	// if src port is not 520, ignore
	// update neighbor information
	a := &net.IPNet{IP: addr, Mask: mask24}
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
		route := r.lookupTable(&net.IPNet{IP: e.Address, Mask: mask24})
		if route == nil {
			// route is not exist
			if err := r.commit(link, &net.IPNet{IP: e.Address, Mask: mask24}, addr, e.Metric+1, false); err != nil {
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
		IP:   r.addrs[index].broadcast,
		Port: r.port,
	}
	r.logger.Info("RIP Rx routine start... %s %s:%d", r.links[index].Attrs().Name, r.addrs[index].broadcast, r.port)
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
			gw := rr.gateway.String()
			if _, err := r.routeManager.DeleteRoute(context.Background(), &pb.DeleteRouteRequest{
				Route: &pb.Route{
					Destination: rr.dst.String(),
					Gw:          &gw,
					Link:        link.Attrs().Name,
					Protocol:    pb.Protocol(routeManager.RT_PROTO_RIP),
				},
			}); err != nil {
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
		r.tx <- message{addr: addr.broadcast, port: PORT, data: data}
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
	r.logger.Info("commit: %s %s to %s metric=%d", dst.String(), gateway.String(), link.Attrs().Name, metric)
	if replace {
		route := r.lookupTable(dst)
		if route == nil {
			return fmt.Errorf("Route is not exist.")
		}
		route.gateway = gateway
		route.metric = metric
		route.state = ROUTE_STATE_UPDATED // set route update flag
		gw := gateway.String()
		if r.routeManager == nil {
			r.logger.Warn("Route manager is not running")
			return nil
		}
		if _, err := r.routeManager.SetRoute(context.Background(), &pb.SetRouteRequest{
			Route: &pb.Route{
				Destination:       dst.String(),
				Gw:                &gw,
				Link:              link.Attrs().Name,
				Protocol:          pb.Protocol(routeManager.RT_PROTO_RIP),
				BgpOriginExternal: false,
			},
		}); err != nil {
			return err
		}
		return nil
	}
	r.table.mutex.Lock()
	defer r.table.mutex.Unlock()
	r.table.routes = append(r.table.routes, &route{
		dst:     dst,
		gateway: gateway,
		out:     uint(link.Attrs().Index),
		metric:  uint32(metric),
	})
	if r.routeManager == nil {
		r.logger.Warn("Route manager is not running")
		return nil
	}
	gw := gateway.String()
	if _, err := r.routeManager.SetRoute(context.Background(), &pb.SetRouteRequest{
		Route: &pb.Route{
			Destination:       dst.String(),
			Gw:                &gw,
			Link:              link.Attrs().Name,
			Protocol:          pb.Protocol(routeManager.RT_PROTO_RIP),
			BgpOriginExternal: false,
		},
	}); err != nil {
		return err
	}
	return nil
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

func linkFromNetwork(network string) (netlink.Link, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				_, target, err := net.ParseCIDR(network)
				if err != nil {
					return nil, err
				}
				if target.Contains(v.IP) {
					return netlink.LinkByIndex(iface.Index)
				}
			}
		}
	}
	return nil, fmt.Errorf("interface that has %s is not found", network)
}

func newTable() *table {
	return &table{
		routes: make([]*route, 0),
		mutex:  sync.Mutex{},
	}
}

func (t *table) addLocalRoute(network *net.IPNet, link netlink.Link, addr net.IP) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.routes = append(t.routes, &route{
		dst:     network,
		metric:  1,
		out:     uint(link.Attrs().Index),
		gateway: addr,
	})
	return nil
}

func broadcast(cidr *net.IPNet) net.IP {
	addr := cidr.IP.To4()
	mask := cidr.Mask
	broadcast := net.IPv4(0x00, 0x00, 0x00, 0x00)
	for i := 0; i < 4; i++ {
		broadcast[i+12] = (addr[i] & mask[i]) | ^mask[i]
	}
	return broadcast
}

func (ba *broadcastAddress) network() *net.IPNet {
	mask := net.IPv4Mask(0x00, 0x00, 0x00, 0x00)
	for i := 0; i < 4; i++ {
		mask[i] = ba.addr[i] ^ ba.broadcast[i]
	}
	return &net.IPNet{
		IP:   ba.addr,
		Mask: mask,
	}
}
