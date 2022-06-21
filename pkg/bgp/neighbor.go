package bgp

import (
	"net"
	"time"
)

type neighbor struct {
	addr         net.IP
	port         int
	as           int
	routerId     net.IP
	hold         time.Duration
	capabilities []Capability
}

func newNeighbor(addr net.IP, asn int) *neighbor {
	return &neighbor{
		addr:         addr,
		as:           asn,
		capabilities: make([]Capability, 0),
	}
}

func (n *neighbor) SetRouterId(id net.IP) {
	n.routerId = id
}

func (n *neighbor) SetPort(port int) {
	n.port = port
}

func (n *neighbor) Equal(target *neighbor) bool {
	if !n.addr.Equal(target.addr) || !n.routerId.Equal(target.routerId) {
		return false
	}
	if n.port != target.port {
		return false
	}
	if n.as != target.as {
		return false
	}
	return true
}
