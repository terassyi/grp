package bgp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

var (
	ErrUnknownPathAttribute error = errors.New("Unkown Path Attribute")
)

type PathAttr interface {
	String() string
	Flags() uint8
	Type() PathAttrType
	ValueLen() int
	IsTransitive() bool
	IsOptional() bool
	IsRecognized() bool
	SetPartial()
	Decode() ([]byte, error)
}

func ParsePathAttrs(buf *bytes.Buffer) ([]PathAttr, error) {
	attrs := make([]PathAttr, 0)
	for buf.Len() > 0 {
		attr, err := parsePathAttr(buf)
		if err != nil {
			return nil, fmt.Errorf("ParsePathAttrs: %w", err)
		}
		attrs = append(attrs, attr)
	}
	return attrs, nil
}

func parsePathAttr(buf *bytes.Buffer) (PathAttr, error) {
	if buf.Len() < 2 {
		return nil, fmt.Errorf("Invalid Path Attribute data")
	}
	base := &pathAttr{}
	b, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("parsePathAttr: flags: %w", err)
	}
	base.flags = b
	b, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("parsePathAttr: type: %w", err)
	}
	base.typ = PathAttrType(b)
	var attr PathAttr
	switch base.typ {
	case AS_PATH:
		attr, err = newASPath(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: AS path: %w", err)
		}
	case ORIGIN:
		attr, err = newOrigin(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: Origin: %w", err)
		}
	case NEXT_HOP:
		attr, err = newNextHop(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: Next hop: %w", err)
		}
	case LOCAL_PREF:
		attr, err = newLocalPref(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: Local pref: %w", err)
		}
	case ATOMIC_AGGREGATE:
		attr, err = newAtomicAggrerate(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: Atomic aggregate: %w", err)
		}
	case AGGREGATOR:
		attr, err = newAggregator(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePatAttr: Aggregator: %w", err)
		}
	case MULTI_EXIT_DISC:
		attr, err = newMultiExitDisc(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: Multi exit disc: %w", err)
		}
	case COMMUNITIES:
		attr, err = newCommunities(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: Communities: %w", err)
		}
	case EXTENDED_COMMUNITIES:
		attr, err = newExtendedCommunities(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: Extended communities: %w", err)
		}
	default:
		attr, err = newUnimplementedPathAttr(buf, base)
		if err != nil {
			return nil, fmt.Errorf("parsePathAttr: Unimplemented path attribute: %w", err)
		}
	}
	return attr, nil
}

func GetPathAttr[T *ASPath | *NextHop | *Origin | *MultiExitDisc](attr PathAttr) T {
	return attr.(T)
}

type pathAttr struct {
	flags uint8
	typ   PathAttrType
}

func (p *pathAttr) String() string {
	return fmt.Sprintf("Flag=0x%x Type=%s", p.flags, p.typ)
}

const (
	// It defines whether the attribute is optional(if set to 1) or well-known(if set to 0)
	PATH_ATTR_FLAG_OPTIONAL uint8 = 1 << 7
	// It defines whether an optional attribute is transitive(if set to 1) or non-transitive(if set to 0)
	// For well-known attributes, the Transitive bit MUST be set to 1.
	PATH_ATTR_FLAG_TRANSITIVE uint8 = 1 << 6
	// It defines whether the information contained in the optional transitive attribute is partial(if set to 1) or complete(if set to 0).
	// For well-known attributes and for optional non-transitive attributes, the Partial bit MUST be set to 0.
	PATH_ATTR_FLAG_PARTIAL uint8 = 1 << 5
	// It defines whether the Attribute Length is one byte(if set to 0) or two bytes(if set to 1).
	PATH_ATTR_FLAG_EXTENDED uint8 = 1 << 4
)

func (p *pathAttr) IsTransitive() bool {
	return (p.flags & PATH_ATTR_FLAG_TRANSITIVE) == PATH_ATTR_FLAG_TRANSITIVE
}

func (p *pathAttr) IsOptional() bool {
	return (p.flags & PATH_ATTR_FLAG_OPTIONAL) == PATH_ATTR_FLAG_OPTIONAL
}

func (p *pathAttr) SetPartial() {
	p.flags |= PATH_ATTR_FLAG_PARTIAL
}

type PathAttrType uint8

