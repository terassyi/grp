package bgp

import "sync"

type FSM interface {
	GetState() state
	SetState(state) error
	Change(eventType) error
	IsEstablished() bool
}

type fsm struct {
	mutex sync.RWMutex
	state state
}

func newFSM() FSM {
	return &fsm{state: IDLE}
}

func (f *fsm) GetState() state {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.state
}

func (f *fsm) SetState(s state) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.moveState(s)
}

func (f *fsm) Change(et eventType) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.changeState(et)
}

func (f *fsm) IsEstablished() bool {
	return f.GetState() == ESTABLISHED
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

func (f *fsm) changeState(et eventType) error {
	switch f.state {
	case IDLE:
		if et == event_type_bgp_start {
			f.state = CONNECT
		}
	case CONNECT:
		switch et {
		case event_type_bgp_start, event_type_conn_retry_timer_expired:
			// stay
		case event_type_bgp_trans_conn_open_failed:
			f.state = ACTIVE
		case event_type_bgp_trans_conn_open:
			f.state = OPEN_SENT
		default:
			f.state = IDLE
		}
	case ACTIVE:
		switch et {
		case event_type_bgp_start, event_type_bgp_trans_conn_open_failed:
			// stay
		case event_type_bgp_trans_conn_open:
			f.state = OPEN_SENT
		case event_type_conn_retry_timer_expired:
			f.state = CONNECT
		default:
			f.state = IDLE
		}
	case OPEN_SENT:
		switch et {
		case event_type_bgp_start:
			// stay
		case event_type_bgp_trans_conn_closed:
			f.state = ACTIVE
		case event_type_recv_open_msg:
			// IDLE or OPEN_CONFIRM
			f.state = OPEN_CONFIRM
		default:
			f.state = IDLE
		}
	case OPEN_CONFIRM:
		switch et {
		case event_type_bgp_start, event_type_keepalive_timer_expired:
			// stay
		case event_type_recv_keepalive_msg:
			f.state = ESTABLISHED
		default:
			f.state = IDLE
		}
	case ESTABLISHED:
		switch et {
		case event_type_bgp_start, event_type_keepalive_timer_expired,
			event_type_recv_keepalive_msg:
			// stay
		case event_type_recv_update_msg:
			// IDLE or stay
		default:
			f.state = IDLE
		}
	default:
		return ErrFiniteStateMachineError
	}
	return nil
}

func (f *fsm) moveState(state state) error {
	switch f.state {
	case IDLE:
		if state == IDLE || state == CONNECT {
			f.state = state
		}
	case CONNECT, ACTIVE:
		if state == IDLE || state == CONNECT || state == ACTIVE || state == OPEN_SENT {
			f.state = state
		}
	case OPEN_SENT:
		if state == IDLE || state == ACTIVE || state == OPEN_SENT || state == OPEN_CONFIRM {
			f.state = state
		}
	case OPEN_CONFIRM:
		if state == IDLE || state == OPEN_CONFIRM || state == ESTABLISHED {
			f.state = state
		}
	case ESTABLISHED:
		if state == IDLE || state == ESTABLISHED {
			f.state = state
		}
	default:
		return ErrInvalidBgpState
	}
	return nil
}
