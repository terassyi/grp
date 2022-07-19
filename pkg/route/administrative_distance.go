package route

type AdministrativeDistance uint8

const (
	ADConnected AdministrativeDistance = 0
	ADStatic    AdministrativeDistance = 1
	ADEBGP      AdministrativeDistance = 20
	ADOSPF      AdministrativeDistance = 110
	ADRIP       AdministrativeDistance = 120
	ADIBGP      AdministrativeDistance = 200
	ADUnknown   AdministrativeDistance = 255
)

const (
	RT_PROTO_UNSPEC   int = 0
	RT_PROTO_REDIRECT int = 1
	RT_PROTO_KERNEL   int = 2
	RT_PROTO_BOOT     int = 3
	RT_PROTO_STATIC   int = 4
	RT_PROTO_BGP      int = 186
	RT_PROTO_ISIS     int = 187
	RT_PROTO_OSPF     int = 188
	RT_PROTO_RIP      int = 189
)

func ProtoString(proto int) string {
	switch proto {
	case RT_PROTO_UNSPEC:
		return "Unspec"
	case RT_PROTO_REDIRECT:
		return "Redirect"
	case RT_PROTO_KERNEL:
		return "Kernel"
	case RT_PROTO_BOOT:
		return "Boot"
	case RT_PROTO_STATIC:
		return "Static"
	case RT_PROTO_BGP:
		return "BGP"
	case RT_PROTO_ISIS:
		return "ISIS"
	case RT_PROTO_OSPF:
		return "OSPF"
	case RT_PROTO_RIP:
		return "RIP"
	default:
		return "Unknown"
	}
}

func AdFromProto(proto int, bgpExternal bool) AdministrativeDistance {
	switch proto {
	case RT_PROTO_KERNEL:
		return ADConnected
	case RT_PROTO_STATIC:
		return ADStatic
	case RT_PROTO_BGP:
		if bgpExternal {
			return ADEBGP
		}
		return ADIBGP
	case RT_PROTO_ISIS:
		return ADUnknown
	case RT_PROTO_OSPF:
		return ADOSPF
	case RT_PROTO_RIP:
		return ADRIP
	default:
		return ADUnknown
	}
}

func (ad AdministrativeDistance) String() string {
	switch ad {
	case ADConnected:
		return "Connected"
	case ADStatic:
		return "Static"
	case ADEBGP:
		return "eBGP"
	case ADOSPF:
		return "OSPF"
	case ADRIP:
		return "RIP"
	case ADIBGP:
		return "iBGP"
	default:
		return "Unknown"
	}
}
