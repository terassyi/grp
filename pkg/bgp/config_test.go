package bgp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfigFromBytes(t *testing.T) {
	for _, d := range []struct {
		data []byte
		ext  string
		conf *Config
	}{
		{data: []byte("{\"AS\": 100, \"router-id\": \"1.1.1.1\"}"), ext: "json", conf: &Config{AS: 100, RouterId: "1.1.1.1"}},
		{data: []byte("AS: 100\nrouter-id: \"1.1.1.1\"\n"), ext: "yml", conf: &Config{AS: 100, RouterId: "1.1.1.1"}},
		{data: []byte(
			`AS: 100
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
				AS:        100,
				RouterId:  "1.1.1.1",
				Networks:  []string{"10.0.0.0/24", "10.0.1.0/24"},
				Neighbors: []Neighbor{{Address: "10.0.0.1", AS: 200}, {Address: "10.0.1.1", AS: 300}},
			}},
	} {
		conf, err := ConfigFromBytes(d.data, d.ext)
		require.NoError(t, err)
		assert.Equal(t, d.conf.AS, conf.AS)
		assert.Equal(t, d.conf.RouterId, conf.RouterId)
		assert.Equal(t, d.conf.Networks, conf.Networks)
		assert.Equal(t, d.conf.Neighbors, conf.Neighbors)
	}
}

func TestConfigFromMap(t *testing.T) {
	for _, d := range []struct {
		data []byte
		ext  string
		conf *Config
	}{
		{data: []byte("{\"AS\": 100, \"router-id\": \"1.1.1.1\"}"), ext: "json", conf: &Config{AS: 100, RouterId: "1.1.1.1"}},
		{data: []byte("AS: 100\nrouter-id: \"1.1.1.1\"\n"), ext: "yml", conf: &Config{AS: 100, RouterId: "1.1.1.1"}},
		{data: []byte(
			`AS: 100
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
				AS:        100,
				RouterId:  "1.1.1.1",
				Networks:  []string{"10.0.0.0/24", "10.0.1.0/24"},
				Neighbors: []Neighbor{{Address: "10.0.0.1", AS: 200}, {Address: "10.0.1.1", AS: 300}},
			}},
	} {
		mappedData := make(map[string]any)
		switch d.ext {
		case "json", "JSON":
			if err := json.Unmarshal(d.data, &mappedData); err != nil {
				t.Fatal(err)
			}
		default:
			if err := yaml.Unmarshal(d.data, mappedData); err != nil {
				t.Fatal(err)
			}
		}
		conf, err := ConfigFromMap(mappedData)
		require.NoError(t, err)
		assert.Equal(t, d.conf.AS, conf.AS)
		assert.Equal(t, d.conf.RouterId, conf.RouterId)
		assert.Equal(t, d.conf.Networks, conf.Networks)
		assert.Equal(t, d.conf.Neighbors, conf.Neighbors)
	}
}
