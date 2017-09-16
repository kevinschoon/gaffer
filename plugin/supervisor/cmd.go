package supervisor

import (
	"context"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/util"
)

func (s Supervisor) CLI(cfg *config.Config) cli.CmdInitializer {
	return func(cmd *cli.Cmd) {
		var client RPCClient
		cmd.Before = func() {
			conn, err := util.NewClientConn(*cfg)
			util.Maybe(err)
			client = NewRPCClient(conn)
		}
		cmd.Command("restart", "restart a service", func(cmd *cli.Cmd) {
			cmd.Spec = "ID"
			id := cmd.String(cli.StringArg{
				Name:  "ID",
				Desc:  "service ID to restart",
				Value: "",
			})
			var req *RestartRequest
			cmd.Before = func() {
				req = &RestartRequest{Id: *id}
			}
			cmd.Action = func() {
				resp, err := client.Restart(context.Background(), req, cfg.CallOpts()...)
				util.Maybe(err)
				util.JSONToStdout(resp)
			}
		})
		cmd.Command("status", "return the status of a service", func(cmd *cli.Cmd) {
			cmd.Action = func() {
				resp, err := client.Status(context.Background(), &StatusRequest{}, cfg.CallOpts()...)
				util.Maybe(err)
				util.JSONToStdout(resp)
			}
		})
	}
}
