package config

type Config struct {
	Store  Store
	Runc   Runc
	Server Server
	User   User
}

type Store struct {
	BasePath   string
	ConfigPath string
}

type Runc struct {
	Root string
}

type Server struct {
	Pattern string
}

type User struct {
	User string
}
