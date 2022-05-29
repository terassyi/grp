package bgp

import "net"

type server struct {
	port  int
	local net.IP
}

func newServer(local net.IP, port int) (*server, error) {
	return &server{
		port:  port,
		local: local,
	}, nil
}
