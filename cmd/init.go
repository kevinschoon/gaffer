package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/config"
	"github.com/vektorlab/gaffer/host"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/runc"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/supervisor"
	"os"
)

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			path       = cmd.StringArg("PATH", "/containers", "container init path")
			hard       = cmd.BoolOpt("h hard", false, "fail hard")
			once       = cmd.BoolOpt("o once", false, "run the services only once, synchronously")
			port       = cmd.IntOpt("p port", 10000, "port to listen on")
			configPath = cmd.StringArg("c configPath", "/var/config", "service configuration path")
		)
		cmd.Spec = "[OPTIONS] [PATH]"
		cmd.Action = func() {
			cfg := config.Config{
				Store: config.Store{
					BasePath:   *path,
					ConfigPath: *configPath,
				},
				Runc: config.Runc{},
			}
			db := store.NewFSStore(cfg)
			if *once {
				services, err := db.Services()
				maybe(err)
				for _, svc := range services {
					_, err := runc.New(svc.Id, svc.Bundle, cfg).Run()
					if err != nil {
						if *hard {
							maybe(err)
						} else {
							log.Log.Error(err.Error())
						}
					}
				}
			} else {
				services, err := db.Services()
				maybe(err)
				spv, err := supervisor.New(services, cfg)
				maybe(err)
				maybe(spv.Init())
				name, err := os.Hostname()
				maybe(err)
				maybe(supervisor.NewServer(spv, db, &host.Host{Name: name, Port: int32(*port)}).Listen())
			}
		}
	}
}
