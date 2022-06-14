package bgp

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/terassyi/grp/pkg/log"
	"github.com/vishvananda/netlink"
)

// Finite state machine of BGP

// The state transitions of BGP FSM and the actions triggered by these transitions.
// Event                Actions               Message Sent   Next State
//     --------------------------------------------------------------------
//     Idle (1)
//      1            Initialize resources            none             2
//                   Start ConnectRetry timer
//                   Initiate a transport connection
//      others               none                    none             1
//
//     Connect(2)
//      1                    none                    none             2
//      3            Complete initialization         OPEN             4
//                   Clear ConnectRetry timer
//      5            Restart ConnectRetry timer      none             3
//      7            Restart ConnectRetry timer      none             2
//                   Initiate a transport connection
//      others       Release resources               none             1
//
//     Active (3)
//      1                    none                    none             3
//      3            Complete initialization         OPEN             4
//                   Clear ConnectRetry timer
//      5            Close connection                                 3
//                   Restart ConnectRetry timer
//      7            Restart ConnectRetry timer      none             2
//                   Initiate a transport connection
//      others       Release resources               none             1
//
//     OpenSent(4)
//      1                    none                    none             4
//      4            Close transport connection      none             3
//                   Restart ConnectRetry timer
//      6            Release resources               none             1
//     10            Process OPEN is OK            KEEPALIVE          5
//                   Process OPEN failed           NOTIFICATION       1
//     others        Close transport connection    NOTIFICATION       1
//                   Release resources
//
//     OpenConfirm (5)
//      1                   none                     none             5
//      4            Release resources               none             1
//      6            Release resources               none             1
//      9            Restart KeepAlive timer       KEEPALIVE          5
//     11            Complete initialization         none             6
//                   Restart Hold Timer
//     13            Close transport connection                       1
//                   Release resources
//     others        Close transport connection    NOTIFICATION       1
//                   Release resources
//
//     Established (6)
//      1                   none                     none             6
//      4            Release resources               none             1
//      6            Release resources               none             1
//      9            Restart KeepAlive timer       KEEPALIVE          6
//     11            Restart Hold Timer            KEEPALIVE          6
//     12            Process UPDATE is OK          UPDATE             6
//                   Process UPDATE failed         NOTIFICATION       1
//     13            Close transport connection                       1
//                   Release resources
//     others        Close transport connection    NOTIFICATION       1
//                   Release resources
//    ---------------------------------------------------------------------

// The state transitions table.
// Events| Idle | Connect | Active | OpenSent | OpenConfirm | Estab
//          | (1)  |   (2)   |  (3)   |    (4)   |     (5)     |   (6)
//          |--------------------------------------------------------------
//     1    |  2   |    2    |   3    |     4    |      5      |    6
//          |      |         |        |          |             |
//     2    |  1   |    1    |   1    |     1    |      1      |    1
//          |      |         |        |          |             |
//     3    |  1   |    4    |   4    |     1    |      1      |    1
//          |      |         |        |          |             |
//     4    |  1   |    1    |   1    |     3    |      1      |    1
//          |      |         |        |          |             |
//     5    |  1   |    3    |   3    |     1    |      1      |    1
//          |      |         |        |          |             |
//     6    |  1   |    1    |   1    |     1    |      1      |    1
//          |      |         |        |          |             |
//     7    |  1   |    2    |   2    |     1    |      1      |    1
//          |      |         |        |          |             |
//     8    |  1   |    1    |   1    |     1    |      1      |    1
//          |      |         |        |          |             |
//     9    |  1   |    1    |   1    |     1    |      5      |    6
//          |      |         |        |          |             |
//    10    |  1   |    1    |   1    |  1 or 5  |      1      |    1
//          |      |         |        |          |             |
//    11    |  1   |    1    |   1    |     1    |      6      |    6
//          |      |         |        |          |             |
//    12    |  1   |    1    |   1    |     1    |      1      | 1 or 6
//          |      |         |        |          |             |
//    13    |  1   |    1    |   1    |     1    |      1      |    1
//          |      |         |        |          |             |
//          ---------------------------------------------------------------

