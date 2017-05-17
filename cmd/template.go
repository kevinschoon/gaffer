package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/operator/mock"
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
			tmpl.Services = mock.Mock{20}.Update(tmpl)
			if *asQuery {
				q := query.Query{
					Create: &query.Create{tmpl},
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
