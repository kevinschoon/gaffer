package config

import (
	"encoding/json"
	"google.golang.org/grpc"
	"io/ioutil"
)

// Config holds all configurable options
// within Gaffer.
type Config struct {
	Init   Init   `json:"init"`
	Store  Store  `json:"store"`
	Logger Logger `json:"logger"`
	// RPC Address
	Address string `json:"address"`
	// etcd endpoints
	Endpoints []string `json:"endpoints"`
	// Runc root path
	RuncRoot string `json:"runc_root"`
	// Enabled plugins
	Plugins []string `json:"plugins:`
}

func (c Config) DailOpts() []grpc.DialOption {
	// TODO
	return []grpc.DialOption{grpc.WithInsecure()}
}

func (c Config) CallOpts() []grpc.CallOption {
	// TODO
	return []grpc.CallOption{}
}

// Init holds OS initialization options
type Init struct {
	// Helper is the path to a "helper"
	// script that we execute to initialize
	// our OS on boot.
	Helper string `json:"helper"`
	// NewRoot is the path where the existing
	// tempfs contents are compied and switch
	// moves the base rootfs to.
	NewRoot string `json:"new_root"`
}

// Store holds configuration options for managing
// on-disk runc container FS.
type Store struct {
	BasePath   string `json:"base_path"`
	ConfigPath string `json:"config_path"`
	// Toggle if we should handle overlay
	// mounts ourself.
	Mount bool `json:"mount"`
	// Move lower --> rootfs
	MoveRoot bool `json:"move_root"`
	// Environment contains environment variable
	// overrides for runc apps. This is the primary
	// way os services are configured at boot.
	Environment map[string]map[string]string `json:"environment"`
}

// Logger holds logger specific options.
type Logger struct {
	// Device is the path to a
	// block device like /dev/stdout
	Device string `json:"device"`
	// LogDir is a path to a directory
	// where log files will be
	// written to and rotated.
	LogDir string `json:"log_dir"`
	// Debug toggles debug logging.
	Debug bool `json:"debug"`
	// JSON configures the logger
	// to encode log output with JSON.
	JSON bool `json:"json"`
	// MaxSize specifies
	// the maximum size (mb) of a
	// log before it is rotated. Since
	// Mesanine may operate only in
	// system memory this should be
	// very low by default.
	MaxSize int `json:"max_size"`
	// MaxBackups is the number
	// of backups to retain after
	// rotation. This number should
	// also be very low by default
	MaxBackups int `json:"max_backups"`
	// Compress indicates if
	// rotated log files should be
	// compressed
	Compress bool `json:"compress"`
}

func Load(path string, cfg *Config) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, cfg)
}