const (
	ORIGIN               PathAttrType = 1  // Well-known mandatory attribute
	AS_PATH              PathAttrType = 2  // Well-known mandatory attribute
	NEXT_HOP             PathAttrType = 3  // Well-known mandatory attribute
	MULTI_EXIT_DISC      PathAttrType = 4  // Optional non-transitive attribute
	LOCAL_PREF           PathAttrType = 5  // Well-known discretionary attribute
	ATOMIC_AGGREGATE     PathAttrType = 6  // Well-known discretionary attribute
	AGGREGATOR           PathAttrType = 7  // Optional transitive attribute
	COMMUNITIES          PathAttrType = 8  // Optional transitive attribute
	EXTENDED_COMMUNITIES PathAttrType = 16 // Optional transitive attribute
	AS4_PATH             PathAttrType = 17 // Optional transitive attribute
	AS4_AGGREGATOR       PathAttrType = 18 // Optional transitive attribute
	LARGE_COMMUNITY      PathAttrType = 32 // Optional transitive attribute
)

func (p PathAttrType) String() string {
	switch p {
	case ORIGIN:
		return "ORIGIN"
	case AS_PATH:
		return "AS_PATH "
	case NEXT_HOP:
		return "NEXT_HOP"
	case MULTI_EXIT_DISC:
		return "MULTI_EXIT_DISC"
	case LOCAL_PREF:
		return "LOCAL_PREF"
	case ATOMIC_AGGREGATE:
		return "ATOMIC_AGGREGATE"
	case AGGREGATOR:
		return "AGGREGATOR"
	default:
		return "Unimplemented"
	}
}

type PathAttrTypeSet interface {
	*UnimplementedPathAttr | *Origin | *ASPath | *NextHop | *MultiExitDisc | *LocalPref
}

func GetFromPathAttrs[T PathAttrTypeSet](attrs []PathAttr) T {
	for _, attr := range attrs {
		switch t := attr.(type) {
		case T:
			return t
		}
	}
	return nil
}

type UnimplementedPathAttr struct {
	*pathAttr
	length int
	data   []byte
}

func newUnimplementedPathAttr(buf *bytes.Buffer, base *pathAttr) (*UnimplementedPathAttr, error) {
	attr := &UnimplementedPathAttr{pathAttr: base}
	var length int
	if (attr.flags & PATH_ATTR_FLAG_EXTENDED) == PATH_ATTR_FLAG_EXTENDED {
		// 2 bytes length field
		var l uint16
		if err := binary.Read(buf, binary.BigEndian, &l); err != nil {
			return nil, fmt.Errorf("newUnimplementedPathAttr: length: %w", err)
		}
		length = int(l)
	} else {
		// 1 byte length field
		l, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("newUnimplementedPathAttr: length: %w", err)
		}
		length = int(l)
	}
	attr.length = length
	data := make([]byte, length)
	if err := binary.Read(buf, binary.BigEndian, data); err != nil {
		return nil, fmt.Errorf("newUnimplementedPathAttr: data: %w", err)
	}
	attr.data = data
	return attr, nil
}

func (attr *UnimplementedPathAttr) Type() PathAttrType {
	return attr.typ
}

func (attr *UnimplementedPathAttr) Flags() uint8 {
	return attr.flags
}

func (attr *UnimplementedPathAttr) String() string {
	base := attr.pathAttr.String()
	base += "\n"
	base += "Unimplemented Path Attribute"
	return base
}

func (attr *UnimplementedPathAttr) ValueLen() int {
	return len(attr.data)
}

func (attr *UnimplementedPathAttr) IsRecognized() bool {
	return false
}

func (attr *UnimplementedPathAttr) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(attr.data)))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("UnimplementedPatAttr_Decode: base: %w", err)
	}
	if (attr.flags & PATH_ATTR_FLAG_EXTENDED) == PATH_ATTR_FLAG_EXTENDED {
		if err := binary.Write(buf, binary.BigEndian, uint16(len(attr.data))); err != nil {
			return nil, fmt.Errorf("UnimplementedPathAttr_Decode: length: %w", err)
		}
	} else {
		if err := binary.Write(buf, binary.BigEndian, uint8(len(attr.data))); err != nil {
			return nil, fmt.Errorf("UnimplementedPathAttr_Decode: length: %w", err)
		}
	}
	if err := binary.Write(buf, binary.BigEndian, attr.data); err != nil {
		return nil, fmt.Errorf("UnimplementedPathAttr_Decode: data: %w", err)
	}
	return buf.Bytes(), nil
}

