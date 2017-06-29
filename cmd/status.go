package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/host"
	"github.com/vektorlab/gaffer/supervisor"
	"golang.org/x/net/context"
	"os"
)

func statusCMD(asJSON *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			pattern = cmd.StringOpt("s source", "gaffer://localhost:10000", "gaffer RPC server")
		)
		cmd.Action = func() {
			h, err := host.New(*pattern)
			maybe(err)
			conn, err := supervisor.NewClient(*h)
			maybe(err)
			defer conn.Close()
			resp, err := supervisor.NewSupervisorClient(conn).Status(context.Background(), &supervisor.StatusRequest{})
			maybe(err)
			maybe(json.NewEncoder(os.Stdout).Encode(resp))
		}
	}
}
