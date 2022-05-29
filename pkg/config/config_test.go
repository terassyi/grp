package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terassyi/grp/pkg/bgp"
)

func TestLoadConfig(t *testing.T) {
	for _, d := range []struct {
		data []byte
		ext  string
		conf *Config
	}{
		{data: []byte(`{"bgp": {"AS": 100, "router-id": "1.1.1.1"}}`), ext: "json", conf: &Config{Bgp: bgp.Config{AS: 100, RouterId: "1.1.1.1"}}},
		{data: []byte("bgp:\n  AS: 100\n  router-id: \"1.1.1.1\"\n"), ext: "yml", conf: &Config{Bgp: bgp.Config{AS: 100, RouterId: "1.1.1.1"}}},
		{data: []byte(
			`log:
  level: 1
  out: stdout
bgp:
  AS: 100
  router-id: "1.1.1.1"
  neighbors:
    - address: "10.0.0.1"
      AS: 200
    - address: "10.0.1.1"
      AS: 300
  networks:
    - 10.0.0.0/24
    - 10.0.1.0/24
`),
			ext: "yml",
			conf: &Config{
				Bgp: bgp.Config{
					AS:        100,
					RouterId:  "1.1.1.1",
					Networks:  []string{"10.0.0.0/24", "10.0.1.0/24"},
					Neighbors: []bgp.Neighbor{{Address: "10.0.0.1", AS: 200}, {Address: "10.0.1.1", AS: 300}},
				},
			}},
	} {
		conf, err := loadConfig(d.data, d.ext)
		require.NoError(t, err)
		assert.Equal(t, d.conf.Bgp.AS, conf.Bgp.AS)
		assert.Equal(t, d.conf.Bgp.RouterId, conf.Bgp.RouterId)
		assert.Equal(t, d.conf.Bgp.Networks, conf.Bgp.Networks)
		assert.Equal(t, d.conf.Bgp.Neighbors, conf.Bgp.Neighbors)
	}
}
