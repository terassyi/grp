package bgp

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePathAttrs(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		attrs []PathAttrType
	}{
		{
			name: "CASE 1",
			data: []byte{0x40, 0x01, 0x01, 0x00, 0x50, 0x02, 0x00, 0x04, 0x02, 0x01, 0x00, 0xc8, 0x40, 0x03, 0x04, 0x0a,
				0x00, 0x00, 0x02},
			attrs: []PathAttrType{ORIGIN, AS_PATH, NEXT_HOP},
		},
		{
			name: "CASE 2",
			data: []byte{0x40, 0x01, 0x01, 0x00, 0x50, 0x02, 0x00, 0x06, 0x02, 0x02, 0x00, 0xc8, 0x01, 0x90, 0x40, 0x03,
				0x04, 0x0a, 0x00, 0x00, 0x02},
			attrs: []PathAttrType{ORIGIN, AS_PATH, NEXT_HOP},
		},
		{
			name:  "CASE 3",
			data:  []byte{0x40, 0x01, 0x01, 0x00, 0x50, 0x02, 0x00, 0x04, 0x02, 01, 0x00, 0xc8, 0x40, 0x03, 0x04, 0x0a, 0x00, 0x00, 0x02, 0x80, 0x04, 0x04, 0x00, 0x00, 0x00, 0x00},
			attrs: []PathAttrType{ORIGIN, AS_PATH, NEXT_HOP, MULTI_EXIT_DISC},
		},
		{
			name: "COMMUNITIES",
			data: []byte{0x40, 0x01, 0x01, 0x00, 0x40, 0x02, 0x04, 0x02, 0x01, 0x00, 0x64, 0x40, 0x03,
				0x04, 0x0a, 0x01, 0x0c, 0x01, 0x80, 0x04, 0x04, 0x00, 0x00, 0x00, 0x00, 0xc0, 0x08, 0x04, 0xff, 0xff, 0xff, 0x02},
			attrs: []PathAttrType{ORIGIN, AS_PATH, NEXT_HOP, MULTI_EXIT_DISC, COMMUNITIES},
		},
		{
			name: "MP_REACH_NLRI Unimplemented",
			data: []byte{0x40, 0x01, 0x01, 0x00, 0x40, 0x02, 0x04, 0x02, 0x01, 0xfd, 0xea, 0x80, 0x04, 0x04, 0x00,
				0x00, 0x00, 0x00, 0x80, 0x0e, 0x40, 0x00, 0x02, 0x01, 0x20, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00,
				00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xfe, 0x80, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0xc0, 0x02, 0x0b, 0xff, 0xfe, 0x7e, 0x00, 0x00, 0x00, 0x40, 0x20, 0x01, 0x0d, 0xb8,
				0x00, 0x02, 0x00, 0x02, 0x40, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x02, 0x00, 0x01, 0x40, 0x20, 0x01,
				0x0d, 0xb8, 0x00, 0x02, 0x00, 0x00},
			attrs: []PathAttrType{ORIGIN, AS_PATH, MULTI_EXIT_DISC, PathAttrType(14)},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			attrs, err := ParsePathAttrs(bytes.NewBuffer(tt.data))
			require.NoError(t, err)
			for i := 0; i < len(tt.attrs); i++ {
				assert.Equal(t, tt.attrs[i], attrs[i].Type())
			}
		})
	}
}

