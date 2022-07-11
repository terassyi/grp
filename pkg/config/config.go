package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/terassyi/grp/pkg/bgp"
	"gopkg.in/yaml.v3"
)

type Config struct {
	*Log `json:"log,omitempty" yaml:"log,omitempty"`
	Bgp  *bgp.Config `json:"bgp,omitempty" yaml:"bgp,omitempty"`
}

type Log struct {
	Level int    `json:"level" yaml:"level"`
	Out   string `json:"out,omitempty" yaml:"out,omitempty"`
}

func loadConfig(data []byte, ext string) (*Config, error) {
	conf := &Config{}
	switch ext {
	case "json", "JSON":
		if err := json.Unmarshal(data, conf); err != nil {
			return nil, err
		}
	case "yaml", "yml", "YAML":
		if err := yaml.Unmarshal(data, conf); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Invalid config file format: %s", ext)
	}
	return conf, nil
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	ext := filepath.Ext(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	conf, err := loadConfig(data, ext[1:])
	if err != nil {
		return nil, err
	}
	proto := ""
	if conf.Bgp != nil {
		proto = "bgp"
	}
	if conf.Log == nil {
		conf.Log = &Log{
			Level: 1,
			Out:   fmt.Sprintf("/var/log/grp/%s", proto),
		}
	}
	return conf, nil
}
