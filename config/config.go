package config

type Config struct {
	Store              Store
	Runc               Runc
	RegistrationServer RegistrationServer
	RPCServer          RPCServer
	HTTPServer         HTTPServer
	User               User
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

type RegistrationServer struct {
	EtcdEndpoints []string
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