func TestParsePathAttribute(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		attr     PathAttr
		attrType PathAttrType
	}{
		{
			name: "AS_PATH 200",
			data: []byte{0x50, 0x02, 0x00, 0x04, 0x02, 0x01, 0x00, 0xc8},
			attr: &ASPath{
				pathAttr: &pathAttr{flags: 0x50, typ: AS_PATH},
				Segments: []*ASPathSegment{
					{Type: SEG_TYPE_AS_SEQUENCE, Length: 1, AS2: []uint16{200}},
				},
			},
		},
		{
			name: "AS_PATH 200, 400",
			data: []byte{0x50, 0x02, 0x00, 0x06, 0x02, 0x02, 0x00, 0xc8, 0x01, 0x90},
			attr: &ASPath{
				pathAttr: &pathAttr{flags: 0x50, typ: AS_PATH},
				Segments: []*ASPathSegment{
					{Type: SEG_TYPE_AS_SEQUENCE, Length: 2, AS2: []uint16{200, 400}},
				},
			},
		},
		{
			name: "ORIGIN IGP",
			data: []byte{0x40, 0x01, 0x01, 0x00},
			attr: &Origin{
				pathAttr: &pathAttr{flags: 0x40, typ: ORIGIN},
				value:    ORIGIN_IGP,
			},
		},
		{
			name: "ORIGIN EGP",
			data: []byte{0x40, 0x01, 0x01, 0x01},
			attr: &Origin{
				pathAttr: &pathAttr{flags: 0x40, typ: ORIGIN},
				value:    ORIGIN_EGP,
			},
		},
		{
			name: "NEXT_HOP 10.0.0.2",
			data: []byte{0x40, 0x03, 0x04, 0x0a, 0x00, 0x00, 0x02},
			attr: &NextHop{
				pathAttr: &pathAttr{flags: 0x40, typ: NEXT_HOP},
				length:   4,
				next:     net.ParseIP("10.0.0.2"),
			},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			attr, err := parsePathAttr(bytes.NewBuffer(tt.data))
			require.NoError(t, err)
			assert.Equal(t, tt.attr.Flags(), attr.Flags())
			assert.Equal(t, tt.attr.Type(), attr.Type())
			assert.Equal(t, tt.attr.String(), attr.String())
		})
	}
}

func TestPathAttr_Decode(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		attrType PathAttrType
	}{
		{
			name:     "AS_PATH 1",
			data:     []byte{0x50, 0x02, 0x00, 0x04, 0x02, 0x01, 0x00, 0xc8},
			attrType: AS_PATH,
		},
		{
			name:     "AS_PATH 2",
			data:     []byte{0x50, 0x02, 0x00, 0x06, 0x02, 0x02, 0x00, 0xc8, 0x01, 0x90},
			attrType: AS_PATH,
		},
		{
			name:     "ORIGIN IGP",
			data:     []byte{0x40, 0x01, 0x01, 0x00},
			attrType: ORIGIN,
		},
		{
			name:     "ORIGIN EGP",
			data:     []byte{0x40, 0x01, 0x01, 0x01},
			attrType: ORIGIN,
		},
		{
			name:     "NEXT_HOP 10.0.0.2",
			data:     []byte{0x40, 0x03, 0x04, 0x0a, 0x00, 0x00, 0x02},
			attrType: NEXT_HOP,
		},
		{
			name:     "COMMUNITIES",
			data:     []byte{0xc0, 0x08, 0x04, 0xff, 0xff, 0xff, 0x02},
			attrType: COMMUNITIES,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			attr, err := parsePathAttr(bytes.NewBuffer(tt.data))
			require.NoError(t, err)
			res, err := attr.Decode()
			require.NoError(t, err)
			assert.Equal(t, tt.data, res)
		})
	}
}

func TestPathAttr_IsTransitive(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		typ          PathAttrType
		isTransitive bool
	}{
		{
			name:         "ORIGIN: Transitive",
			data:         []byte{0x40, 0x01, 0x01, 0x00},
			typ:          ORIGIN,
			isTransitive: true,
		},
		{
			name:         "COMMUNITIES: Transitive",
			data:         []byte{0xc0, 0x08, 0x04, 0xff, 0xff, 0xff, 0x02},
			typ:          COMMUNITIES,
			isTransitive: true,
		},
		{
			name:         "MULTI_EXIT_DISC: Non transitive",
			data:         []byte{0x80, 0x04, 0x04, 0x00, 0x00, 0x00, 0x00},
			typ:          MULTI_EXIT_DISC,
			isTransitive: false,
		},
		{
			name:         "LOCAL_PREF: Transitive",
			data:         []byte{0x40, 0x05, 0x04, 0x00, 0x00, 0x00, 0x64},
			typ:          LOCAL_PREF,
			isTransitive: true,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr, err := parsePathAttr(bytes.NewBuffer(tt.data))
			require.NoError(t, err)
			assert.Equal(t, tt.typ, attr.Type())
			assert.Equal(t, tt.isTransitive, attr.IsTransitive())
		})
	}
}

