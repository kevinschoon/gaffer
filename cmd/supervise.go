package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/supervisor"
	"github.com/vektorlab/gaffer/user"
	"strings"
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
			var u *user.User
			if *auth != "" {
				split := strings.Split(*auth, ":")
				if len(split) != 2 {
					maybe(fmt.Errorf("bad auth %s", *auth))
				}
				u = &user.User{split[0], split[1]}
			}
			if *clusterID == "" {
				maybe(fmt.Errorf("must specify cluster ID"))
			}
			if *service == "" {
				maybe(fmt.Errorf("must specify service name"))
			}
			supervisor.Run(
				supervisor.Opts{
					Client:    client.New(*endpoint, u),
					ClusterID: *clusterID,
					Service:   *service,
				},
			)
		}
	}
}