type peer struct {
	*peerInfo
	// neighbor       *neighbor
	// addr           net.IP
	// port           int
	// link           netlink.Link
	// as             int
	// routerId       net.IP
	locRib         *LocRib // reference of bgp.rib. This field is called by every peers. When handling this, we must get lock.
	rib            *AdjRib
	pib            any // Policy Information Base
	holdTime       time.Duration
	capabilities   []Capability
	conn           *net.TCPConn
	connCh         chan *net.TCPConn
	state          state
	eventQueue     chan event
	tx             chan *Packet
	rx             chan *Packet
	logger         log.Logger
	connRetryTimer *timer
	keepAliveTimer *timer
	holdTimer      *timer
	initialized    bool
	listenerOpen   bool
}

type peerInfo struct {
	neighbor  *neighbor
	link      netlink.Link
	addr      net.IP
	port      int
	as        int
	routerId  net.IP
	timestamp time.Time
}

func (pi *peerInfo) isIBGP() bool {
	return pi.as == pi.neighbor.as
}

const (
	DEFAULT_CONNECT_RETRY_TIME_INTERVAL time.Duration = 120 * time.Second
	DEFAULT_KEEPALIVE_TIME_INTERVAL     time.Duration = 60 * time.Second
	DEFAULT_HOLD_TIME_INTERVAL          time.Duration = 180 * time.Second
)

func newPeer(logger log.Logger, link netlink.Link, local, addr, routerId net.IP, myAS, peerAS int, locRib *LocRib, adjRib *AdjRib) *peer {
	return &peer{
		peerInfo: &peerInfo{
			neighbor:  newNeighbor(addr, peerAS),
			addr:      local,
			link:      link,
			as:        myAS,
			routerId:  routerId,
			timestamp: time.Now(),
		},
		locRib:         locRib,
		rib:            adjRib,
		holdTime:       DEFAULT_HOLD_TIME_INTERVAL,
		capabilities:   defaultCaps(),
		state:          IDLE,
		eventQueue:     make(chan event, 1),
		tx:             make(chan *Packet, 128),
		rx:             make(chan *Packet, 128),
		connCh:         make(chan *net.TCPConn, 1),
		logger:         logger,
		connRetryTimer: newTimer(DEFAULT_CONNECT_RETRY_TIME_INTERVAL),
		keepAliveTimer: newTimer(DEFAULT_KEEPALIVE_TIME_INTERVAL),
		holdTimer:      newTimer(DEFAULT_HOLD_TIME_INTERVAL),
		initialized:    false,
		listenerOpen:   false,
	}
}

func (p *peer) logInfo(format string, v ...any) {
	p.logger.Infof("[%s:%d(%d) %s] %s", p.neighbor.addr, p.neighbor.port, p.neighbor.as, p.state, fmt.Sprintf(format, v...))
}

func (p *peer) logWarn(format string, v ...any) {
	p.logger.Warnf("[%s:%d(%d) %s] %s", p.neighbor.addr, p.neighbor.port, p.neighbor.as, p.state, fmt.Sprintf(format, v...))
}

func (p *peer) logErr(format string, v ...any) {
	p.logger.Errorf("[%s:%d(%d) %s] %s", p.neighbor.addr, p.neighbor.port, p.neighbor.as, p.state, fmt.Sprintf(format, v...))
}

