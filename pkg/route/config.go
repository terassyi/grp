package route

import "github.com/terassyi/grp/pkg/log"

type Config struct {
	Log *log.Log `json:"log,omitempty" yaml:"log,omitempty"`
	Run bool `json:"run" yaml:"run"`
	Host string `json:"host" yaml:"host"`
	Port int `json:"port" yaml:"port"`
}