// ORIGIN is a well-known mandatory attribute.
// The ORIGIN attribute is generated by the speaker that originates the associated routing information.
// Its value SHOULD NOT be changed by any other speaker.
//   0: IGP - Network Layer Reachability Information is interior to the originating AS
//   1: EGP - Network Layer Reachability Information learned via the EGP protocol
//   2: INCOMPLETE - Network Layer Reachability Information learned by some other means
type Origin struct {
	*pathAttr
	length uint8
	value  uint8
}

const (
	ORIGIN_IGP        uint8 = iota
	ORIGIN_EGP        uint8 = iota
	ORIGIN_INCOMPLETE uint8 = iota
)

func newOrigin(buf *bytes.Buffer, base *pathAttr) (*Origin, error) {
	attr := &Origin{pathAttr: base}
	var err error
	attr.length, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newOrigin: length: %w", err)
	}
	attr.value, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newOrigin: value: %w", err)
	}
	return attr, nil
}

func (*Origin) Type() PathAttrType {
	return ORIGIN
}

func (attr *Origin) Flags() uint8 {
	return attr.flags
}

func (attr *Origin) String() string {
	base := attr.pathAttr.String()
	base += "\n"
	var o string
	switch attr.value {
	case ORIGIN_IGP:
		o = "IGP(0)"
	case ORIGIN_EGP:
		o = "EGP(1)"
	case ORIGIN_INCOMPLETE:
		o = "INCOMPLETE(2)"
	}
	base += fmt.Sprintf("Origin=%s", o)
	return base
}

func (attr *Origin) ValueLen() int {
	return int(attr.length)
}

func (attr *Origin) IsRecognized() bool {
	return true
}

func (attr *Origin) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, attr.length))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("Origin_Decode: base: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.length); err != nil {
		return nil, fmt.Errorf("Origin_Decode: length: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.value); err != nil {
		return nil, fmt.Errorf("Origin_Decode: value: %w", err)
	}
	return buf.Bytes(), nil
}

// AS_PATH is a well-known mandatory attribute.
// This attribute identifies the autonomous systems through which routing information carried in this UPDATE message has passed.
// The components of this list can be AS_SETs or AS_SEQUENCEs.
// 1: AS_SET - unordered set of ASes a route in the UPDATE message has traversed
// 2: AS_SEQUENCE - ordered set of ASes a route in the UPDATE message has traversed
//
// When a BGP speaker propagates a route it learned from another BGP speaker's UPDATE message,
// it modifies the route's AS_PATH attribute based on the location of the BGP speaker to which the route will be sent:
//   a) When a given BGP speaker advertises the route to an internal peer, the advertising speaker SHALL NOT modify the AS_PATH attribute associated with the route.
//   b) When a given BGP speaker advertises the route to an external peer, the advertising speaker updates the AS_PATH attribute as follows:
//     1) if the first path segment of the AS_PATH is of type AS_SEQUENCE, the local system prepends its own AS number as the last element of the sequence
//        (put it in the leftmost position with respect to the position of octets in the porotocol message).
//        If the act of prepending will cause an it overflow in the AS_PATH segment(i.e., more than 255 ASes),
//        it SHOULD prepend a new segment of type AS_SEQUENCE and prepend its own AS number to this new segment.
//     2) if the first path segment of the AS_PATH is of type AS_SET, the local system prepends a new path segment of type AS_SEQUENCE to the AS_PATH, including its own AS number in that segment.
//     3) if the AS_PATH is empty, the local system creates a path segment of type AS_SEQUENCE, places its own AS into that segment,
//        and places that segment into the AS_PATH.
//
// When a BGP speaker originates a route then:
//   a) the originating speaker includes its own AS number in a path segment, of type AS_SEQUENCE, in the AS_PATH attribute of all UPDATE messages sent to an external peer.
//      In this case, the AS number of the originating speaker's autonomous system will be the only entry the path segment,
//      and this path segment will be the only segment in the AS_PATH attribute.
//   b) the originating speaker includes an empty AS_PATH attribute in all UPDATE messages sent to internal peers.
//      (An empty AS_PATH attribute is one whose length field contains the value zero).
type ASPath struct {
	*pathAttr
	length   int
	Segments []*ASPathSegment
}

type ASPathSegment struct {
	Type   uint8
	Length uint8
	AS2    []uint16
}

