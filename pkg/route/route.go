package route

import (
	"fmt"
	"net"

	"github.com/terassyi/grp/pb"
	"github.com/vishvananda/netlink"
)

type Route struct {
	netlink.Route
	Ad AdministrativeDistance
}

func RouteFromNetLink(route netlink.Route, bgpExternal bool) *Route {
	ad := AdFromProto(route.Protocol, bgpExternal)
	return &Route{
		Route: route,
		Ad:    ad,
	}
}

func RouteFromReq(req *pb.Route) (*Route, error) {
	link, err := netlink.LinkByName(req.GetLink())
	if err != nil {
		return nil, err
	}
	_, dst, err := net.ParseCIDR(req.GetDestination())
	if err != nil {
		return nil, err
	}
	var src, gw net.IP
	if req.Gw != nil {
		gw = net.ParseIP(*req.Gw)
	}
	if req.Src != nil {
		src = net.ParseIP(*req.Src)
	}
	return &Route{
		Route: netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       dst,
			Src:       src,
			Gw:        gw,
			Protocol:  int(req.GetProtocol()),
		},
		Ad: AdFromProto(int(req.Protocol), req.BgpOriginExternal),
	}, nil
}

func (r *Route) add() error {
	link, err := netlink.LinkByIndex(r.LinkIndex)
	if err != nil {
		return err
	}
	return add4(link, r.Dst, r.Gw, r.Protocol)
}

func (r *Route) replace() error {
	link, err := netlink.LinkByIndex(r.LinkIndex)
	if err != nil {
		return err
	}
	return replace4(link, r.Dst, r.Gw, r.Protocol)
}

func (r Route) String() string {
	var d, s, g string
	if r.Gw != nil {
		g = fmt.Sprintf("gw=%s", r.Gw.String())
	}
	if r.Src != nil {
		s = fmt.Sprintf("src=%s", r.Src.String())
	}
	if r.Dst == nil {
		d = "default"
	} else {
		d = r.Dst.String()
	}
	return fmt.Sprintf("dst=%s %s %s proto=%s ifindex=%d ad=%s", d, s, g, ProtoString(r.Protocol), r.LinkIndex, r.Ad)
}

func lookUp4(link netlink.Link) ([]netlink.Route, error) {
	handler, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	return handler.RouteList(link, netlink.FAMILY_V4)
}

func get4(link netlink.Link, addr net.IP) ([]netlink.Route, error) {
	handler, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("NewHandle: %w", err)
	}
	routes, err := handler.RouteGet(addr)
	if err != nil {
		if err.Error() == "network is unreachable" {
			return nil, nil
		}
		return nil, fmt.Errorf("RouteGet(%s): %w", addr, err)
	}
	return routes, nil
}

func add4(link netlink.Link, dst *net.IPNet, gateway net.IP, proto int) error {
	handler, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("Add: NewHandle: %w", err)
	}
	target := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Gw:        gateway,
		Protocol:  proto,
	}
	linkAddrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("Add: AddrList(%s): %w", link.Attrs().Name, err)
	}
	if dst.Contains(linkAddrs[0].IP) {
		target.Scope = netlink.SCOPE_LINK
	}
	if err := handler.RouteAdd(target); err != nil {
		return fmt.Errorf("Add: err=[%w] route=[%s]", err, target.String())
	}
	return nil
}

func delete4(link netlink.Link, dst *net.IPNet, src net.IP, gateway net.IP) error {
	handler, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	target := &netlink.Route{
		Dst: dst,
		Gw:  gateway,
		Src: src,
	}
	linkAddrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	if dst.Contains(linkAddrs[0].IP) {
		target.Scope = netlink.SCOPE_LINK
	}
	return handler.RouteDel(target)
}

func replace4(link netlink.Link, dst *net.IPNet, gateway net.IP, proto int) error {
	handler, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	target := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Gw:        gateway,
		Protocol:  proto,
	}
	if err := handler.RouteReplace(target); err != nil {
		return fmt.Errorf("Replace: err=[%w] route=[%s]", err, target.String())
	}
	return nil
}

func (r *Route) complareWithAd(target *Route) bool {
	return r.Ad < target.Ad
}
