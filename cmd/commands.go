package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/version"
	"os"
)

func maybe(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func Run() {
	app := cli.App("gaffer", "Mesos Init System")
	app.Spec = "[OPTIONS]"
	app.Version("version", fmt.Sprintf("%s (%s)", version.Version, version.GitSHA))
	var (
		debug = app.BoolOpt("d debug", false, "output debug information")
	)
	app.Before = func() {
		if *debug {
			log.Debug()
		}
	}
	app.Command("server", "Run the scheduler HTTP server", serverCMD(debug))
	app.Command("query", "Perform HTTP queries", queryCMD())
	app.Command("template", "Generate a configuration template", templateCMD())
	app.Command("supervise", "Supervise a cluster process", superviseCMD())
	app.Command("init", "Initialize the cluster database", initCMD())
	maybe(app.Run(os.Args))
}
