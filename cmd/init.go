package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/host"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/supervisor"
	"os"
)

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			path = cmd.StringArg("PATH", "/containers", "container init path")
			port = cmd.IntOpt("p port", 10000, "port to listen on")
		)
		cmd.Spec = "[OPTIONS] [PATH]"
		cmd.Action = func() {
			db := store.NewFSStore(*path)
			spv, err := supervisor.New(db)
			maybe(err)
			spv.Init()
			name, err := os.Hostname()
			maybe(err)
			maybe(supervisor.NewServer(spv, db, &host.Host{Name: name, Port: int32(*port)}).Listen())
		}
	}
}
