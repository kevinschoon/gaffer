package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/server"
	"github.com/vektorlab/gaffer/store"
)

func serverCMD(debug *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			pattern   = cmd.StringOpt("p pattern", "0.0.0.0:9090", "ip address and port to bind on")
			dbStr     = cmd.StringOpt("d db", "./gaffer.db", "database connection string")
			anonymous = cmd.BoolOpt("a anonymous", false, "allow anonymous access")
			init      = cmd.BoolOpt("init", false, "initialize the database")
		)
		cmd.Action = func() {
			db, err := store.NewSQLStore(*dbStr, *init)
			maybe(err)
			maybe(
				server.Run(
					server.New(db, *anonymous),
					*pattern,
				),
			)
		}
	}
}