func TestASPath_CheckLoop(t *testing.T) {
	tests := []struct {
		name   string
		asPath ASPath
		res    bool
	}{
		{name: "NO LOOP 1", asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 300}}}}, res: false},
		{name: "NO LOOP 2", asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 300, 500, 1000}}}}, res: false},
		{name: "LOOP 1", asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 300, 100}}}}, res: true},
		{name: "LOOP 2", asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 1000, 300, 500, 1000}}}}, res: true},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.res, tt.asPath.CheckLoop())
		})
	}
}

func TestASPath_Contains(t *testing.T) {
	tests := []struct {
		name   string
		asPath ASPath
		as     int
		res    bool
	}{
		{name: "CONTAIN 1", asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 300}}}}, as: 200, res: true},
		{name: "CONTAIN 2", asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 300, 500, 1000}}}}, as: 1000, res: true},
		{name: "NOT CONTAIN 1", asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 300, 100}}}}, as: 11, res: false},
		{name: "NOT CONTAIN 2", asPath: ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{100, 200, 1000, 300, 500, 1000}}}}, as: 600, res: false},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.res, tt.asPath.Contains(tt.as))
		})
	}
}

func TestGetFromPathAttrs(t *testing.T) {
	tests := []struct {
		name      string
		attrs     []PathAttr
		want      PathAttrType
		expectNil bool
	}{
		{name: "EXIST 1", attrs: []PathAttr{&Origin{}}, want: ORIGIN},
		{name: "EXIST 2", attrs: []PathAttr{&Origin{}, &ASPath{}, &MultiExitDisc{}}, want: MULTI_EXIT_DISC},
		{name: "Not EXIST 1", attrs: []PathAttr{&Origin{}, &ASPath{}, &MultiExitDisc{}}, want: NEXT_HOP, expectNil: true},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want {
			case ORIGIN:
				res := GetFromPathAttrs[*Origin](tt.attrs)
				if tt.expectNil && res != nil {
					t.Fatalf("expect nil, but got %s", res.typ.String())
				} else {
					assert.Equal(t, &Origin{}, res)
				}
			case AS_PATH:
				res := GetFromPathAttrs[*ASPath](tt.attrs)
				if tt.expectNil && res != nil {
					t.Fatalf("expect nil, but got %s", res.typ.String())
				} else {
					assert.Equal(t, &ASPath{}, res)
				}
			case MULTI_EXIT_DISC:
				res := GetFromPathAttrs[*MultiExitDisc](tt.attrs)
				if tt.expectNil && res != nil {
					t.Fatalf("expect nil, but got %s", res.typ.String())
				} else {
					assert.Equal(t, &MultiExitDisc{}, res)
				}
			}
		})
	}
}

