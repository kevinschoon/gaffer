package config

type Config struct {
	Store      Store
	Runc       Runc
	Server     Server
	User       User
	Supervisor Supervisor
}

type Store struct {
	BasePath   string
	ConfigPath string
}

type Runc struct {
	Root string
	// Toggle if we should handle overlay
	// mounts ourself.
	Mount bool
}

type Supervisor struct {
	Port int
}

type Server struct {
	Pattern string
}

type User struct {
	User string
}
