package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/supervisor"
)

func superviseCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "-s"
		var (
			service = cmd.StringOpt("s service", "", "service names to supervise, e.g. svc1,svc2")
		)
		cmd.Action = func() {
			db, err := store.NewSQLStore("", "./gaffer.db", false)
			maybe(err)
			maybe(supervisor.Launch(db, *service))
		}
	}
}
