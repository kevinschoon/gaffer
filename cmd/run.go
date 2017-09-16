package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/plugin"
	"github.com/mesanine/gaffer/plugin/register"
	"github.com/mesanine/gaffer/plugin/supervisor"
	"github.com/mesanine/gaffer/store"
	"github.com/mesanine/gaffer/util"
	"github.com/mesanine/ginit"
)

func runCMD(cfg *config.Config) cli.CmdInitializer {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[OPTIONS]"
		configPath := cmd.String(cli.StringOpt{
			Name:   "config-path",
			Desc:   "service configuration path",
			Value:  config.Default.Store.ConfigPath,
			EnvVar: "GAFFER_STORE_CONFIG_PATH",
		})
		basePath := cmd.String(cli.StringOpt{
			Name:   "store-path",
			Desc:   "container store path",
			Value:  config.Default.Store.BasePath,
			EnvVar: "GAFFER_STORE_PATH",
		})
		runcRoot := cmd.String(cli.StringOpt{
			Name:   "runc-root",
			Desc:   "runc root path",
			Value:  config.Default.RuncRoot,
			EnvVar: "GAFFER_RUNC_ROOT",
		})

		mount := cmd.Bool(cli.BoolOpt{
			Name:   "mount",
			Desc:   "handle filesystem mounts",
			Value:  config.Default.Store.Mount,
			EnvVar: "GAFFER_STORE_MOUNT",
		})
		moveRoot := cmd.Bool(cli.BoolOpt{
			Name:   "move-root",
			Desc:   "migrate moby created lower path to rootfs",
			Value:  config.Default.Store.MoveRoot,
			EnvVar: "GAFFER_STORE_MOVE_ROOT",
		})
		cmd.Before = func() {
			cfg.RuncRoot = *runcRoot
			cfg.Store.ConfigPath = *configPath
			cfg.Store.BasePath = *basePath
			cfg.Store.Mount = *mount
			cfg.Store.MoveRoot = *moveRoot
		}
		cmd.Action = func() {
			log.Log.Info("starting onboot services")
			// Launch any containers synchronously
			// that exist in directory "onboot" in
			// the store root.
			db := store.New(*cfg, "onboot")
			util.Maybe(db.Init())
			util.Maybe(supervisor.Once(*cfg, db))
			util.Maybe(db.Close())
			log.Log.Info("onboot services finished")
			// Launch all system services from the
			// "services" path in the store root.
			db = store.New(*cfg, "services")
			util.Maybe(db.Init())
			reg := plugin.NewRegistry()
			reg.Register(&register.Server{})
			reg.Register(supervisor.New())
			util.Maybe(reg.Configure(*cfg))
			server, err := plugin.NewServer(*cfg)
			util.Maybe(err)
			go func() {
				util.Maybe(server.Run(reg))
			}()
			go func() {
				util.Maybe(reg.Run())
			}()
			util.Maybe(ginit.Init(reg, server))
			util.Maybe(db.Close())
		}
	}
}
