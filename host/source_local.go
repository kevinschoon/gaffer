package host

type SingleSource struct {
	pattern string
}

func (ss SingleSource) Get() (*Config, error) {
	host, err := New(ss.pattern)
	if err != nil {
		return nil, err
	}
	return &Config{Hosts: Hosts{host}}, nil
}

func (_ SingleSource) Set(config *Config) error { return nil }
