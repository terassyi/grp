package rib

import (
	"fmt"
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

func Add4(link netlink.Link, dst *net.IPNet, gateway net.IP, proto int) error {
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
