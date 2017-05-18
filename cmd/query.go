package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/store/query"
	"io/ioutil"
	"os"
)

func queryCMD(sp string) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command("create", "create a new cluster", func(cmd *cli.Cmd) {
			var path = cmd.StringArg("PATH", "", "path to the cluster configuration file, use \"-\" to read from stdin")
			cmd.Action = func() {
				db, err := store.NewStore(sp)
				maybe(err)
				var config *cluster.Cluster
				if *path == "-" {
					config = &cluster.Cluster{}
					maybe(json.NewDecoder(os.Stdin).Decode(config))
				} else {
					raw, err := ioutil.ReadFile(*path)
					maybe(err)
					config = &cluster.Cluster{}
					maybe(json.Unmarshal(raw, config))
				}
				resp, err := db.Query(&query.Query{
					Create: &query.Create{config},
				})
				maybe(err)
				maybe(json.NewEncoder(os.Stdout).Encode(resp))
			}
		})
		cmd.Command("read", "read cluster configuration", func(cmd *cli.Cmd) {
			cmd.Action = func() {
				db, err := store.NewStore(sp)
				maybe(err)
				resp, err := db.Query(&query.Query{Read: &query.Read{}})
				maybe(err)
				maybe(json.NewEncoder(os.Stdout).Encode(resp))
			}
		})
	}
}
