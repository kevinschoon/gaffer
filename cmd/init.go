package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/config"
	"github.com/vektorlab/gaffer/fatal"
	"github.com/vektorlab/gaffer/host"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/runc"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/supervisor"
	"go.uber.org/zap"
	"os"
)

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			root       = cmd.StringOpt("root", "/run/runc", "runc root path")
			path       = cmd.StringArg("PATH", "/containers", "container init path")
			once       = cmd.BoolOpt("o once", false, "run the services only once, synchronously")
			port       = cmd.IntOpt("p port", 10000, "port to listen on")
			configPath = cmd.StringOpt("c configPath", "/var/mesanine", "service configuration path")
		)
		cmd.Spec = "[OPTIONS] [PATH]"
		cmd.Action = func() {
			cfg := config.Config{
				Store: config.Store{
					BasePath:   *path,
					ConfigPath: *configPath,
				},
				Runc: config.Runc{
					Root: *root,
				},
			}
			db := store.NewFSStore(cfg)
			if *once {
				services, err := db.Services()
				maybe(err)
				for _, svc := range services {
					log.Log.Debug(fmt.Sprintf("launching service %s", svc.Id), zap.Any("service", svc))
					code, err := runc.New(svc.Id, svc.Bundle, cfg).Run()
					log.Log.Info(fmt.Sprintf("service %s exited with code %d", svc.Id, code))
					if code != 0 {
						fatal.Fatal()
					}
					maybe(err)
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
