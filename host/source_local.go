package host

type LocalSource struct {
	pattern string
}

func (ls LocalSource) Get() (*Config, error) {
	host, err := New(ls.pattern)
	if err != nil {
		return nil, err
	}
	return &Config{Hosts: Hosts{host}}, nil
}

func (_ LocalSource) Set(config *Config) error { return nil }
