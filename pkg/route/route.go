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
