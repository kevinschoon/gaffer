package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/plugin"
)

func remoteCMD(cfg *config.Config) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[OPTIONS]"
		// Default RPC target is a local socket connection
		// to gaffer running on the host.
		address := cmd.String(cli.StringOpt{
			Name:   "a address",
			Desc:   "RPC server address",
			Value:  config.Default.Address,
			EnvVar: "GAFFER_ADDRESS",
		})
		cmd.Before = func() {
			cfg.Address = *address
		}
		for _, p := range getPlugins(cfg) {
			if c, ok := p.(plugin.CLI); ok {
				cmd.Command(p.Name(), fmt.Sprintf("%s commands", p.Name()), c.CLI(cfg))
			}
		}
	}
}
