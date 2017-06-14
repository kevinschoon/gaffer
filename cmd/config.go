package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/gosuri/uitable"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster"
	"io/ioutil"
	"os"
	"strings"
)

func servicesToStdout(services []*cluster.Service) {
	table := uitable.New()
	table.AddRow("ID", "ARGS", "MD5")
	for _, service := range services {
		table.AddRow(service.ID, strings.Join(service.Args, " "), service.Hash())
	}
	fmt.Println(table)
}

func hostsToStdout(hosts []*cluster.Host) {
	table := uitable.New()
	table.AddRow("HOST", "PORT")
	for _, host := range hosts {
		table.AddRow(host.Name, host.Port)
	}
	fmt.Println(table)
}

func getCMD(asJSON *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			pattern      = cmd.StringOpt("s source", "file://gaffer.json", "gaffer config source")
			showHosts    = cmd.BoolOpt("hosts", false, "show hosts")
			showServices = cmd.BoolOpt("services", true, "show services")
		)
		cmd.Action = func() {
			source, err := cluster.NewSource(*pattern)
			maybe(err)
			cfg, err := source.Get()
			maybe(err)
			if *asJSON {
				maybe(json.NewEncoder(os.Stdout).Encode(cfg))
				return
			}
			if *showHosts {
				hostsToStdout(cfg.Hosts)
				if *showServices {
					fmt.Println("-----")
				}
			}
			if *showServices {
				servicesToStdout(cfg.Services)
			}
		}
	}
}

func setCMD(asJSON *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			path    = cmd.StringArg("PATH", "gaffer.json", "path to config")
			pattern = cmd.StringOpt("s source", "file://gaffer.json", "gaffer config source")
		)
		cmd.Spec = "[OPTIONS] PATH"
		cmd.Action = func() {
			source, err := cluster.NewSource(*pattern)
			maybe(err)
			raw, err := ioutil.ReadFile(*path)
			maybe(err)
			cfg := &cluster.Config{}
			maybe(json.Unmarshal(raw, cfg))
			maybe(source.Set(cfg))
		}
	}
}

func configCMD(asJSON *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command("get", "read the cluster configuration", getCMD(asJSON))
		cmd.Command("set", "set the cluster configuration", setCMD(asJSON))
	}
}