func (p *peer) poll(ctx context.Context) {
	p.logInfo("polling start...")
	go func() {
		for {
			select {
			case <-p.connRetryTimer.tick().C:
				// p.logger.Infof("peer[%s] Connect Retry Timer is Expired.\n", p.neighbor.addr)
				p.enqueueEvent(&connRetryTimerExpired{})
			case <-p.keepAliveTimer.tick().C:
				// p.logger.Infof("peer[%s] Keep Alive Timer is Expired.\n", p.neighbor.addr)
				p.enqueueEvent(&keepaliveTimerExpired{})
			case <-p.holdTimer.tick().C:
				// p.logger.Infof("peer[%s] Hold Timer is Expired.\n", p.neighbor.addr)
				p.enqueueEvent(&holdTimerExpired{})
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		select {
		case evt := <-p.eventQueue:
			if err := p.handleEvent(ctx, evt); err != nil {
				p.logErr("handleEvent: evnt_type=%s %s", evt.typ(), err)
			}
		case <-ctx.Done():
			p.logInfo("polling stop...")
			return
		}
	}
}

func (p *peer) handleEvent(ctx context.Context, evt event) error {
	switch evt.typ() {
	case event_type_bgp_start:
		return p.startEvent()
	case event_type_bgp_stop:
		return p.stopEvent()
	case event_type_bgp_trans_conn_open:
		return p.transOpenEvent(ctx)
	case event_type_bgp_trans_conn_closed:
		return p.transClosedEvent()
	case event_type_bgp_trans_conn_open_failed:
		return p.transOpenFailedEvent()
	case event_type_bgp_trans_fatal_error:
		return p.transFatalErrorEvent()
	case event_type_conn_retry_timer_expired:
		return p.connRetryTimerExpiredEvent()
	case event_type_hold_timer_expired:
		return p.holdTimerExpiredEvent()
	case event_type_keepalive_timer_expired:
		return p.keepaliveTimerExpiredEvent()
	case event_type_recv_open_msg:
		return p.recvOpenMsgEvent(evt)
	case event_type_recv_keepalive_msg:
		return p.recvKeepAliveMsgEvent(evt)
	case event_type_recv_update_msg:
		return p.recvUpdateMsgEvent(evt)
	case event_type_recv_notification_msg:
		return p.recvNotificationMsgEvent(evt)
	case event_type_trigger_decision_process:
		return p.triggerDecisionProcessEvent(evt)
	default:
		return ErrInvalidEventType
	}
}

func (p *peer) enqueueEvent(evt event) error {
	if p.eventQueue == nil {
		return ErrEventQueueNotExist
	}
	p.eventQueue <- evt
	p.logInfo(" <- %s", evt.typ())
	return nil
}

func (p *peer) finishInitiate(conn *net.TCPConn) error {
	p.conn = conn
	// net.Conn must be *net.TCPConn
	_, lport, err := SplitAddrAndPort(conn.LocalAddr().String())
	if err != nil {
		return fmt.Errorf("failed to parse local addr and port: %w", err)
	}
	raddr, rport, err := SplitAddrAndPort(conn.RemoteAddr().String())
	if err != nil {
		return fmt.Errorf("failed to parse peer addr and port: %w", err)
	}
	if p.port == 0 {
		p.port = lport
	}
	if !raddr.Equal(p.neighbor.addr) {
		return fmt.Errorf("finishInitiate: %w", ErrInvalidNeighborAddress)
	}
	if p.neighbor.port == 0 {
		p.neighbor.SetPort(rport)
	}
	p.connRetryTimer.stop()
	return nil
}

func (p *peer) sendOpenMsg(options []*Option) error {
	builder := Builder(OPEN)
	builder.AS(p.as)
	builder.HoldTime(p.holdTime)
	builder.Identifier(p.addr) // TODO: Identifier is specified by peer. it shall be the largest address in the host.
	builder.Options(options)
	p.tx <- builder.Packet()
	return nil
}

func (p *peer) handleConn(ctx context.Context) error {
	// handle tcp connection
	if p.conn == nil {
		return fmt.Errorf("handleConn: %w", ErrTransportConnectionClose)
	}
	p.logInfo("establish TCP connection")
	childCtx, cancel := context.WithCancel(ctx)
	// TX handle
	go func() {
		p.logInfo("Tx handle start...")
		for {
			select {
			case msg := <-p.tx:
				data, err := msg.Decode()
				if err != nil {
					p.logErr("Tx handler: failed to decode packet: %s", err)
					continue
				}
				if _, err := p.conn.Write(data); err != nil {
					p.logErr("Tx handler: failed to write packet: %s", err)
					continue
				}
				if msg.Header.Type == NOTIFICATION {
					// close connection
					notification := GetMessage[*Notification](msg.Message)
					p.logWarn("send Notification msg: %s", notification.ErrorCode)
					p.conn.Close()
					cancel()
				}
			case <-childCtx.Done():
				p.logInfo("Tx handle stop...")
				return
			}
		}
	}()
	// RX handle
	go func() {
		p.logInfo("Rx handle start...")
		for {
			select {
			case <-childCtx.Done():
				p.logInfo("Rx handle stop...")
				return
			default:
				data := make([]byte, 1024*5)
				n, err := p.conn.Read(data)
				if err != nil {
					if err == io.EOF {
						cancel()
						break
					} else {
						cancel()
						p.enqueueEvent(&bgpTransFatalError{})
						p.logErr("Rx handler: failed to read %s", err)
					}
				}
				packets, err := preParse(data[:n])
				if err != nil {
					p.logErr("RX handler: falied to pre-parse packet %s", err)
					continue
				}
				for _, d := range packets {
					packet, err := Parse(d)
					if err != nil {
						p.logErr("RX handler: falied to parse packet %s\n%s", err, hex.Dump(d))
						continue
					}
					switch packet.Header.Type {
					case OPEN:
						p.enqueueEvent(&recvOpenMsg{msg: packet})
					case KEEPALIVE:
						p.enqueueEvent(&recvKeepaliveMsg{msg: packet})
					case UPDATE:
						p.enqueueEvent(&recvUpdateMsg{msg: packet})
					case NOTIFICATION:
						p.enqueueEvent(&recvNotificationMsg{msg: packet})
					default:
						p.logErr("Rx handler: %s(%d)", ErrInvalidMessageType, packet.Header.Type)
					}
				}
			}
		}
	}()
	return nil
}

func (p *peer) notifyError(d []byte, err *ErrorCode) error {
	builder := Builder(NOTIFICATION)
	builder.ErrorCode(err)
	if d != nil {
		builder.Data(d)
	}
	p.tx <- builder.Packet()
	return fmt.Errorf("notify: %w", err)
}

func (p *peer) changeState(et eventType) error {
	old := p.state
	switch p.state {
	case IDLE:
		if et == event_type_bgp_start {
			p.state = CONNECT
		}
	case CONNECT:
		switch et {
		case event_type_bgp_start, event_type_conn_retry_timer_expired:
			// stay
		case event_type_bgp_trans_conn_open_failed:
			p.state = ACTIVE
		case event_type_bgp_trans_conn_open:
			p.state = OPEN_SENT
		default:
			p.state = IDLE
		}
	case ACTIVE:
		switch et {
		case event_type_bgp_start, event_type_bgp_trans_conn_open_failed:
			// stay
		case event_type_bgp_trans_conn_open:
			p.state = OPEN_SENT
		case event_type_conn_retry_timer_expired:
			p.state = CONNECT
		default:
			p.state = IDLE
		}
	case OPEN_SENT:
		switch et {
		case event_type_bgp_start:
			// stay
		case event_type_bgp_trans_conn_closed:
			p.state = ACTIVE
		case event_type_recv_open_msg:
			// IDLE or OPEN_CONFIRM
			p.state = OPEN_CONFIRM
		default:
			p.state = IDLE
		}
	case OPEN_CONFIRM:
		switch et {
		case event_type_bgp_start, event_type_keepalive_timer_expired:
			// stay
		case event_type_recv_keepalive_msg:
			p.state = ESTABLISHED
		default:
			p.state = IDLE
		}
	case ESTABLISHED:
		switch et {
		case event_type_bgp_start, event_type_keepalive_timer_expired,
			event_type_recv_keepalive_msg:
			// stay
		case event_type_recv_update_msg:
			// IDLE or stay
		default:
			p.state = IDLE
		}
	default:
		return p.notifyError(nil, ErrFiniteStateMachineError)
	}
	if old != p.state {
		p.logInfo("%s -> %s", old, p.state)
	}
	return nil
}

func (p *peer) moveState(state state) error {
	old := p.state
	switch p.state {
	case IDLE:
		if state == IDLE || state == CONNECT {
			p.state = state
		}
	case CONNECT, ACTIVE:
		if state == IDLE || state == CONNECT || state == ACTIVE || state == OPEN_SENT {
			p.state = state
		}
	case OPEN_SENT:
		if state == IDLE || state == ACTIVE || state == OPEN_SENT || state == OPEN_CONFIRM {
			p.state = state
		}
	case OPEN_CONFIRM:
		if state == IDLE || state == OPEN_CONFIRM || state == ESTABLISHED {
			p.state = state
		}
	case ESTABLISHED:
		if state == IDLE || state == ESTABLISHED {
			p.state = state
		}
	default:
		return ErrInvalidBgpState
	}
	if old != p.state {
		p.logInfo("%s -> %s", old, p.state)
	}
	return nil
}

func (p *peer) isValidEvent(et eventType) (bool, error) {
	switch p.state {
	case IDLE:
		if et == event_type_bgp_start {
			return true, nil
		}
		return false, nil
	case CONNECT:
		switch et {
		case event_type_bgp_start, event_type_conn_retry_timer_expired:
			return false, nil
		default:
			return true, nil
		}
	case ACTIVE:
		switch et {
		case event_type_bgp_start, event_type_bgp_trans_conn_open_failed:
			return false, nil
		default:
			return true, nil
		}
	case OPEN_SENT:
		switch et {
		case event_type_bgp_start:
			return false, nil
		default:
			return true, nil
		}
	case OPEN_CONFIRM:
		switch et {
		case event_type_bgp_start, event_type_keepalive_timer_expired:
			return false, nil
		default:
			return true, nil
		}
	case ESTABLISHED:
		switch et {
		case event_type_bgp_start, event_type_keepalive_timer_expired, event_type_recv_update_msg:
			return true, nil
		default:
			return false, nil
		}
	default:
		return false, ErrInvalidBgpState
	}
}

func (p *peer) startEvent() error {
	switch p.state {
	case IDLE:
		if err := p.init(); err != nil {
			return fmt.Errorf("startEvent: %w", err)
		}
		p.enqueueEvent(&bgpTransConnOpen{})
	}
	return p.changeState(event_type_bgp_start)
}

func (p *peer) stopEvent() error {
	return nil
}

func (p *peer) transOpenEvent(ctx context.Context) error {
	switch p.state {
	case CONNECT:
		conn, err := net.DialTCP("tcp", &net.TCPAddr{IP: p.addr}, &net.TCPAddr{IP: p.neighbor.addr, Port: PORT})
		if err != nil {
			// restart connect retry timer
			p.logWarn("TCP peer is not open")
			p.connRetryTimer.restart()
			// wait for connecting from remote peer
			if !p.listenerOpen {
				l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: p.addr, Port: PORT})
				if err != nil {
					return fmt.Errorf("transOpenEvent: %w", err)
				}
				p.logInfo("start to listen connection from remote peer")
				p.listenerOpen = true
				go func() {
					conn, err := l.AcceptTCP()
					if err != nil {
						p.logErr("failed to accept TCP connection: %s", err)
					}
					p.connCh <- conn
					p.enqueueEvent(&bgpTransConnOpen{})
					l.Close()
					p.listenerOpen = false
				}()
			}
			// move to ACTIVE
			return p.moveState(ACTIVE)
		} else {
			// success to connect
			p.finishInitiate(conn)
			if err := p.handleConn(ctx); err != nil {
				return fmt.Errorf("tansOpenEvent: failed to handle conn: %w", err)
			}
			opts := make([]*Option, 0)
			// auth options
			for _, cap := range p.capabilities {
				b, err := cap.Decode()
				if err != nil {
					return nil
				}
				opts = append(opts, &Option{Type: CAPABILITY, Length: uint8(len(b)), Value: b})
			}
			if err := p.sendOpenMsg(opts); err != nil {
				return fmt.Errorf("transOpenEvent: failed to send OPEN message: %w", err)
			}
		}
	case ACTIVE:
		conn := <-p.connCh
		p.finishInitiate(conn)
		if err := p.handleConn(ctx); err != nil {
			return fmt.Errorf("tansOpenEvent: failed to handle conn: %w", err)
		}
		time.Sleep(time.Second * 1) // sleep for sync
		opts := make([]*Option, 0)
		// auth options
		for _, cap := range p.capabilities {
			b, err := cap.Decode()
			if err != nil {
				return nil
			}
			opts = append(opts, &Option{Type: CAPABILITY, Length: uint8(len(b)), Value: b})
		}
		if err := p.sendOpenMsg(opts); err != nil {
			return err
		}
	default:
		return fmt.Errorf("transOpenEvent: %w", ErrInvalidEventInCurrentState)
	}
	return p.changeState(event_type_bgp_trans_conn_open)
}

