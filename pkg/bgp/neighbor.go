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
