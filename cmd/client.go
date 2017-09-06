package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/client"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/host"
	rpc "github.com/mesanine/gaffer/plugin/rpc-server"
	"os"
)

func hostsCMD(cfg *config.Config) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[OPTIONS]"
		var (
		//etcdSrvs = cmd.StringOpt("etcd", "http://localhost:2379", "list of etcd endpoints seperated by ,")
		)
		cmd.Action = func() {
			cli, err := client.New(*cfg)
			maybe(err)
			hosts, err := cli.Hosts()
			maybe(err)
			maybe(json.NewEncoder(os.Stdout).Encode(hosts))
		}
	}
}

func statusCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[OPTIONS]"
		var (
			hostAddr = cmd.StringOpt("host-addr", "127.0.0.1", "host ip address")
			hostPort = cmd.IntOpt("host-port", 10000, "host rpc port")
		)
		cmd.Action = func() {
			cli := client.Client{}
			resp, err := cli.Status(&rpc.StatusRequest{Host: &host.Host{
				Address: *hostAddr,
				Port:    int32(*hostPort),
			}})
			maybe(err)
			maybe(json.NewEncoder(os.Stdout).Encode(resp))
		}
	}
}

func restartCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			id       = cmd.StringArg("ID", "", "service identifier")
			hostAddr = cmd.StringOpt("host-addr", "127.0.0.1", "host ip address")
			hostPort = cmd.IntOpt("host-port", 10000, "host rpc port")
		)
		cmd.Spec = "[OPTIONS] ID"
		cmd.Action = func() {
			cli := client.Client{}
			resp, err := cli.Restart(&rpc.RestartRequest{
				Id: *id,
				Host: &host.Host{
					Address: *hostAddr,
					Port:    int32(*hostPort),
				}})
			maybe(err)
			maybe(json.NewEncoder(os.Stdout).Encode(resp))
		}
	}
}
