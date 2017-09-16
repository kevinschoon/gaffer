package cmd

import (
	"encoding/json"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/client"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/util"
	"os"
)

func hostsCMD(cfg *config.Config) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[OPTIONS]"
		cmd.Action = func() {
			cli, err := client.New(*cfg)
			util.Maybe(err)
			hosts, err := cli.Hosts()
			util.Maybe(err)
			util.Maybe(json.NewEncoder(os.Stdout).Encode(hosts))
		}
	}
}
