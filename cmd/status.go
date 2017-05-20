package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/store"
	"os"
)

func statusCMD(sp string) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			db, err := store.NewStore(sp)
			maybe(err)
			processes, err := client.NewClient(db).Processes()
			maybe(err)
			json.NewEncoder(os.Stdout).Encode(processes)
		}
	}
}