func (s *ASPathSegment) String() string {
	return fmt.Sprintf("type=%d AS2=%v", s.Type, s.Type)
}

const (
	SEG_TYPE_AS_SET      uint8 = 1
	SEG_TYPE_AS_SEQUENCE uint8 = 2
)

func newASPath(buf *bytes.Buffer, base *pathAttr) (*ASPath, error) {
	var err error
	attr := &ASPath{pathAttr: base}
	if (attr.flags & PATH_ATTR_FLAG_EXTENDED) == PATH_ATTR_FLAG_EXTENDED {
		// 2 bytes length field
		var l uint16
		if err := binary.Read(buf, binary.BigEndian, &l); err != nil {
			return nil, fmt.Errorf("newASPath: length: %w", err)
		}
		attr.length = int(l)
	} else {
		// 1 byte length field
		l, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("newASPath: length: %w", err)
		}
		attr.length = int(l)
	}
	segBuf := bytes.NewBuffer(buf.Next(attr.length))
	segs := make([]*ASPathSegment, 0)
	for segBuf.Len() > 0 {
		seg := &ASPathSegment{}
		seg.Type, err = segBuf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("newASPath: seg type: %w", err)
		}
		seg.Length, err = segBuf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("newASPath: seg length: %w", err)
		}
		seg.AS2 = make([]uint16, seg.Length)
		if err := binary.Read(segBuf, binary.BigEndian, &seg.AS2); err != nil {
			return nil, fmt.Errorf("newASPath: AS2: %w", err)
		}
		segs = append(segs, seg)
	}
	attr.Segments = segs
	return attr, nil
}

func (*ASPath) Type() PathAttrType {
	return AS_PATH
}

func (attr *ASPath) Flags() uint8 {
	return attr.flags
}

func (attr *ASPath) String() string {
	base := attr.pathAttr.String()
	base += "\n"
	for _, seg := range attr.Segments {
		base += fmt.Sprintf("  %s\n", seg)
	}
	return base
}

func (attr *ASPath) ValueLen() int {
	return int(attr.length)
}

func (attr *ASPath) IsRecognized() bool {
	return true
}

func (attr *ASPath) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, attr.length))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("ASPath_Decode: base: %w", err)
	}
	if (attr.flags & PATH_ATTR_FLAG_EXTENDED) == PATH_ATTR_FLAG_EXTENDED {
		if err := binary.Write(buf, binary.BigEndian, uint16(attr.length)); err != nil {
			return nil, fmt.Errorf("ASPath_Decode: length: %w", err)
		}
	} else {
		if err := binary.Write(buf, binary.BigEndian, uint8(attr.length)); err != nil {
			return nil, fmt.Errorf("ASPath_Decode: length: %w", err)
		}
	}
	for _, seg := range attr.Segments {
		if err := binary.Write(buf, binary.BigEndian, seg.Type); err != nil {
			return nil, fmt.Errorf("ASPath_Decode: seg type: %w", err)
		}
		if err := binary.Write(buf, binary.BigEndian, seg.Length); err != nil {
			return nil, fmt.Errorf("ASPath_Decode: seg length: %w", err)
		}
		if err := binary.Write(buf, binary.BigEndian, seg.AS2); err != nil {
			return nil, fmt.Errorf("ASPath_Decode: AS2: %w", err)
		}

	}
	return buf.Bytes(), nil
}

// if detect loop, return true, otherwise return false
func (attr ASPath) CheckLoop() bool {
	dup := make(map[uint16]struct{})
	for _, seg := range attr.Segments {
		for _, a := range seg.AS2 {
			switch seg.Type {
			case SEG_TYPE_AS_SET:
				dup[a] = struct{}{}
			case SEG_TYPE_AS_SEQUENCE:
				_, ok := dup[a]
				if ok {
					return true
				}
				dup[a] = struct{}{}
			}
		}
	}
	return false
}

// if ASPath contains as number given as argument, return true, otherwise return false
func (attr ASPath) Contains(as int) bool {
	for _, seg := range attr.Segments {
		for _, a := range seg.AS2 {
			if int(a) == as {
				return true
			}
		}
	}
	return false
}

func (attr ASPath) GetSequence() []uint16 {
	for _, seg := range attr.Segments {
		if seg.Type == SEG_TYPE_AS_SEQUENCE {
			return seg.AS2
		}
	}
	return nil
}

