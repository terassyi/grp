package bgp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMultiProtocolExtensions(t *testing.T) {
	for _, tt := range []struct {
		data []byte
		m    *MultiProtocolExtensions
	}{
		{data: []byte{0x00, 0x01, 0x00, 0x01}, m: &MultiProtocolExtensions{AFI: AFI_IPv4, SAFI: SAFI_UNICAST}},
		{data: []byte{0x00, 0x02, 0x00, 0x01}, m: &MultiProtocolExtensions{AFI: AFI_IPv6, SAFI: SAFI_UNICAST}},
		{data: []byte{0x00, 0x01, 0x00, 0x02}, m: &MultiProtocolExtensions{AFI: AFI_IPv4, SAFI: SAFI_MULTICAST}},
	} {
		m := newMultiProtocolExtensions(tt.data)
		assert.Equal(t, tt.m.AFI, m.AFI)
		assert.Equal(t, tt.m.SAFI, m.SAFI)
	}
}

func TestMultiProtocolExtensionsDecode(t *testing.T) {
	for _, tt := range []struct {
		m    *MultiProtocolExtensions
		data []byte
	}{
		{data: []byte{0x01, 0x04, 0x00, 0x01, 0x00, 0x01}, m: &MultiProtocolExtensions{AFI: AFI_IPv4, SAFI: SAFI_UNICAST}},
		{data: []byte{0x01, 0x04, 0x00, 0x02, 0x00, 0x01}, m: &MultiProtocolExtensions{AFI: AFI_IPv6, SAFI: SAFI_UNICAST}},
		{data: []byte{0x01, 0x04, 0x00, 0x01, 0x00, 0x02}, m: &MultiProtocolExtensions{AFI: AFI_IPv4, SAFI: SAFI_MULTICAST}},
	} {
		b, err := tt.m.Decode()
		require.NoError(t, err)
		assert.Equal(t, tt.data, b)
	}
}

func TestNewGracefulRestartCapability(t *testing.T) {
	for _, tt := range []struct {
		g    *GracefulRestartCapability
		data []byte
	}{
		{
			g:    &GracefulRestartCapability{RestartStateFlag: false, GracefulNotification: false, RestartTime: 120},
			data: []byte{0x00, 0x78},
		},
	} {
		g, err := newGracefulRestartCapability(tt.data)
		require.NoError(t, err)
		assert.Equal(t, tt.g.RestartStateFlag, g.RestartStateFlag)
		assert.Equal(t, tt.g.GracefulNotification, g.GracefulNotification)
		assert.Equal(t, tt.g.RestartTime, g.RestartTime)
	}
}

func TestGracefulRestartCapabilityDecode(t *testing.T) {
	for _, tt := range []struct {
		g    *GracefulRestartCapability
		data []byte
	}{
		{
			g:    &GracefulRestartCapability{RestartStateFlag: false, GracefulNotification: false, RestartTime: 120},
			data: []byte{0x00, 0x78},
		},
	} {
		d, err := tt.g.Decode()
		require.NoError(t, err)
		assert.Equal(t, tt.data, d)
	}
}
