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
		store = app.StringOpt("s store", "sqlite://gaffer.db", "store configuration pattern")
	)
	app.Before = func() {
		if *debug {
			log.Debug()
		}
	}
	app.Command("server", "run the Gaffer HTTP server process", serverCMD(*store))
	app.Command("status", "check the status of the cluster", statusCMD(*store))
	app.Command("query", "perform queries", queryCMD(*store))
	app.Command("template", "generate a cluster configuration", templateCMD())
	app.Command("service", "supervise one or more a cluster services", serviceCMD(*store))
	app.Command("init", "initialize the a cluster sqlite store", initCMD())
	maybe(app.Run(os.Args))
}
