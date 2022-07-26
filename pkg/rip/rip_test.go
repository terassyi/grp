package rip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	for _, d := range []struct {
		data []byte
		exp  *Packet
	}{
		{
			data: []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 192, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			exp:  &Packet{Command: REQUEST, Version: 1, Entries: []*Entry{{Family: 4, Address: net.IP([]byte{192, 0, 0, 1}), Metric: 1}}},
		},
	} {
		packet, err := Parse(d.data)
		require.NoError(t, err)
		assert.Equal(t, d.exp.Command, packet.Command)
		assert.Equal(t, d.exp.Version, packet.Version)
		assert.Equal(t, d.exp.Entries, packet.Entries)
	}
}

func TestPacketDecode(t *testing.T) {
	for _, d := range []struct {
		data []byte
		exp  *Packet
	}{
		{
			data: []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 192, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		},
	} {
		packet, err := Parse(d.data)
		require.NoError(t, err)
		data, err := packet.Decode()
		require.NoError(t, err)
		assert.Equal(t, d.data, data)
	}
}

func TestBroadcast(t *testing.T) {
	tests := []struct {
		name string
		cidr *net.IPNet
		b    net.IP
	}{
		{
			name: "CASE 1",
			cidr: &net.IPNet{
				IP:   net.ParseIP("10.0.0.0"),
				Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0x00),
			},
			b: net.ParseIP("10.0.0.255"),
		},
		{
			name: "CASE 2",
			cidr: &net.IPNet{
				IP:   net.ParseIP("10.1.0.0"),
				Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0x00),
			},
			b: net.ParseIP("10.1.0.255"),
		},
		{
			name: "CASE 3",
			cidr: &net.IPNet{
				IP:   net.ParseIP("10.10.10.0"),
				Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0x00),
			},
			b: net.ParseIP("10.10.10.255"),
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := broadcast(tt.cidr)
			t.Log(tt.b)
			t.Log(b)
			assert.Equal(t, tt.b, b)
		})
	}
}
