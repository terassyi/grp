package bgp

import (
	"net"
	"time"
)

type neighbor struct {
	addr         net.IP
	port         int
	as           int
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