func (attr ASPath) GetSet() []uint16 {
	for _, seg := range attr.Segments {
		if seg.Type == SEG_TYPE_AS_SET {
			return seg.AS2
		}
	}
	return nil
}

// The NEXT_HOP is a well-known mandatory attribute that defines the IP address of the router that SHOULD be used as the next hop to the destinations listed in the UPDATE message.
// The NEXT_HOP attribute is calculated as follows:
//   1) When sending a message to an internal peer, if the route is not locally originated, the BGP speaker SHOULD NOT modify the NEXT_HOP attribute unless it has been explicitly configured
//      to announce its own IP address as the NEXT_HOP.
//      When announcinsg a locally-originated route to an internal peer, the BGP speaker SHOULD use the interface address of the router through
//      which the announced network is reachable for the speaker as the NEXT_HOP.
//      If the route is directly connected to the speaker, announced network is reachable for the speaker is the internal peer's address,
//      then the BGP speaker SHOULD use its own IP address for the NEXT_HOP attribute (the address of the interface that is used to reach the peer).
//   2) When sending a message to an external peer X, and the peer is one IP hop away from the speaker:
//        - If the route being announced was learned from an internal peer or is locally originated,
//          the BGP speaker can use an interface address of the internal peer router (or the internal router) through which the announced network
//          is reachable for the speaker for the NEXT_HOP attribute,
//          provided that peer X shares a common subnet with this address.
//          This is a form of "third party" NEXT_HOP attribute.
//        - Otherwise, if the route being announced was learned from an external peer, the speaker can use an IP address of any adjacent router
//          (known from the received NEXT_HOP attribute) that the speaker itself uses for local route calculation in the NEXT_HOP attribute,
//          provided that peer X shares a common subnet with this address.
//          This is a form of "third party" NEXT_HOP attribute.
//        - By default (if none of the above conditions apply), the BGP speaker SHOULD use the IP address of the interface that the speaker uses to establish the BGP connection to peer X in the NEXT_HOP attribute.
//   3) When sending a message to an external peer X, and the peer is multiple IP hops away from the speaker (aka multihop EBGP):
//        - The speaker MAY be configured to propagate the NEXT_HOP attribute.
//          In this case, when advertising a route that the speaker learned from one of its peers, the NEXT_HOP attribute
//          of the advertised route is exactly the same as the NEXT_HOP attribute of the learned route (the speaker does not modify the NEXT_HOP attribute).
// 		  - By default, the BGP speaker SHOULD use the IP address of the interface that the speaker uses in the NEXT_HOP attribute to establish the BGP connection to peer X.
//
// Normally, the NEXT_HOP attribute is chosen such that the shortest available path will be taken.
// A BGP speaker MUST be able to support the disabling advertisement of third party NEXT_HOP attributes in order to handle imperfectly bridged media.
//
// A route originated by a BGP speaker SHALL NOT be advertised to a peer using an address of that peer as NEXT_HOP.
// A BGP speaker SHALL NOT install a route with itself as the next-hop.
//
// The NEXT_HOP attribute is used by the BGP speaker to determine the actual outbound interface and immediate next-hop address that SHOULD be used
// to forward transit packets to the associated destinations.
//
// The immediate next-hop address is determined by performing a recursive route lookup operation for the Ip address in the NEXT_HOP attribute,
// using the contentsof the Routing Table, selecting one entry if multiple entries of equal cost exist.
// The Routing Table entry that resolves the IP address in the NEXT_HOP attribute will always specify the outbound interface.
// If the entry specifies an attached subnet, but dows not specify a next-hop address, then the address in the NEXT_HOP attribute SHOULD be used as the immediate next-hop address.
// If the entry also specifies the next-hop address, this address SHOULD be used as the immediate next-hop address for packet forwarding.
type NextHop struct {
	*pathAttr
	length uint8
	next   net.IP
}

func newNextHop(buf *bytes.Buffer, base *pathAttr) (*NextHop, error) {
	attr := &NextHop{pathAttr: base}
	var err error
	attr.length, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newNextHop: length: %w", err)
	}
	b := make([]byte, attr.length)
	if err := binary.Read(buf, binary.BigEndian, b); err != nil {
		return nil, fmt.Errorf("newNextHop: next hop: %w", err)
	}
	attr.next = net.IP(b)
	return attr, nil
}

func (*NextHop) Type() PathAttrType {
	return NEXT_HOP
}

func (attr *NextHop) Flags() uint8 {
	return attr.flags
}