func (p *peer) transClosedEvent() error {
	switch p.state {
	case CONNECT:
		p.connRetryTimer.restart()
	case ACTIVE:
		p.conn.Close()
	default:
		return fmt.Errorf("transClosed: %w", ErrInvalidEventInCurrentState)
	}
	return p.changeState(event_type_bgp_trans_conn_closed)
}

func (p *peer) transOpenFailedEvent() error {
	switch p.state {
	case CONNECT:
		p.connRetryTimer.restart()
	case ACTIVE:
		if p.conn != nil {
			p.conn.Close()
		}
		p.connRetryTimer.restart()
	}
	return p.changeState(event_type_bgp_trans_conn_open_failed)
}

func (p *peer) transFatalErrorEvent() error {
	if p.conn != nil {
		p.conn.Close()
	}
	return p.changeState(event_type_bgp_trans_fatal_error)
}

func (p *peer) connRetryTimerExpiredEvent() error {
	switch p.state {
	case CONNECT, ACTIVE:
		p.connRetryTimer.restart()
		p.enqueueEvent(&bgpTransConnOpen{})
	default:
		return fmt.Errorf("connRetryTimeExpired: %w", ErrInvalidEventInCurrentState)
	}
	return p.changeState(event_type_conn_retry_timer_expired)
}

