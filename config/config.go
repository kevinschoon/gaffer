package config

type Config struct {
	Store      Store
	Runc       Runc
	RPCServer  RPCServer
	HTTPServer HTTPServer
	User       User
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

type RPCServer struct {
	Port int
}

type HTTPServer struct {
	Port int
}

type User struct {
	User string
}
