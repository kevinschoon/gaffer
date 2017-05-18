package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/server"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/user"
)

func serverCMD(sp string) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			pattern = cmd.StringOpt("p pattern", "0.0.0.0:9090", "ip address and port to bind on")
			userStr = cmd.StringOpt("u user", "", "user/token combo")
		)
		cmd.Action = func() {
			db, err := store.NewStore(sp)
			maybe(err)
			var usr *user.User
			if *userStr != "" {
				u, err := user.FromString(*userStr)
				maybe(err)
				usr = u
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
