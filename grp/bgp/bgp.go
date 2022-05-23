package bgp

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/terassyi/grp/grp/log"
	"github.com/vishvananda/netlink"
)

// RFC 1771 Version BGP-4
// https://www.rfc-editor.org/rfc/pdfrfc/rfc1771.txt.pdf

const (
	PORT    int = 179
	VERSION int = 4
)

type Bgp struct {
	// tx    chan message
	// rx    chan message
	port         int
	server       *server
	peers        map[string]*peer  // key: ipaddr string, value: peer struct pointer
	neighborMap  map[string]net.IP // key: local addr, value: neighbor addr
	requestQueue chan *Request
	logger       log.Logger
}

type state uint8

const (
	// Idle state:
	// In this state BGP refuses all incoming BGP connections.
	// No resources are allocated to the peer.
	// In response to the Start event, the local system initialized all BGP resources.
	IDLE         state = iota
	CONNECT      state = iota
	ACTIVE       state = iota
	OPEN_SENT    state = iota
	OPEN_CONFIRM state = iota
	ESTABLISHED  state = iota
)

func (s state) String() string {
	switch s {
	case IDLE:
		return "IDLE"
	case CONNECT:
		return "CONNECT"
	case ACTIVE:
		return "ACTIVE"
	case OPEN_SENT:
		return "OPEN_SENT"
	case OPEN_CONFIRM:
		return "OPEN_CONFIRM"
	case ESTABLISHED:
		return "ESTABLISHED"
	default:
		return "Unknown"
	}
}

type message struct {
	data []byte
}

var (
	ErrInvalidBgpState            error = errors.New("Invalid BGP state.")
	ErrPeerAlreadyRegistered      error = errors.New("Peer already registered.")
	ErrInvalidNeighborAddress     error = errors.New("Invalid neighbor address.")
	ErrInvalidBgpApiArguments     error = errors.New("Invalid BGP API arguments.")
	ErrUnknownBgpApiRequest       error = errors.New("Unknown BGP API request.")
	ErrInvalidEventType           error = errors.New("Invalid BGP Event type.")
	ErrEventQueueNotExist         error = errors.New("Event queue does'nt exist.")
	ErrUnsupportedAddrType        error = errors.New("Unsupported Address Type.")
	ErrGivenAddrIsNotNeighbor     error = errors.New("Given address is not a neighbor.")
	ErrInvalidEventInCurrentState error = errors.New("Invalid event in current state.")
	ErrUnreachableState           error = errors.New("Unreachable state.")
	ErrTransportConnectionClose   error = errors.New("Transport connection is closed.")
)

func New(port int, logLevel int, out string) (*Bgp, error) {
	logger, err := log.New(log.Level(logLevel), out)
	if err != nil {
		return nil, err
	}
	return &Bgp{
		port:         port,
		peers:        make(map[string]*peer),
		logger:       logger,
		requestQueue: make(chan *Request, 16),
	}, nil
}

func (b *Bgp) Poll() error {
	b.logger.Infoln("BGP daemon start.")
	return b.poll()
}

func (b *Bgp) poll() error {
	ctx, _ := context.WithCancel(context.Background())
	// request handling routine
	go func() {
		for {
			select {
			case req := <-b.requestQueue:
				if err := b.requestHandle(ctx, req); err != nil {
					b.logger.Errorf("Request handler: %s", err)
				}
			}
		}
	}()
	time.Sleep(time.Second * 4)
	b.logger.Infoln("sleep finish. submit request.")
	b.requestQueue <- &Request{Command: "neighbor", Args: []string{"10.1.0.3", "remote-as", "1"}}
	for {
	}
	return nil
}

func (b *Bgp) requestHandle(ctx context.Context, req *Request) error {
	b.logger.Infof("Receive request: %v\n", req)
	switch req.Command {
	case "neighbor":
		if len(req.Args) != 3 {
			return ErrInvalidBgpApiArguments
		}
		addr := net.ParseIP(req.Args[0])
		as := 0
		if req.Args[1] == "remote-as" {
			a, err := strconv.Atoi(req.Args[2])
			if err != nil {
				return err
			}
			as = a
		}
		// create new peer
		peer, err := b.registerPeer(addr, as, false)
		if err != nil {
			return err
		}
		go peer.poll(ctx)
		// wait for running peer.poll goroutine
		time.Sleep(time.Second)
		// enqueue BGP start event
		if err := peer.enqueueEvent(&bgpStart{}); err != nil {
			return err
		}
	case "network":
	default:
		return ErrUnknownBgpApiRequest
	}

	return nil
}

func (b *Bgp) registerPeer(addr net.IP, as int, force bool) (*peer, error) {
	idx, local, err := lookupLocalAddr(addr)
	if err != nil {
		return nil, err
	}
	link, err := netlink.LinkByIndex(idx)
	if err != nil {
		return nil, err
	}
	p := newPeer(b.logger, link, local, addr, as)
	if _, ok := b.peers[addr.String()]; ok && !force {
		return nil, ErrPeerAlreadyRegistered
	}
	p.logger.Infof("Register peer local->%s remote->%s ASN->%d", local, addr, as)
	b.peers[addr.String()] = p
	return p, nil
}

func (b *Bgp) listen() (net.Listener, error) {
	tcpAddr := &net.TCPAddr{Port: b.port}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, nil
	}
	return listener, nil
}
