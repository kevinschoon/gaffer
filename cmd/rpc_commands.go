package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/gosuri/uilive"
	"github.com/gosuri/uitable"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/supervisor"
	"os"
	"time"
)

func toStdout(ch <-chan supervisor.Response) {
	writer := uilive.New()
	writer.Start()
	table := uitable.New()
	table.AddRow("HOST", "PORT", "PID", "UPTIME", "MD5", "ERROR")
	for resp := range ch {
		writer.Flush()
		host := resp.Host
		var (
			hash   string
			pid    int
			uptime time.Duration
		)
		switch {
		case resp.Status != nil:
			pid = resp.Status.Pid
			uptime = resp.Status.Uptime
			hash = resp.Status.Hash
		case resp.Update != nil:
			pid = resp.Update.Pid
			uptime = resp.Update.Uptime
			hash = resp.Update.Hash
		case resp.Restart != nil:
			pid = resp.Restart.Pid
			uptime = resp.Restart.Uptime
			hash = resp.Restart.Hash
		}
		table.AddRow(host.Name, host.Port, pid, uptime, hash, resp.Error)
		fmt.Fprintln(writer, table.String())
	}
	writer.Stop()
}

func toStdoutJSON(ch <-chan supervisor.Response) {
	enc := json.NewEncoder(os.Stdout)
	for resp := range ch {
		maybe(enc.Encode(resp))
	}
}

func parseFilters(hosts []string, ports []int, all bool) []cluster.Filter {
	filters := []cluster.Filter{}
	if all {
		filters = append(filters, cluster.Any())
		return filters
	}
	for _, name := range hosts {
		filters = append(filters, cluster.ByName(name))
	}
	for _, port := range ports {
		filters = append(filters, cluster.ByPort(port))
	}
	return filters
}

func statusCMD(asJSON *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			pattern = cmd.StringOpt("s source", "file://gaffer.json", "gaffer config source")
			hosts   = cmd.StringsOpt("h host", []string{}, "filter by hostname")
			ports   = cmd.IntsOpt("p port", []int{}, "filter by port number")
			all     = cmd.BoolOpt("a all", true, "match all hosts")
		)
		cmd.Action = func() {
			source, err := cluster.NewSource(*pattern)
			maybe(err)
			cfg, err := source.Get()
			maybe(err)
			filters := parseFilters(*hosts, *ports, *all)
			mux := supervisor.ClientMux{supervisor.Clients(cfg.Hosts.Filter(filters...))}
			if *asJSON {
				toStdoutJSON(mux.Status())
			} else {
				toStdout(mux.Status())
			}
		}
	}
}

func applyCMD(asJSON *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			name    = cmd.StringArg("SERVICE", "", "service to apply")
			pattern = cmd.StringOpt("s source", "file://gaffer.json", "gaffer config source")
			hosts   = cmd.StringsOpt("h host", []string{}, "filter by hostname")
			ports   = cmd.IntsOpt("p port", []int{}, "filter by port number")
			all     = cmd.BoolOpt("a all", false, "match all hosts")
		)
		cmd.Spec = "[OPTIONS] SERVICE"
		cmd.Action = func() {
			source, err := cluster.NewSource(*pattern)
			maybe(err)
			cfg, err := source.Get()
			maybe(err)
			svc := cfg.Services.Find(*name)
			if svc == nil {
				maybe(fmt.Errorf("no service named %s", *name))
			}
			filters := parseFilters(*hosts, *ports, *all)
			mux := supervisor.ClientMux{supervisor.Clients(cfg.Hosts.Filter(filters...))}
			if *asJSON {
				toStdoutJSON(mux.Apply(svc))
			} else {
				toStdout(mux.Apply(svc))
			}
		}
	}
}

func restartCMD(asJSON *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			pattern = cmd.StringOpt("s source", "file://gaffer.json", "gaffer config source")
			hosts   = cmd.StringsOpt("h host", []string{}, "filter by hostname")
			ports   = cmd.IntsOpt("p port", []int{}, "filter by port number")
			all     = cmd.BoolOpt("a all", false, "match all hosts")
		)
		cmd.Action = func() {
			source, err := cluster.NewSource(*pattern)
			maybe(err)
			cfg, err := source.Get()
			maybe(err)
			filters := parseFilters(*hosts, *ports, *all)
			mux := supervisor.ClientMux{supervisor.Clients(cfg.Hosts.Filter(filters...))}
			if *asJSON {
				toStdoutJSON(mux.Restart())
			} else {
				toStdout(mux.Restart())
			}
		}
	}
}
