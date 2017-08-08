package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/fatal"
	"github.com/mesanine/gaffer/host"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/runc"
	"github.com/mesanine/gaffer/store"
	"github.com/mesanine/gaffer/supervisor"
	"os"
)

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			path       = cmd.StringArg("PATH", "/containers", "container init path")
			root       = cmd.StringOpt("root", "/run/runc", "runc root path")
			once       = cmd.BoolOpt("o once", false, "run the services only once, synchronously")
			port       = cmd.IntOpt("p port", 10000, "port to listen on")
			mount      = cmd.BoolOpt("m mounts", true, "handle overlay mounts")
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
					Root:  *root,
					Mount: *mount,
				},
			}
			db := store.NewFSStore(cfg)
			if *once {
				log.Log.Info(fmt.Sprintf("starting on-boot services from %s", *path))
				services, err := db.Services()
				maybe(err)
				for _, svc := range services {
					ro, err := svc.ReadOnly()
					maybe(err)
					log.Log.Info(fmt.Sprintf("starting on-boot service %s", svc.Id))
					code, err := runc.New(svc.Id, svc.Bundle, ro, cfg).Run()
					log.Log.Info(fmt.Sprintf("on-boot service %s exited with code %d", svc.Id, code))
					if code != 0 {
						fatal.Fatal()
					}
					maybe(err)
				}
			} else {
				log.Log.Info(fmt.Sprintf("starting long-running services from %s", *path))
				services, err := db.Services()
				maybe(err)
				log.Log.Info(fmt.Sprintf("creating supervisor for %d services", len(services)))
				spv, err := supervisor.New(services, cfg)
				maybe(err)
				spv.Init()
				name, err := os.Hostname()
				maybe(err)
				maybe(supervisor.NewServer(spv, db, &host.Host{Name: name, Port: int32(*port)}).Listen())
			}
		}
	}
}
