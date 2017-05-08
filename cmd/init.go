package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/store"
)

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			name  = cmd.StringOpt("n name", "gaffer", "cluster name")
			dbStr = cmd.StringOpt("d db", "./gaffer.db", "database connection string")
		)
		cmd.Action = func() {
			_, err := store.NewSQLStore(*name, *dbStr, true)
			maybe(err)
		}
	}
}
