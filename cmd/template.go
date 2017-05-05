package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/store/query"
	"os"
)

func templateCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			name    = cmd.StringOpt("n name", "my-cluster", "cluster name")
			size    = cmd.IntOpt("s size", 3, "cluster size")
			asQuery = cmd.BoolOpt("q query", false, "generate a gaffer API query")
		)
		cmd.Action = func() {
			tmpl := cluster.New(*name, *size)
			for i := 0; i < *size; i++ {
				tmpl.Hosts = append(tmpl.Hosts, cluster.NewHost())
			}
			if *asQuery {
				q := query.Query{
					Type: query.CREATE,
					Create: &query.Create{
						Clusters: []*cluster.Cluster{
							tmpl,
						},
					},
				}
				raw, err := json.Marshal(q)
				maybe(err)
				os.Stdout.Write(raw)
			} else {
				raw, err := json.Marshal(tmpl)
				maybe(err)
				os.Stdout.Write(raw)
			}
		}
	}
}