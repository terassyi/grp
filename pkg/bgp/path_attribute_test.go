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

func TestPathAttrDecode(t *testing.T) {
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

func TestPathAttrIsTransitive(t *testing.T) {
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
