package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/supervisor"
	"net/rpc"
)

func statusCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			client, err := rpc.DialHTTP("tcp", "localhost:9091")
			maybe(err)
			resp := &supervisor.StatusResponse{}
			maybe(client.Call("Supervisor.Status", supervisor.StatusRequest{}, resp))
			fmt.Println(resp)
			restartResp := &supervisor.RestartResponse{}
			maybe(client.Call("Supervisor.Restart", supervisor.RestartRequest{}, restartResp))
			fmt.Println(restartResp)
			updateResp := &supervisor.UpdateResponse{}
			svc := &service.Service{
				ID:          "fuu",
				Args:        []string{"sleep"},
				Environment: []*service.Env{&service.Env{"fuu", "bar"}},
			}
			maybe(client.Call("Supervisor.Update", supervisor.UpdateRequest{svc}, updateResp))
			fmt.Println(updateResp)
		}
	}
}
