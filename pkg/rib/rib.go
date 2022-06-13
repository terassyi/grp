package rib

import (
	"net"

	"github.com/vishvananda/netlink"
)

type Route struct {
	*netlink.Route
	Link netlink.Link
}

const (
	RT_PROTO_UNSPEC   int = 0
	RT_PROTO_REDIRECT int = 1
	RT_PROTO_KERNEL   int = 2
	RT_PROTO_BOOT     int = 3
	RT_PROTO_STATIC   int = 4
	RT_PROTO_BGP      int = 186
	RT_PROTO_ISIS     int = 187
	RT_PROTO_OSPF     int = 188
	RT_PROTO_RIP      int = 189
)

func LookUp4(link netlink.Link) ([]netlink.Route, error) {
	handler, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	return handler.RouteList(link, netlink.FAMILY_V4)
}

func Get4(link netlink.Link, addr net.IP) ([]netlink.Route, error) {
	handler, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	return handler.RouteGet(addr)
}

func Add4(link netlink.Link, dst *net.IPNet, gateway net.IP, proto int) error {
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
	linkAddrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	if dst.Contains(linkAddrs[0].IP) {
		target.Scope = netlink.SCOPE_LINK
	}
	return handler.RouteAdd(target)
}

func Delete4(link netlink.Link, dst *net.IPNet, src net.IP, gateway net.IP) error {
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

func Replace4(link netlink.Link, dst *net.IPNet, gateway net.IP, proto int) error {
	handler, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	return handler.RouteReplace(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Gw:        gateway,
		Protocol:  proto,
	})
}