func TestASPathSegment_UpdateSequence(t *testing.T) {
	tests := []struct {
		name string
		seg  *ASPathSegment
		asn  uint16
		want []uint16
	}{
		{
			name: "200 100",
			seg:  &ASPathSegment{Type: SEG_TYPE_AS_SEQUENCE, Length: 1, AS2: []uint16{100}},
			asn:  200,
			want: []uint16{200, 100},
		},
		{
			name: "300 200 100",
			seg:  &ASPathSegment{Type: SEG_TYPE_AS_SEQUENCE, Length: 2, AS2: []uint16{200, 100}},
			asn:  300,
			want: []uint16{300, 200, 100},
		},
		{
			name: "100",
			seg:  &ASPathSegment{Type: SEG_TYPE_AS_SEQUENCE, Length: 0, AS2: nil},
			asn:  100,
			want: []uint16{100},
		},
		{
			name: "AS_SET",
			seg:  &ASPathSegment{Type: SEG_TYPE_AS_SET, Length: 1, AS2: []uint16{100}},
			asn:  200,
			want: []uint16{100},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.seg.UpdateSequence(tt.asn)
			assert.Equal(t, tt.want, tt.seg.AS2)
			assert.Equal(t, len(tt.want), int(tt.seg.Length))
		})
	}
}

func TestASPath_Equal(t *testing.T) {
	tests := []struct {
		name   string
		asPath *ASPath
		target *ASPath
		res    bool
	}{
		{
			name:   "Only Sequence Equal",
			asPath: CreateASPath([]uint16{100, 200, 300}),
			target: CreateASPath([]uint16{100, 200, 300}),
			res:    true,
		},
		{
			name:   "Only Sequence Not Equal",
			asPath: CreateASPath([]uint16{100, 200, 300}),
			target: CreateASPath([]uint16{100, 400, 300}),
			res:    false,
		},
		{
			name: "Set and Sequence Equal",
			asPath: &ASPath{
				pathAttr: &pathAttr{typ: AS_PATH, flags: PATH_ATTR_FLAG_EXTENDED},
				Segments: []*ASPathSegment{
					{
						Type:   SEG_TYPE_AS_SEQUENCE,
						Length: 4,
						AS2:    []uint16{100, 200, 300, 400},
					},
					{
						Type:   SEG_TYPE_AS_SET,
						Length: 2,
						AS2:    []uint16{500, 600},
					},
				},
			},
			target: &ASPath{
				pathAttr: &pathAttr{typ: AS_PATH, flags: PATH_ATTR_FLAG_EXTENDED},
				Segments: []*ASPathSegment{
					{
						Type:   SEG_TYPE_AS_SEQUENCE,
						Length: 4,
						AS2:    []uint16{100, 200, 300, 400},
					},
					{
						Type:   SEG_TYPE_AS_SET,
						Length: 2,
						AS2:    []uint16{500, 600},
					},
				},
			},
			res: true,
		},
		{
			name: "Set and Sequence Not Equal",
			asPath: &ASPath{
				pathAttr: &pathAttr{typ: AS_PATH, flags: PATH_ATTR_FLAG_EXTENDED},
				Segments: []*ASPathSegment{
					{
						Type:   SEG_TYPE_AS_SEQUENCE,
						Length: 4,
						AS2:    []uint16{100, 200, 300, 400},
					},
					{
						Type:   SEG_TYPE_AS_SET,
						Length: 2,
						AS2:    []uint16{500, 800},
					},
				},
			},
			target: &ASPath{
				pathAttr: &pathAttr{typ: AS_PATH, flags: PATH_ATTR_FLAG_EXTENDED},
				Segments: []*ASPathSegment{
					{
						Type:   SEG_TYPE_AS_SEQUENCE,
						Length: 4,
						AS2:    []uint16{100, 200, 300, 400},
					},
					{
						Type:   SEG_TYPE_AS_SET,
						Length: 2,
						AS2:    []uint16{500, 600},
					},
				},
			},
			res: false,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res := tt.asPath.Equal(tt.target)
			assert.Equal(t, tt.res, res)
		})
	}
}

func TestASPath_UpdateSequence(t *testing.T) {
	tests := []struct {
		name   string
		asPath *ASPath
		asn    uint16
		exp    []uint16
	}{
		{
			name:   "CASE 1",
			asPath: CreateASPath([]uint16{}),
			asn:    100,
			exp:    []uint16{100},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.asPath.UpdateSequence(tt.asn)
			assert.Equal(t, tt.exp, tt.asPath.GetSequence())
		})
	}
}
