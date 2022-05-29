package bgp

import (
	"encoding/binary"
	"net"
	"strconv"
)

func lookupLocalAddr(remote net.IP) (int, net.IP, error) {
	ifList, err := net.Interfaces()
	if err != nil {
		return -1, nil, err
	}
	for _, i := range ifList {
		addrs, err := i.Addrs()
		if err != nil {
			return -1, nil, err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.Contains(remote) {
					return i.Index, v.IP, nil
				}
			}
		}
	}
	return -1, nil, ErrGivenAddrIsNotNeighbor
}

func PickLargestAddr() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var m uint32 = 0
	var res net.IP
	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			if !v.IP.IsLoopback() {
				var a uint32
				binary.BigEndian.PutUint32([]byte(v.IP.To4()), a)
				if a > m {
					m = a
					res = v.IP.To4()
				}
			}
		}
	}
	return res, nil
}

func SplitAddrAndPort(host string) (net.IP, int, error) {
	// Expected host format: ipv4:port or [ipv6]:port
	a, p, err := net.SplitHostPort(host)
	if err != nil {
		return nil, -1, err
	}
	pn, err := strconv.Atoi(p)
	if err != nil {
		return nil, -1, err
	}
	return net.ParseIP(a), pn, nil
}
