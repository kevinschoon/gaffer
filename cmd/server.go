package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/server"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/user"
)

func serverCMD(debug *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			pattern = cmd.StringOpt("p pattern", "0.0.0.0:9090", "ip address and port to bind on")
			dbStr   = cmd.StringOpt("d db", "./gaffer.db", "database connection string")
			userStr = cmd.StringOpt("u user", "", "user/token combo")
		)
		cmd.Action = func() {
			db, err := store.NewSQLStore("", *dbStr, false)
			maybe(err)
			var usr *user.User
			if *userStr != "" {
				usr, err = user.FromString(*userStr)
				maybe(err)
			}
			maybe(
				server.Run(
					server.New(db, usr),
					*pattern,
				),
			)
		}
	}
}
