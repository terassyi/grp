package bgp

import (
	"context"
	"io"
	"net"
	"runtime"
	"time"

	"github.com/terassyi/grp/grp/log"
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
	local          net.IP
	addr           net.IP
	link           netlink.Link
	as             int
	holdTime       time.Duration
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

const (
	DEFAULT_CONNECT_RETRY_TIME_INTERVAL time.Duration = 120 * time.Second
	DEFAULT_KEEPALIVE_TIME_INTERVAL     time.Duration = 60 * time.Second
	DEFAULT_HOLD_TIME_INTERVAL          time.Duration = 180 * time.Second
)

func newPeer(logger log.Logger, link netlink.Link, local, addr net.IP, as int) *peer {
	return &peer{
		local:          local,
		addr:           addr,
		link:           link,
		as:             as,
		holdTime:       DEFAULT_HOLD_TIME_INTERVAL,
		state:          IDLE,
		eventQueue:     make(chan event, 1),
		tx:             make(chan *Packet, 128),
		rx:             make(chan *Packet, 128),
		logger:         logger,
		connRetryTimer: newTimer(DEFAULT_CONNECT_RETRY_TIME_INTERVAL),
		keepAliveTimer: newTimer(DEFAULT_KEEPALIVE_TIME_INTERVAL),
		holdTimer:      newTimer(DEFAULT_HOLD_TIME_INTERVAL),
		initialized:    false,
		listenerOpen:   false,
	}
}

func (p *peer) poll(ctx context.Context) {
	p.logger.Infof("Peer polling start.")
	go func() {
		for {
			select {
			case <-p.connRetryTimer.tick().C:
				p.logger.Infof("peer[%s] Connect Retry Timer is Expired.\n", p.addr)
				p.enqueueEvent(&connRetryTimerExpired{})
			case <-p.keepAliveTimer.tick().C:
				p.logger.Infof("peer[%s] Hold Timer is Expired.\n", p.addr)
			case <-p.holdTimer.tick().C:
				p.logger.Infof("peer[%s] Keep Alive Timer is Expired.\n", p.addr)
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		select {
		case evt := <-p.eventQueue:
			if err := p.handleEvent(ctx, evt); err != nil {
				p.logger.Errorf("Event handler(%s): %s", evt.typ(), err)
			}
		case <-ctx.Done():
			p.logger.Infof("Stop polling.")
			return
		}
		p.logger.Warnf("Goroutines %d", runtime.NumGoroutine())
	}
}

func (p *peer) handleEvent(ctx context.Context, evt event) error {
	p.logger.Infoln(evt.typ().String())
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
	default:
		return ErrInvalidEventType
	}
}

func (p *peer) enqueueEvent(evt event) error {
	if p.eventQueue == nil {
		return ErrEventQueueNotExist
	}
	p.eventQueue <- evt
	p.logger.Infof("enqueue %s", evt.typ())
	return nil
}

func (p *peer) finishInitiate(conn *net.TCPConn) error {
	p.conn = conn
	p.connRetryTimer.stop()
	return nil
}

func (p *peer) sendOpenMsg(options []*Option) error {
	builder := Builder(OPEN)
	builder.AS(p.as)
	builder.HoldTime(p.holdTime)
	builder.Identifier(p.local)
	builder.Options(options)
	p.tx <- builder.Packet()
	return nil
}

func (p *peer) handleConn(ctx context.Context) error {
	// handle tcp connection
	if p.conn == nil {
		return ErrTransportConnectionClose
	}
	p.logger.Infof("transport connection is established.")
	childCtx, cancel := context.WithCancel(ctx)
	// TX handle
	go func() {
		p.logger.Infof("Start TX handle.")
		for {
			select {
			case msg := <-p.tx:
				data, err := msg.Decode()
				if err != nil {
					p.logger.Errorf("TX handler: decoding packet %s", err)
					continue
				}
				if _, err := p.conn.Write(data); err != nil {
					p.logger.Errorf("TX handler: write to transport connection %s", err)
					continue
				}
			case <-childCtx.Done():
				p.logger.Infof("Stop TX handle.")
				return
			}
		}
	}()
	// RX handle
	go func() {
		p.logger.Infof("Start RX handle.")
		for {
			select {
			case <-childCtx.Done():
				p.logger.Infof("Stop RX handler.")
				return
			default:
				data := make([]byte, 512)
				n, err := p.conn.Read(data)
				if err != nil {
					if err == io.EOF {
						cancel()
						break
					} else {
						cancel()
						p.enqueueEvent(&bgpTransFatalError{})
						p.logger.Errorf("RX handler: read from transport connection %s", err, n)
					}
				}
				packet, err := Parse(data[:n])
				if err != nil {
					p.logger.Errorf("RX handler: parse packet %s", err)
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
					p.logger.Errorf("RX handler: %s(%s)", ErrInvalidMessageType, packet.Header.Type)
				}
			}
		}
	}()
	return nil
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
		return ErrInvalidBgpState
	}
	p.logger.Infof("change state %s -> %s\n", old, p.state)
	return nil
}

func (p *peer) moveState(state state) error {
	old := p.state
	switch p.state {
	case IDLE:
		if state == IDLE || state == CONNECT {
			p.state = state
			p.logger.Infof("move state %s -> %s\n", old, p.state)
			return nil
		}
	case CONNECT, ACTIVE:
		if state == IDLE || state == CONNECT || state == ACTIVE || state == OPEN_SENT {
			p.state = state
			p.logger.Infof("move state %s -> %s\n", old, p.state)
			return nil
		}
	case OPEN_SENT:
		if state == IDLE || state == ACTIVE || state == OPEN_SENT || state == OPEN_CONFIRM {
			p.state = state
			p.logger.Infof("move state %s -> %s\n", old, p.state)
			return nil
		}
	case OPEN_CONFIRM:
		if state == IDLE || state == OPEN_CONFIRM || state == ESTABLISHED {
			p.state = state
			p.logger.Infof("move state %s -> %s\n", old, p.state)
			return nil
		}
	case ESTABLISHED:
		if state == IDLE || state == ESTABLISHED {
			p.state = state
			p.logger.Infof("move state %s -> %s\n", old, p.state)
			return nil
		}
	default:
		return ErrInvalidBgpState
	}
	return ErrUnreachableState
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
			return err
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
		conn, err := net.DialTCP("tcp", &net.TCPAddr{IP: p.local}, &net.TCPAddr{IP: p.addr, Port: PORT})
		if err != nil {
			// restart connect retry timer
			p.logger.Infoln("peer connection is not open.")
			p.connRetryTimer.restart()
			// wait for connecting from remote peer
			if !p.listenerOpen {
				l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: p.local, Port: PORT})
				if err != nil {
					return err
				}
				p.logger.Infoln("start to listen connection from remote peer.")
				p.listenerOpen = true
				go func() {
					conn, err := l.AcceptTCP()
					if err != nil {
						p.logger.Errorf("Accept transport connection: %s", err)
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
				return err
			}
			time.Sleep(time.Second * 2)
			if err := p.sendOpenMsg(nil); err != nil {
				return err
			}
		}
	case ACTIVE:
		conn := <-p.connCh
		p.finishInitiate(conn)
		if err := p.handleConn(ctx); err != nil {
			return err
		}
		if err := p.sendOpenMsg(nil); err != nil {
			return err
		}
	default:
		return ErrInvalidEventInCurrentState
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
		return ErrInvalidEventInCurrentState
	}
	return p.changeState(event_type_bgp_trans_conn_closed)
}

