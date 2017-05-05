package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/user"
	"io/ioutil"
	"os"
)

func queryCMD() func(*cli.Cmd) {
	var c *client.Client
	return func(cmd *cli.Cmd) {
		var (
			endpoint = cmd.StringOpt("e endpoint", "http://localhost:9090", "gaffer API server")
			auth     = cmd.StringOpt("u user", "", "user:pass basic auth string")
		)
		cmd.Before = func() {
			var usr *user.User
			if *auth != "" {
				u, err := user.FromString(*auth)
				maybe(err)
				usr = u
			}
			c = client.New(*endpoint, usr)
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
				resp, err := c.Query(&query.Query{
					Type:   query.CREATE,
					Create: &query.Create{Clusters: []*cluster.Cluster{config}},
				})
				maybe(err)
				maybe(json.NewEncoder(os.Stdout).Encode(resp))
			}
		})
		cmd.Command("read", "read cluster configuration", func(cmd *cli.Cmd) {
			var clusterID = cmd.StringOpt("i id", "", "cluster id")
			cmd.Action = func() {
				resp, err := c.Query(&query.Query{Type: query.READ, Read: &query.Read{ID: *clusterID}})
				maybe(err)
				maybe(json.NewEncoder(os.Stdout).Encode(resp))
			}
		})
		cmd.Command("update", "update an existing cluster", func(cmd *cli.Cmd) {
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
				resp, err := c.Query(&query.Query{
					Type:   query.UPDATE,
					Update: &query.Update{Clusters: []*cluster.Cluster{config}},
				})
				maybe(err)
				maybe(json.NewEncoder(os.Stdout).Encode(resp))
			}
		})
		cmd.Command("delete", "delete a cluster", func(cmd *cli.Cmd) {
			var clusterID = cmd.StringOpt("i id", "", "cluster id")
			cmd.Spec = "--id"
			cmd.Action = func() {
				resp, err := c.Query(&query.Query{Type: query.DELETE, Delete: &query.Delete{ID: *clusterID}})
				maybe(err)
				maybe(json.NewEncoder(os.Stdout).Encode(resp))
			}
		})
	}
}
