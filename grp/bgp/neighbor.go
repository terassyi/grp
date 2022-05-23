package bgp

import (
	"net"
	"time"
)

type neighbor struct {
	addr net.IP
	port int
	as   int
	hold time.Duration
}

func newNeighbor(addr net.IP, asn int) *neighbor {
	return &neighbor{
		addr: addr,
		as:   asn,
	}
}
