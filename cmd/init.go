package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/store"
)

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			dbStr = cmd.StringOpt("d db", "./gaffer.db", "database connection string")
		)
		cmd.Action = func() {
			_, err := store.NewSQLStore(*dbStr, true)
			maybe(err)
		}
	}
}
