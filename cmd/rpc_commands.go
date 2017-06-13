package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/gosuri/uilive"
	"github.com/gosuri/uitable"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/supervisor"
	"io/ioutil"
	"os"
	"time"
)

func getConfig(path string) *cluster.Config {
	raw, err := ioutil.ReadFile(path)
	maybe(err)
	config := &cluster.Config{}
	maybe(json.Unmarshal(raw, config))
	return config
}

func toStdout(ch <-chan supervisor.Response) {
	writer := uilive.New()
	writer.Start()
	table := uitable.New()
	table.AddRow("HOST", "PORT", "PID", "UPTIME", "ERROR")
	for resp := range ch {
		writer.Flush()
		host := resp.Host
		var (
			pid    int
			uptime time.Duration
		)
		switch {
		case resp.Status != nil:
			pid = resp.Status.Pid
			uptime = resp.Status.Uptime
		case resp.Update != nil:
			pid = resp.Update.Pid
			uptime = resp.Update.Uptime
		case resp.Restart != nil:
			pid = resp.Restart.Pid
			uptime = resp.Restart.Uptime
		}
		table.AddRow(host.Name, host.Port, pid, uptime, resp.Error)
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

func statusCMD(asJSON *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			path    = cmd.StringOpt("c config", "gaffer.json", "gaffer cluster config file")
			pattern = cmd.StringOpt("p pattern", ".*", "regular expression matching hostname")
		)
		cmd.Action = func() {
			cfg := getConfig(*path)
			mux := supervisor.ClientMux{supervisor.Clients(cfg.Hosts.Match(cluster.ByName(*pattern)))}
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
			path    = cmd.StringOpt("c config", "gaffer.json", "gaffer cluster config file")
			pattern = cmd.StringOpt("p pattern", ".*", "regular expression matching hostname")
		)
		cmd.Spec = "[OPTIONS] SERVICE"
		cmd.Action = func() {
			cfg := getConfig(*path)
			svc := cfg.Services.Find(*name)
			if svc == nil {
				maybe(fmt.Errorf("no service named %s", *name))
			}
			mux := supervisor.ClientMux{supervisor.Clients(cfg.Hosts.Match(cluster.ByName(*pattern)))}
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
			path    = cmd.StringOpt("c config", "gaffer.json", "gaffer cluster config file")
			pattern = cmd.StringOpt("p pattern", ".*", "regular expression matching hostname")
		)
		cmd.Action = func() {
			cfg := getConfig(*path)
			mux := supervisor.ClientMux{supervisor.Clients(cfg.Hosts.Match(cluster.ByName(*pattern)))}
			if *asJSON {
				toStdoutJSON(mux.Restart())
			} else {
				toStdout(mux.Restart())
			}
		}
	}
}
