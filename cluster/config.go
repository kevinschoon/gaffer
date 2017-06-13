package cluster

type Config struct {
	Hosts    Hosts    `json:"hosts"`
	Services Services `json:"services"`
}
