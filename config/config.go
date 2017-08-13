package config

type Config struct {
	Store      Store
	Runc       Runc
	Etcd       Etcd
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

type Etcd struct {
	Endpoints []string
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