func (p *peer) holdTimerExpiredEvent() error {
	builder := Builder(NOTIFICATION)
	builder.ErrorCode(&ErrorCode{Code: HOLD_TIMER_EXPIRED, Subcode: 0})
	p.tx <- builder.Packet()
	return p.changeState(event_type_hold_timer_expired)
}

func (p *peer) keepaliveTimerExpiredEvent() error {
	switch p.state {
	case OPEN_CONFIRM, ESTABLISHED:
		// send KEEPALIVE
		p.keepAliveTimer.restart()
		builder := Builder(KEEPALIVE)
		p.tx <- builder.Packet()
		p.logInfo("send keepalive")
	}
	return nil
}

func (p *peer) recvOpenMsgEvent(evt event) error {
	switch p.state {
	case OPEN_SENT:
		msg := evt.(*recvOpenMsg)
		op := GetMessage[*Open](msg.msg.Message)
		if err := op.Validate(); err != nil {
			// OPEN failed
			p.logErr("recvOpenMsg: Open failed: %s", err)
			return p.notifyError(nil, err)
		}
		caps, _ := op.Capabilities() // ignore unsupported capability error
		p.neighbor.capabilities = append(p.neighbor.capabilities, caps...)
		p.neighbor.SetRouterId(op.Identifier)
		// Process OPEN is ok
		p.keepAliveTimer.start()
		builder := Builder(KEEPALIVE)
		p.tx <- builder.Packet()
	default:
		// Close transport connection
		// Release resources
		p.release()
		// Send NOTIFICATION
		builder := Builder(NOTIFICATION)
		builder.ErrorCode(ErrFiniteStateMachineError)
		p.tx <- builder.Packet()
	}
	return p.changeState(event_type_recv_open_msg)
}

