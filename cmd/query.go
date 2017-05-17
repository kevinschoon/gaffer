package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/store/query"
	//"github.com/vektorlab/gaffer/user"
	"io/ioutil"
	"os"
)

func queryCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[-d]"
		var (
			//endpoint = cmd.StringOpt("e endpoint", "http://localhost:9090", "gaffer API server")
			//auth     = cmd.StringOpt("u user", "", "user:pass basic auth string")
			dbStr = cmd.StringOpt("d db", "./gaffer.db", "database connection string")
		)
		var db store.Store
		cmd.Before = func() {
			/*
				var usr *user.User
				if *auth != "" {
					u, err := user.FromString(*auth)
					maybe(err)
					usr = u
				}
				fmt.Println(dbStr)
				c = store.NewHTTPStore(*endpoint, usr)
			*/
			d, err := store.NewSQLStore("", *dbStr, false)
			maybe(err)
			db = d
		}
		cmd.Command("create", "create a new cluster", func(cmd *cli.Cmd) {
			var path = cmd.StringArg("PATH", "", "path to the cluster configuration file, use \"-\" to read from stdin")
			cmd.Action = func() {
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
				resp, err := db.Query(&query.Query{Read: &query.Read{}})
				maybe(err)
				maybe(json.NewEncoder(os.Stdout).Encode(resp))
			}
		})
	}
}
