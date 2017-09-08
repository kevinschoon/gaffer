package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config holds all configurable options
// within Gaffer.
type Config struct {
	Init   Init   `json:"init"`
	Store  Store  `json:"store"`
	Runc   Runc   `json:"runc"`
	Etcd   Etcd   `json:"etcd"`
	User   User   `json:"user"`
	Logger Logger `json:"logger"`
	// List of enabled plugins
	Enabled []string `json:"enabled"`
	Plugins struct {
		RPCServer  RPCServer  `json:"rpc_server"`
		HTTPServer HTTPServer `json:"http_server"`
	}
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

// Runc holds runc specific options.
type Runc struct {
	Root string `json:"root"`
}

// Etcd holds etcd specific options.
type Etcd struct {
	Endpoints []string `json:"endpoints"`
}

// RPCServer holds rpc-server plugin specific options.
type RPCServer struct {
	Port int `json:"port"`
}

// HTTPServer holds http-server plugin specific options.
type HTTPServer struct {
	Port int `json:"port"`
}

// User holds user specific options.
type User struct {
	User string `json:"user"`
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

// New creates a new Config based on
// pre-configured defaults.
func New() *Config {
	config := *DefaultConfig
	return &config
}

// Load updates a configuration with options
// specified within a file.
func Load(path string, cfg *Config) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(raw, cfg)
	if err != nil {
		return err
	}
	return nil
}
