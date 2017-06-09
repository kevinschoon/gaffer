package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/supervisor"
)

func serviceCMD(sp string) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			service = cmd.StringOpt("s service", "", "service to supervise")
		)
		cmd.Action = func() {
			db, err := store.NewStore(sp)
			maybe(err)
			maybe(supervisor.Launch(db, *service))
		}
	}
}
