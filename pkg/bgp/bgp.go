package bgp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/terassyi/grp/pkg/log"
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
	as           int
	port         int
	routerId     net.IP
	server       *server
	peers        map[string]*peer // key: ipaddr string, value: peer struct pointer
	locRib       *LocRib
	adjRibIn     *AdjRibIn
	networks     []*net.IPNet
	requestQueue chan *Request
	config       *BgpConfig
	logger       log.Logger
	signalCh     chan os.Signal
}

type BgpConfig struct {
	bestPathConfig *BestPathConfig
}

var (
	ErrASNumberIsRequired         error = errors.New("AS Number is required.")
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
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGTERM)
	return &Bgp{
		port:         port,
		peers:        make(map[string]*peer),
		locRib:       NewLocRib(),
		adjRibIn:     newAdjRibIn(),
		logger:       logger,
		requestQueue: make(chan *Request, 16),
		networks:     make([]*net.IPNet, 0),
		signalCh:     sigCh,
	}, nil
}

func FromConfig(conf *Config, logLevel int, logOut string) (*Bgp, error) {
	port := PORT
	if conf.Port != port {
		port = conf.Port
	}
	b, err := New(port, logLevel, logOut)
	if err != nil {
		return nil, err
	}
	if conf.AS == 0 {
		return nil, ErrASNumberIsRequired
	}
	b.setAS(conf.AS)
	if conf.RouterId != "" {
		b.setRouterId(conf.RouterId)
	}
	for _, neighbor := range conf.Neighbors {
		peerAddr := net.ParseIP(neighbor.Address)
		_, err := b.registerPeer(peerAddr, b.routerId, b.as, neighbor.AS, false)
		if err != nil {
			return nil, err
		}
	}
	for _, target := range conf.Networks {
		_, cidr, err := net.ParseCIDR(target)
		if err != nil {
			return nil, err
		}
		b.networks = append(b.networks, cidr)
	}
	return b, nil
}

func (b *Bgp) setAS(as int) error {
	b.as = as
	b.logger.Infof("AS Number: %d", as)
	return nil
}

func (b *Bgp) setRouterId(routerId string) error {
	b.routerId = net.ParseIP(routerId)
	b.logger.Infof("Router ID: %s", routerId)
	return nil
}

func (b *Bgp) Poll() error {
	b.logger.Infoln("BGP daemon start.")
	for _, network := range b.networks {
		localPath, err := CreateLocalPath(network.String(), b.as)
		if err != nil {
			return err
		}
		b.adjRibIn.Insert(localPath)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := b.poll(ctx); err != nil { // BGP daemon main routine
		cancel()
		return err
	}
	for _, p := range b.peers {
		cctx, _ := context.WithCancel(ctx)
		go p.poll(cctx)
		p.enqueueEvent(&bgpStart{})
	}
	<-b.signalCh
	b.logger.Infof("Receive a signal. Terminate GRP BGP daemon.")
	cancel()
	return nil
}

func (b *Bgp) poll(ctx context.Context) error {
	// request handling routine
	b.logger.Infof("GRP BGP Polling Start.")
	go func() {
		for {
			select {
			case req := <-b.requestQueue:
				if err := b.requestHandle(ctx, req); err != nil {
					b.logger.Errorf("Request handler: %s", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	go b.pollRib(ctx)
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
		peer, err := b.registerPeer(addr, b.routerId, b.as, as, false)
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
		if len(req.Args) < 1 {
			return ErrInvalidBgpApiArguments
		}

	default:
		return ErrUnknownBgpApiRequest
	}

	return nil
}

func (b *Bgp) registerPeer(addr, routerId net.IP, myAS, peerAS int, force bool) (*peer, error) {
	idx, local, err := lookupLocalAddr(addr)
	if err != nil {
		return nil, err
	}
	link, err := netlink.LinkByIndex(idx)
	if err != nil {
		return nil, err
	}
	ri := routerId
	if ri == nil {
		// If router-id is not specified, pick the largest IP address in the host
		ri, err = PickLargestAddr()
		if err != nil {
			return nil, err
		}
	}
	p := newPeer(b.logger, link, local, addr, ri, myAS, peerAS, b.locRib, b.adjRibIn)
	if _, ok := b.peers[addr.String()]; ok && !force {
		return nil, ErrPeerAlreadyRegistered
	}
	p.logger.Infof("Register peer local->%s remote->%s ASN->%d", local, addr, peerAS)
	b.peers[addr.String()] = p
	return p, nil
}

func (b *Bgp) pollRib(ctx context.Context) {
	b.logger.Infof("LocRib poll start")
	for {
		select {
		case pathes := <-b.locRib.queue:
			for _, path := range pathes {
				if err := b.locRib.Insert(path); err != nil {
					b.logger.Errorf("pollRib: ", err)
				}
			}
			for _, peer := range b.peers {
				peer.enqueueEvent(&triggerDissemination{pathes: pathes, withdrawn: false})
			}
		case <-ctx.Done():
			return
		}
	}
}

func (b *Bgp) originateRoutes(networks []*net.IPNet) error {
	pathes := make([]*Path, len(networks))
	for _, network := range networks {
		path, err := CreateLocalPath(network.String(), b.as)
		if err != nil {
			return fmt.Errorf("Bgp_originateRoutes: %w", err)
		}
		selected, err := b.adjRibIn.Select(b.as, path, false, b.config.bestPathConfig)
		if err != nil {
			return fmt.Errorf("Bgp_originateRoutes: %w", err)
		}
		if selected == nil {
			continue
		}
		pathes = append(pathes, selected)
	}

	return nil
}
