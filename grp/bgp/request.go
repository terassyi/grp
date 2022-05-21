package bgp

type Request struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}