func (attr *NextHop) Stirng() string {
	base := attr.pathAttr.String()
	base += "\n"
	base += fmt.Sprintf("next hop=%s", attr.next)
	return base
}

func (attr *NextHop) ValueLen() int {
	return int(attr.length)
}

func (attr *NextHop) IsRecognized() bool {
	return true
}

func (attr *NextHop) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, attr.length))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("NextHop_Decode: base: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.length); err != nil {
		return nil, fmt.Errorf("NextHop_Decode: length: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.next); err != nil {
		return nil, fmt.Errorf("NextHop_Decode: next hop: %w", err)
	}
	return buf.Bytes(), nil
}

type LocalPref struct {
	*pathAttr
	length uint8
	value  uint32
}

func newLocalPref(buf *bytes.Buffer, base *pathAttr) (*LocalPref, error) {
	attr := &LocalPref{pathAttr: base}
	var err error
	attr.length, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newLocalPref: base: %w", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &attr.value); err != nil {
		return nil, fmt.Errorf("newLocalPref: value: %w", err)
	}
	return attr, nil
}

func (*LocalPref) Type() PathAttrType {
	return LOCAL_PREF
}

func (attr *LocalPref) Flags() uint8 {
	return attr.flags
}

func (attr *LocalPref) ValueLen() int {
	return int(attr.length)
}

func (attr *LocalPref) String() string {
	base := attr.pathAttr.String()
	base += "\n"
	base += fmt.Sprintf("Preference=%d", attr.value)
	return base
}

func (attr *LocalPref) IsRecognized() bool {
	return false
}

func (attr *LocalPref) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, attr.length))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("LocalPref_Decode: base: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.length); err != nil {
		return nil, fmt.Errorf("LocalPref_Decode: length: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.value); err != nil {
		return nil, fmt.Errorf("LocalPref_Decode: value: %w", err)
	}
	return buf.Bytes(), nil
}

type AtomicAggregate struct {
	*pathAttr
}

func newAtomicAggrerate(buf *bytes.Buffer, base *pathAttr) (*AtomicAggregate, error) {
	attr := &AtomicAggregate{pathAttr: base}
	_, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newAtomicAggregate: length: %w", err)
	}
	return attr, nil
}

func (*AtomicAggregate) Type() PathAttrType {
	return AS4_AGGREGATOR
}

func (attr *AtomicAggregate) Flags() uint8 {
	return attr.flags
}

func (attr *AtomicAggregate) ValueLen() int {
	return 0
}

func (attr *AtomicAggregate) String() string {
	return attr.pathAttr.String()
}

func (attr *AtomicAggregate) IsRecognized() bool {
	return false
}

func (attr *AtomicAggregate) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("AtomicAggregate_Decode: base: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, 0); err != nil {
		return nil, fmt.Errorf("AtomicAggregate_Decode: length: %w", err)
	}
	return buf.Bytes(), nil
}

type Aggregator struct {
	*pathAttr
	length  uint8
	AS      uint16
	Address net.IP
}

func newAggregator(buf *bytes.Buffer, base *pathAttr) (*Aggregator, error) {
	attr := &Aggregator{pathAttr: base}
	var err error
	attr.length, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newAggregator: length: %w", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &attr.AS); err != nil {
		return nil, fmt.Errorf("newAggregator: AS: %w", err)
	}
	b := make([]byte, 4)
	if err := binary.Read(buf, binary.BigEndian, b); err != nil {
		return nil, fmt.Errorf("newAggregator: Address: %w", err)
	}
	attr.Address = net.IP(b)
	return attr, nil
}

func (*Aggregator) Type() PathAttrType {
	return AGGREGATOR
}

func (attr *Aggregator) Flags() uint8 {
	return attr.flags
}

func (attr *Aggregator) ValueLen() int {
	return int(attr.length)
}

func (attr *Aggregator) String() string {
	base := attr.pathAttr.String()
	base += "\n"
	base += fmt.Sprintf("AS=%d Address=%s", attr.AS, attr.Address)
	return base
}

func (attr *Aggregator) IsRecognized() bool {
	return false
}

