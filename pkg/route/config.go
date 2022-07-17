package route

type Config struct {
	Run bool `json:"run" yaml:"run"`
	Host string `json:"host" yaml:"host"`
	Port int `json:"port" yaml:"port"`
}
