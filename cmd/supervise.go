package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/supervisor"
	"github.com/vektorlab/gaffer/user"
)

func superviseCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			service   = cmd.StringOpt("s service", "", "name of the service to supervise")
			clusterID = cmd.StringOpt("c cluster", "", "name of the cluster")
			endpoint  = cmd.StringOpt("e endpoint", "http://localhost:9090", "gaffer API server")
			auth      = cmd.StringOpt("u user", "", "user:pass basic auth string")
		)
		cmd.Action = func() {
			var usr *user.User
			if *auth != "" {
				u, err := user.FromString(*auth)
				maybe(err)
				usr = u
			}
			if *clusterID == "" {
				maybe(fmt.Errorf("must specify cluster ID"))
			}
			if *service == "" {
				maybe(fmt.Errorf("must specify service name"))
			}
			supervisor.Run(
				supervisor.Opts{
					Client:    client.New(*endpoint, usr),
					ClusterID: *clusterID,
					Service:   *service,
				},
			)
		}
	}
}
