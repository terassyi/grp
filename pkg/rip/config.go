package rip

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Networks             []string `json:"networks,omitempty" yaml:"networks,omitempty"`
	Port                 int      `json:"port" "yaml:"port"`
	Timeout              int      `json:"timeout" yaml:"timeout"`
	Gc                   int      `json:"gc" yaml:"gc`
	RouteManagerEndpoint string   `json:"routeManagerEndpoint,omitempty" yaml:"routeManagerEndpoint,omitempty"`
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