func (p *peer) recvKeepAliveMsgEvent(evt event) error {
	switch p.state {
	case OPEN_CONFIRM:
		// Complete initialization
		// Restart hold timer
		p.holdTimer.restart()
	case ESTABLISHED:
		p.holdTimer.restart()
	default:

	}
	return p.changeState(event_type_recv_keepalive_msg)
}

func (p *peer) recvUpdateMsgEvent(evt event) error {
	msg := evt.(*recvUpdateMsg)
	update := GetMessage[*Update](msg.msg.Message)
	d, err := update.Validate(msg.msg.Header.Length)
	if err != nil {
		p.logErr("recvUpdateMsg: validate error: %s", err)
		return p.notifyError(d, err)
	}
	p.holdTimer.restart()
	if err := p.handleUpdateMsg(update); err != nil {
		return err
	}
	return p.changeState(event_type_recv_update_msg)
}

func (p *peer) recvNotificationMsgEvent(evt event) error {
	msg := evt.(*recvNotificationMsg)
	p.conn.Close()
	p.logErr("%s", GetMessage[*Notification](msg.msg.Message).ErrorCode)
	p.release()
	return p.changeState(event_type_recv_notification_msg)
}

func (p *peer) triggerDecisionProcessEvent(evt event) error {
	arg := evt.(*triggerDecisionProcess)
	withdrawn := arg.withdrawn
	for _, path := range arg.pathes {
		if err := p.Decide(path, withdrawn); err != nil {
			return fmt.Errorf("triggerDecisionProcessEvent: %w", err)
		}
	}
	return nil
}

// initialize resources
func (p *peer) init() error {
	p.initialized = true
	p.connRetryTimer.start()
	return nil
}

// release resources
func (p *peer) release() error {
	p.keepAliveTimer.stop()
	p.holdTimer.stop()
	return nil
}

type timer struct {
	ticker   *time.Ticker
	interval time.Duration
	flag     bool
}

func newTimer(interval time.Duration) *timer {
	t := &timer{
		ticker:   time.NewTicker(interval),
		interval: interval,
		flag:     false,
	}
	t.ticker.Stop()
	return t
}

func (t *timer) start() {
	t.ticker.Reset(t.interval)
	t.flag = true
}

