package config

type Config struct {
	Store  Store
	Runc   Runc
	Server Server
	User   User
}

type Store struct {
	BasePath string
}

type Runc struct {
	LogPath string
}

type Server struct {
	Pattern string
}

type User struct {
	User string
}