func (attr *Aggregator) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, attr.length))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("Aggregator_Decode: base: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.length); err != nil {
		return nil, fmt.Errorf("Aggregator_Decode: length: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.AS); err != nil {
		return nil, fmt.Errorf("Aggregator_Decode: AS: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.Address); err != nil {
		return nil, fmt.Errorf("Aggregator_Decode: Address: %w", err)
	}
	return buf.Bytes(), nil
}

type MultiExitDisc struct {
	*pathAttr
	length        uint8
	discriminator uint32
}

func newMultiExitDisc(buf *bytes.Buffer, base *pathAttr) (*MultiExitDisc, error) {
	attr := &MultiExitDisc{pathAttr: base}
	var err error
	attr.length, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newMultiExitDisc: length: %w", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &attr.discriminator); err != nil {
		return nil, fmt.Errorf("newMultiExitDisc: discriminator: %w", err)
	}
	return attr, nil
}

func (*MultiExitDisc) Type() PathAttrType {
	return MULTI_EXIT_DISC
}

func (attr *MultiExitDisc) Flags() uint8 {
	return attr.flags
}

func (attr *MultiExitDisc) ValueLen() int {
	return int(attr.length)
}

func (attr *MultiExitDisc) String() string {
	base := attr.pathAttr.String()
	base += "\n"
	base += fmt.Sprintf("Multiple exit discreminator=%d", attr.discriminator)
	return base
}

func (attr *MultiExitDisc) IsRecognized() bool {
	return true
}

func (attr *MultiExitDisc) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, attr.length))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("MultiExitDisc_Decode: base: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.length); err != nil {
		return nil, fmt.Errorf("MultiExitDisc_Decode: length: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.discriminator); err != nil {
		return nil, fmt.Errorf("MultiExitDisc_Decode: discriminator: %w", err)
	}
	return buf.Bytes(), nil
}

type Commutities struct {
	*pathAttr
	length uint8
	value  uint32
}

func newCommunities(buf *bytes.Buffer, base *pathAttr) (*Commutities, error) {
	attr := &Commutities{pathAttr: base}
	var err error
	attr.length, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newCommunities: length: %w", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &attr.value); err != nil {
		return nil, fmt.Errorf("newCommunities: value: %w", err)
	}
	return attr, nil
}

func (*Commutities) Type() PathAttrType {
	return COMMUNITIES
}

func (attr *Commutities) Flags() uint8 {
	return attr.flags
}

func (attr *Commutities) ValueLen() int {
	return int(attr.length)
}

func (attr *Commutities) String() string {
	base := attr.pathAttr.String()
	base += "\n"
	base += fmt.Sprintf("Community=%d", attr.value)
	return base
}

func (attr *Commutities) IsRecognized() bool {
	return false
}

func (attr *Commutities) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, attr.length))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("Communities_Decode: base: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.length); err != nil {
		return nil, fmt.Errorf("Communities_Decode: length: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.value); err != nil {
		return nil, fmt.Errorf("Communities_Decode: value: %w", err)
	}
	return buf.Bytes(), nil
}

type ExtendedCommutities struct {
	*pathAttr
	length uint8
	value  uint64
}

func newExtendedCommunities(buf *bytes.Buffer, base *pathAttr) (*ExtendedCommutities, error) {
	attr := &ExtendedCommutities{pathAttr: base}
	var err error
	attr.length, err = buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("newExtendedCommunities: length: %w", err)
	}
	if err := binary.Read(buf, binary.BigEndian, &attr.value); err != nil {
		return nil, fmt.Errorf("newExtendedCommunities: value: %w", err)
	}
	return attr, nil
}

func (*ExtendedCommutities) Type() PathAttrType {
	return EXTENDED_COMMUNITIES
}

func (attr *ExtendedCommutities) Flags() uint8 {
	return attr.flags
}

func (attr *ExtendedCommutities) ValueLen() int {
	return int(attr.length)
}

func (attr *ExtendedCommutities) String() string {
	base := attr.pathAttr.String()
	base += "\n"
	base += fmt.Sprintf("Extended Community=%d", attr.value)
	return base
}

func (attr *ExtendedCommutities) IsRecognized() bool {
	return false
}

func (attr *ExtendedCommutities) Decode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, attr.length))
	if err := binary.Write(buf, binary.BigEndian, attr.pathAttr); err != nil {
		return nil, fmt.Errorf("ExtendedCommunities_Decode: base: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.length); err != nil {
		return nil, fmt.Errorf("ExtendedCommunities_Decode: length: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, attr.value); err != nil {
		return nil, fmt.Errorf("ExtendedCommunities_Decode: value: %w", err)
	}
	return buf.Bytes(), nil
}
