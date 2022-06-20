package bgp

// BGP Events
// 1.  BGP Start
// 2.  BGP Stop
// 3.  BGP Transport connection open
// 4.  BGP Transport connection closed
// 5.  BGP Transport connection open failed
// 6.  BGP Transport fatal error
// 7.  ConnectRetry timer expired
// 8.  Hold Timer expired
// 9.  KeepAlive timer expired
// 10. Receive OPEN message
// 11. Receive KEEPALIVE message
// 12. Receive UPDATE messages
// 13. Receive NOTIFICATION message
//
// Original event
// 14. Trigger Decision Process event
type event interface {
	typ() eventType
}

type eventType uint8

const (
	event_type_bgp_start                  eventType = iota + 1
	event_type_bgp_stop                   eventType = iota + 1
	event_type_bgp_trans_conn_open        eventType = iota + 1
	event_type_bgp_trans_conn_closed      eventType = iota + 1
	event_type_bgp_trans_conn_open_failed eventType = iota + 1
	event_type_bgp_trans_fatal_error      eventType = iota + 1
	event_type_conn_retry_timer_expired   eventType = iota + 1
	event_type_hold_timer_expired         eventType = iota + 1
	event_type_keepalive_timer_expired    eventType = iota + 1
	event_type_recv_open_msg              eventType = iota + 1
	event_type_recv_keepalive_msg         eventType = iota + 1
	event_type_recv_update_msg            eventType = iota + 1
	event_type_recv_notification_msg      eventType = iota + 1

	event_type_trigger_decision_process eventType = iota + 1
	event_type_trigger_dissemination    eventType = iota + 1
)

func (t eventType) String() string {
	switch t {
	case event_type_bgp_start:
		return "BGP start"
	case event_type_bgp_stop:
		return "BGP stop"
	case event_type_bgp_trans_conn_open:
		return "BGP transport connection open"
	case event_type_bgp_trans_conn_closed:
		return "BGP transport connection closed"
	case event_type_bgp_trans_conn_open_failed:
		return "BGP transport connection open failed"
	case event_type_bgp_trans_fatal_error:
		return "BGP transport fatal error"
	case event_type_conn_retry_timer_expired:
		return "BGP conn_retry timer expired"
	case event_type_hold_timer_expired:
		return "Hold timer expired"
	case event_type_keepalive_timer_expired:
		return "Keepalive timer expired"
	case event_type_recv_open_msg:
		return "Receive open message"
	case event_type_recv_keepalive_msg:
		return "Receive keepalive message"
	case event_type_recv_update_msg:
		return "Receive update message"
	case event_type_recv_notification_msg:
		return "Receive notification message"
	case event_type_trigger_decision_process:
		return "Trigger decision process"
	case event_type_trigger_dissemination:
		return "Trigger dissemination"
	default:
		return "Unknown event type"
	}
}

type bgpStart struct{}

type bgpStop struct{}

type bgpTransConnOpen struct{}

type bgpTransConnClosed struct{}

type bgpTransConnOpenFailed struct{}

type bgpTransFatalError struct{}

type connRetryTimerExpired struct{}

type holdTimerExpired struct{}

type keepaliveTimerExpired struct{}

type recvOpenMsg struct {
	msg *Packet
}

type recvKeepaliveMsg struct {
	msg *Packet
}

type recvUpdateMsg struct {
	msg *Packet
}

type recvNotificationMsg struct {
	msg *Packet
}

type triggerDecisionProcess struct {
	pathes    []*Path
	withdrawn bool
}

type triggerDissemination struct{}

func (*bgpStart) typ() eventType {
	return event_type_bgp_start
}

func (*bgpStop) typ() eventType {
	return event_type_bgp_stop
}

func (*bgpTransConnOpen) typ() eventType {
	return event_type_bgp_trans_conn_open
}

func (*bgpTransConnClosed) typ() eventType {
	return event_type_bgp_trans_conn_closed
}

func (*bgpTransConnOpenFailed) typ() eventType {
	return event_type_bgp_trans_conn_open_failed
}

func (*bgpTransFatalError) typ() eventType {
	return event_type_bgp_trans_fatal_error
}

func (*connRetryTimerExpired) typ() eventType {
	return event_type_conn_retry_timer_expired
}

func (*holdTimerExpired) typ() eventType {
	return event_type_hold_timer_expired
}

func (*keepaliveTimerExpired) typ() eventType {
	return event_type_keepalive_timer_expired
}

func (*recvOpenMsg) typ() eventType {
	return event_type_recv_open_msg
}

func (*recvKeepaliveMsg) typ() eventType {
	return event_type_recv_keepalive_msg
}

func (*recvUpdateMsg) typ() eventType {
	return event_type_recv_update_msg
}

func (*recvNotificationMsg) typ() eventType {
	return event_type_recv_notification_msg
}

func (*triggerDecisionProcess) typ() eventType {
	return event_type_trigger_decision_process
}

func (*triggerDissemination) typ() eventType {
	return event_type_trigger_dissemination
}
