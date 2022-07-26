package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/terassyi/grp/pkg/bgp"
	"github.com/terassyi/grp/pkg/log"
	"github.com/terassyi/grp/pkg/rip"
	"github.com/terassyi/grp/pkg/route"
	"gopkg.in/yaml.v3"
)

const (
	defaultBgpLogPath       string = "/var/log/grp/bgp"
	defaultRipLogPath       string = "/var/log/grp/rip"
	defaultRouteManagerPath string = "/var/log/grp/route"
)

type Config struct {
	Log          *log.Log      `json:"log,omitempty" yaml:"log,omitempty"`
	RouteManager *route.Config `json:"routeManager,omitempty" yaml:"routeManager,omitempty"`
	Bgp          *bgp.Config   `json:"bgp,omitempty" yaml:"bgp,omitempty"`
	Rip          *rip.Config   `json:"rip,omitempty" yaml:"rip,omitempty"`
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
	if conf.Bgp != nil {
		if conf.Bgp.Log == nil {
			conf.Bgp.Log = &log.Log{
				Level: 1,
				Out:   defaultBgpLogPath,
			}
		}
	}
	if conf.Rip != nil {
		if conf.Rip.Log == nil {
			conf.Rip.Log = &log.Log{
				Level: 1,
				Out:   defaultRipLogPath,
			}
		}
	}
	return conf, nil
}
