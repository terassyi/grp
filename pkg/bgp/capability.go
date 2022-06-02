package bgp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// RFC 3392: Capabilities Advertisement with BGP-4
type Capability interface {
	Code() CapabilityCode
	String() string
	Decode() ([]byte, error)
}

func ParseCap(data []byte) (Capability, error) {
	type cap struct {
		code   CapabilityCode
		length uint8
		val    []byte
	}
	c := cap{code: CapabilityCode(data[0]), length: data[1], val: data[2:]}
	switch c.code {
	case MULTI_PROTOCOL_EXTENSIONS:
	default:
		return nil, fmt.Errorf("Unsupported Capability")
	}
	return nil, nil
}

func GetCap[T *MultiProtocolExtensions](cap Capability) T {
	return cap.(T)
}

// type capability struct {
// 	Code   CapabilityCode
// 	Length uint8
// 	Value  []byte
// }

type CapabilityCode uint8

// https://www.iana.org/assignments/capability-codes/capability-codes.xhtml
const (
	MULTI_PROTOCOL_EXTENSIONS   CapabilityCode = 1  // RFC 2858
	GRACEFUL_RESTART_CAPABILITY CapabilityCode = 64 // RFC 4724
)

// RFC 4760: Multi protocol extensions for BGP-4
// https://datatracker.ietf.org/doc/html/rfc4760
type MultiProtocolExtensions struct {
	AFI uint16
	// Reserved 8bit value should be 0
	SAFI uint8
}

const (
	AFI_IPv4 uint16 = 1
	AFI_IPv6 uint16 = 2

	SAFI_UNICAST               uint8 = 1
	SAFI_MULTICAST             uint8 = 2
	SAFI_UNICAST_AND_MULTICAST uint8 = 3
	SAFI_MLPS_LABEL            uint8 = 4
	SAFI_MLPS_LABELED_VPN      uint8 = 128
)

func newMultiProtocolExtensions(data []byte) *MultiProtocolExtensions {
	m := &MultiProtocolExtensions{}
	m.AFI = binary.BigEndian.Uint16(data[:2])
	m.SAFI = data[3]
	return m
}

func (*MultiProtocolExtensions) Code() CapabilityCode {
	return MULTI_PROTOCOL_EXTENSIONS
}

func (m *MultiProtocolExtensions) String() string {
	return fmt.Sprintf("") // TODO:
}

func (m *MultiProtocolExtensions) Decode() ([]byte, error) {
	buf := []byte{byte(m.Code()), 0x04}
	b := make([]byte, 2, 4)
	binary.BigEndian.PutUint16(b, m.AFI)
	b = append(b, 0x00)
	b = append(b, m.SAFI)
	return append(buf, b...), nil
}

// https://datatracker.ietf.org/doc/html/rfc8538
// RFC 4724: Graceful Restart Mechanism for BGP
// RFC 8538: Notification Message Support for BGP Graceful Restart
type GracefulRestartCapability struct {
	RestartStateFlag     bool
	GracefulNotification bool
	RestartTime          uint16 // 12bit
	Tuples               []AFITuple
}

type AFITuple struct {
	AFI  uint16
	SAFI uint8
	Flag uint8
}

func newGracefulRestartCapability(data []byte) (*GracefulRestartCapability, error) {
	g := &GracefulRestartCapability{}
	a := data[0]
	if 0b1000_0000&a == 0b1000_0000 {
		g.RestartStateFlag = true
	}
	if 0b0100_0000&a == 0b0100_0000 {
		g.GracefulNotification = true
	}
	g.RestartTime = uint16(data[1]) + uint16((a&0b0000_1111)<<4)
	if len(data) == 2 {
		return g, nil
	}
	buf := bytes.NewBuffer(data[2:])
	g.Tuples = make([]AFITuple, 0)
	for buf.Len() > 0 {
		t := &AFITuple{}
		if err := binary.Read(buf, binary.BigEndian, t); err != nil {
			return nil, err
		}
		g.Tuples = append(g.Tuples, *t)
	}
	return g, nil
}

func (GracefulRestartCapability) Code() CapabilityCode {
	return GRACEFUL_RESTART_CAPABILITY
}

func (g *GracefulRestartCapability) String() string {
	return ""
}

func (g *GracefulRestartCapability) Decode() ([]byte, error) {
	var a uint8 = 0
	if g.RestartStateFlag {
		a += 0b1000_0000
	}
	if g.GracefulNotification {
		a += 0b0100_0000
	}
	b := uint8(g.RestartTime)
	a += uint8(g.RestartTime >> 8)
	buf := bytes.NewBuffer([]byte{a, b})
	for _, t := range g.Tuples {
		if err := binary.Write(buf, binary.BigEndian, &t); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func defaultCaps() []Capability {
	return []Capability{&MultiProtocolExtensions{AFI: AFI_IPv4, SAFI: SAFI_UNICAST}}
}
