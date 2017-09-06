package config

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"strconv"
	"strings"
)

func SetCLIOpts(cmd *cli.Cli, cfg *Config) {
	cmd.VarOpt(
		"d device",
		value{
			setFn: func(s string) error {
				cfg.Logger.Device = s
				return nil
			},
			getFn: func() string { return cfg.Logger.Device },
		},
		"send log output to a block device",
	)
	cmd.VarOpt(
		"log-dir",
		value{
			setFn: func(s string) error {
				cfg.Logger.LogDir = s
				return nil
			},
			getFn: func() string { return cfg.Logger.LogDir },
		},
		"send log output to files in this directory",
	)
	cmd.VarOpt(
		"max-log-size",
		value{
			setFn: func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}
				cfg.Logger.MaxSize = int(i)
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%d", cfg.Logger.MaxSize) },
		},
		"maximum log file size in mb",
	)
	cmd.VarOpt(
		"max-backups",
		value{
			setFn: func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}
				cfg.Logger.MaxBackups = int(i)
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%d", cfg.Logger.MaxBackups) },
		},
		"maximum number of backups to rotate",
	)
	cmd.VarOpt(
		"compress",
		value{
			setFn: func(s string) error {
				v, err := strconv.ParseBool(s)
				if err != nil {
					return err
				}
				cfg.Logger.Compress = v
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%t", cfg.Logger.Compress) },
		},
		"compress rotated log files",
	)
	cmd.VarOpt(
		"debug",
		value{
			setFn: func(s string) error {
				v, err := strconv.ParseBool(s)
				if err != nil {
					return err
				}
				cfg.Logger.Debug = v
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%t", cfg.Logger.Debug) },
		},
		"output debug information",
	)
}

func SetInitOpts(cmd *cli.Cmd, cfg *Config) {
	cmd.VarOpt(
		"config-path",
		value{
			setFn: func(s string) error {
				cfg.Store.ConfigPath = s
				return nil
			},
			getFn: func() string { return cfg.Store.ConfigPath },
		},
		"service configuration path",
	)
	cmd.VarOpt(
		"store-path",
		value{
			setFn: func(s string) error {
				cfg.Store.BasePath = s
				return nil
			},
			getFn: func() string { return cfg.Store.BasePath },
		},
		"base store path",
	)
	cmd.VarOpt(
		"runc-root",
		value{
			setFn: func(s string) error {
				cfg.Runc.Root = s
				return nil
			},
			getFn: func() string { return cfg.Runc.Root },
		},
		"runc root path",
	)
	cmd.VarOpt(
		"http-port",
		value{
			setFn: func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}
				cfg.Plugins.HTTPServer.Port = int(i)
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%d", cfg.Plugins.HTTPServer.Port) },
		},
		"http server port",
	)
	cmd.VarOpt(
		"rpc-port",
		value{
			setFn: func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}
				cfg.Plugins.RPCServer.Port = int(i)
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%d", cfg.Plugins.RPCServer.Port) },
		},
		"rpc server port",
	)
	cmd.VarOpt(
		"etcd",
		value{
			setFn: func(s string) error {
				cfg.Etcd.Endpoints = strings.Split(s, ",")
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%s", cfg.Etcd.Endpoints) },
		},
		"list of etcd endpoints seperated by ,",
	)
	cmd.VarOpt(
		"mount",
		value{
			setFn: func(s string) error {
				v, err := strconv.ParseBool(s)
				if err != nil {
					return err
				}
				cfg.Store.Mount = v
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%t", cfg.Store.Mount) },
		},
		"handle filesystem mounts",
	)
	cmd.VarOpt(
		"move-root",
		value{
			setFn: func(s string) error {
				v, err := strconv.ParseBool(s)
				if err != nil {
					return err
				}
				cfg.Store.MoveRoot = v
				return nil
			},
			getFn: func() string { return fmt.Sprintf("%t", cfg.Store.MoveRoot) },
		},
		"migrate moby created lower path to rootfs",
	)
}

type value struct {
	setFn func(string) error
	getFn func() string
}

func (v value) Set(val string) error { return v.setFn(val) }
func (v value) String() string       { return v.getFn() }
