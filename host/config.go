package host

// Config represents the host configuration
// of a cluster of Mesanine servers.
type Config struct {
	Hosts Hosts `json:"hosts"`
}
