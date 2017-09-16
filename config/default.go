package config

var Default = &Config{
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
	Logger: Logger{
		JSON:       false,
		Debug:      false,
		Device:     "/dev/stderr",
		LogDir:     "",
		MaxSize:    1,
		MaxBackups: 2,
		Compress:   true,
	},
	RuncRoot:  "/run/runc",
	Endpoints: []string{"http://127.0.0.1:2379"},
	Address:   "unix:///tmp/gaffer.sock",
}
