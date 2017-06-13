package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/supervisor"
)

func superviseCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			dbPath = cmd.StringOpt("d db", "gaffer.db", "path to service db")
			token  = cmd.StringOpt("t token", "", "secret token")
			port   = cmd.IntOpt("p port", 10000, "port to listen on")
		)
		cmd.Action = func() {
			db, err := store.New(*dbPath)
			maybe(err)
			sup, err := supervisor.NewSupervisor(supervisor.Token(*token), db)
			maybe(err)
			maybe(supervisor.Launch(sup, *port))
		}
	}
}
