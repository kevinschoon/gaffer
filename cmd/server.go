package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/server"
	"github.com/vektorlab/gaffer/user"
)

func serverCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			sourcePtrn = cmd.StringOpt("c config", "file://gaffer.json", "gaffer config source")
			pattern    = cmd.StringOpt("p pattern", "0.0.0.0:9090", "interface and port to listen on")
			userStr    = cmd.StringOpt("u user", "", "user:pass combination")
		)
		cmd.Action = func() {
			var usr *user.User
			if *userStr != "" {
				u, err := user.FromString(*userStr)
				maybe(err)
				usr = u
			}
			source, err := cluster.NewSource(*sourcePtrn)
			maybe(err)
			maybe(server.Run(server.New(source, usr), *pattern))
		}
	}
}