func (p *peer) transOpenFailedEvent() error {
	return nil
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
		return ErrInvalidEventInCurrentState
	}
	return p.changeState(event_type_conn_retry_timer_expired)
}

func (p *peer) holdTimerExpiredEvent() error {
	return nil
}

func (p *peer) keepaliveTimerExpiredEvent() error {
	switch p.state {
	case OPEN_CONFIRM, ESTABLISHED:
		// send KEEPALIVE
		p.keepAliveTimer.restart()
		builder := Builder(KEEPALIVE)
		p.rx <- builder.Packet()
	}
	return nil
}

func (p *peer) recvOpenMsgEvent(evt event) error {
	switch p.state {
	case OPEN_SENT:
		// Process OPEN is ok
		//
	default:
		// Close transport connection
		// Release resources
		// Send NOTIFICATION
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
		builder := Builder(KEEPALIVE)
		p.rx <- builder.Packet()
	default:

	}
	return p.changeState(event_type_recv_keepalive_msg)
}

func (p *peer) recvUpdateMsgEvent(evt event) error {
	return nil
}

func (p *peer) recvNotificationMsgEvent(evt event) error {
	msg := evt.(*recvNotificationMsg)
	p.conn.Close()
	p.logger.Errorln(GetMessage[*Notification](msg.msg.Message).ErrorCode)
	return p.changeState(event_type_recv_notification_msg)
}

// initialize resources
func (p *peer) init() error {
	p.initialized = true
	p.connRetryTimer.start()
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
