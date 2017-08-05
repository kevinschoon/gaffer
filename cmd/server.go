package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/host"
	"github.com/mesanine/gaffer/server"
)

func serverCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			sourcePtrn = cmd.StringOpt("s source", "gaffer://localhost:10000", "gaffer config source")
			pattern    = cmd.StringOpt("p pattern", "0.0.0.0:9090", "interface and port to listen on")
			userStr    = cmd.StringOpt("u user", "", "user:pass combination")
		)
		cmd.Action = func() {
			cfg := config.Config{
				Server: config.Server{
					Pattern: *pattern,
				},
				User: config.User{
					User: *userStr,
				},
			}
			source, err := host.NewSource(*sourcePtrn)
			maybe(err)
			svr, err := server.New(source, cfg)
			maybe(err)
			maybe(server.Run(svr))
		}
	}
}
