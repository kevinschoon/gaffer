package config

var DefaultConfig = &Config{
	Init: Init{
		Helper:  "/bin/gaffer-helper",
		NewRoot: "/mnt",
	},
	Store: Store{
		MoveRoot:   false,
		Mount:      false,
		BasePath:   "/containers",
		ConfigPath: "/var/mesanine",
	},
	Runc: Runc{
		Root: "/run/runc",
	},
	Etcd: Etcd{
		Endpoints: []string{"http://127.0.0.1:2379"},
	},
	User: User{},
	Logger: Logger{
		JSON:       false,
		Debug:      false,
		Device:     "/dev/stderr",
		LogDir:     "",
		MaxSize:    1,
		MaxBackups: 2,
		Compress:   true,
	},
	Plugins: struct {
		RPCServer  RPCServer  `json:"rpc_server"`
		HTTPServer HTTPServer `json:"http_server"`
	}{
		RPCServer: RPCServer{
			Port: 10000,
		},
		HTTPServer: HTTPServer{
			Port: 9090,
		},
	},
}