func (t *timer) restart() {
	t.ticker.Reset(t.interval)
	t.flag = true
}

func (t *timer) stop() {
	t.ticker.Stop()
	t.flag = false
}

func (t *timer) reset(newInterval time.Duration) {
	t.interval = newInterval
	t.ticker.Reset(newInterval)
}

func (t *timer) isStopped() bool {
	return !t.flag
}

func (t *timer) tick() *time.Ticker {
	return t.ticker
}

// UPDATE message handling
// https://datatracker.ietf.org/doc/html/rfc4271#section-9
func (p *peer) handleUpdateMsg(msg *Update) error {
	propagateAttrs := make([]PathAttr, 0)
	recognizedAttrs := make([]PathAttr, 0)
	var (
		next            net.IP
		asPath          ASPath
		origin          Origin
		med             int
		feasiblePathes  []*Path = make([]*Path, 0)
		withdrawnPathes []*Path = make([]*Path, 0)
	)
	for _, attr := range msg.PathAttrs {
		if attr.IsRecognized() {
			recognizedAttrs = append(recognizedAttrs, attr)
		}
		// If an optional non-transitive attribute is unrecognized, it is quietly ignored.
		if attr.IsOptional() && !attr.IsTransitive() {
			continue
		}
		// If an optional transitive attribute is unrecognized, the Partial bit in the attribute flags is set to 1,
		// and the attribute is retained for propagation to other BGP speakers.
		if attr.IsOptional() && attr.IsTransitive() && !attr.IsRecognized() {
			attr.SetPartial()
			propagateAttrs = append(propagateAttrs, attr)
		}

		// If an optional attribute is recognized, and has a valid value,
		// then, depending on the type of the optional attruibute, it is processed locally, ratained, and updated,
		// if necessary, for possible propagation to other BGP speakers.
		if attr.IsOptional() && attr.IsRecognized() {
			// handle attribute
		}

		switch attr.Type() {
		case ORIGIN:
			origin = *GetPathAttr[*Origin](attr)
		case AS_PATH:
			asPath = *GetPathAttr[*ASPath](attr)
		case NEXT_HOP:
			nextHop := GetPathAttr[*NextHop](attr)
			next = nextHop.next
		case MULTI_EXIT_DISC:
			med = int(GetPathAttr[*MultiExitDisc](attr).discriminator)
		}
	}
	// If the UPDATE message contains a non-empty Withdrawn routes field,
	// the previously advertised routes whose destinations (expressed as IP prefixes) are contained in this field, shall be removed from the Adj-RIB-In.
	// This BGP speaker shall run its Decision Process because the previously advertised route is no longer available for use.
	withdrawn := false
	for _, withdrawnRoute := range msg.WithdrawnRoutes {
		// handle withdrawn routes
		withdrawn = true
		withdrawnPathes = append(withdrawnPathes, &Path{nlri: withdrawnRoute})
		p.logInfo("withdrawing route %s", withdrawnRoute)
	}
	if withdrawn {
		p.enqueueEvent(&triggerDecisionProcess{pathes: withdrawnPathes, withdrawn: withdrawn})
	}

	// If the UPDATE message contains a feasible route, the Adj-RIB-In will be updated with this route as follows:
	adjRibInUpdate := false
	for _, feasibleRoute := range msg.NetworkLayerReachabilityInfo {
		// handle network layer reachability information
		p.logInfo("feasible route %s", feasibleRoute)
		// if the NLRI of the new route is idnetical to the one the route currently has stored in the Adj-RIB-In,
		// then the new route shall replace the older route in the Adj-RIB-In,
		// thus implicitly withdrawing the older route from service.

		// Otherwise, if the Adj-RIB-In has no route with NLRI identical to the new route,
		// the new route shall be placed in the Adj-RIB-In.
		path := newPath(p.peerInfo, p.neighbor.as, next, origin, asPath, med, feasibleRoute, recognizedAttrs, p.link)
		feasiblePathes = append(feasiblePathes, path)
		// p.rib.In.Insert(path)
		adjRibInUpdate = true
		// p.logInfo("Insert into Adg-RIB-In: %s", path)
		// Once the BGP speaker updates the Adj-RIB-In, the speaker shall run its Decision Process.
	}
	// trigger decision process event
	if adjRibInUpdate {
		p.enqueueEvent(&triggerDecisionProcess{pathes: feasiblePathes, withdrawn: false})
	}
	return nil
}
