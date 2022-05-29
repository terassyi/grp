package bgp

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AS        int        `json:"AS" yaml:"AS"`
	RouterId  string     `json:"router-id" yaml:"router-id"`
	Port      int        `json:"port,omitempty" yaml:"port,omitempty"`
	Networks  []string   `json:"networks,omitempty" yaml:"networks,omitempty"`
	Neighbors []Neighbor `json:"neighbors,omitempty" yaml:"neighbors,omitempty"`
}

type Neighbor struct {
	Address string `json:"address" yaml:"address"`
	AS      int    `json:"as" yaml:"AS"`
}

func ConfigFromBytes(data []byte, ext string) (*Config, error) {
	conf := &Config{}
	switch ext {
	case "json", "JSON":
		if err := json.Unmarshal(data, &conf); err != nil {
			return nil, err
		}
		return conf, nil
	case "yaml", "yml", "YAML":
		if err := yaml.Unmarshal(data, &conf); err != nil {
			return nil, err
		}
		return conf, nil
	default:
		return nil, fmt.Errorf("Invalid config file ext.")
	}
}

func ConfigFromMap(data map[string]any) (*Config, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	if err := json.Unmarshal(jsonData, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
