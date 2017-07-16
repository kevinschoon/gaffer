package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/version"
	"go.uber.org/zap"
	"os"
)

func maybe(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func Run() {
	app := cli.App("gaffer", "Distributed Init System")
	app.Spec = "[OPTIONS]"
	app.Version("version", fmt.Sprintf("%s (%s)", version.Version, version.GitSHA))
	var (
		json    = app.BoolOpt("json", false, "enables json log output")
		debug   = app.BoolOpt("d debug", false, "output debug information")
		logPath = app.StringOpt("l log", "", "log output to a file")
	)
	app.Before = func() {
		// Enable JSON logging output
		if *json {
			log.Json()
		}
		// Enable debugging with dev features
		if *debug {
			log.Level.SetLevel(zap.DebugLevel)
			log.Debug()
		}
		// Change log output from stderr to a file
		if *logPath != "" {
			log.Output(*logPath)
		}
	}
	app.Command("config", "modify a cluster config", configCMD(json))
	app.Command("init", "launch the Gaffer init process", initCMD())
	app.Command("restart", "restart a local service", restartCMD(json))
	app.Command("server", "run a gaffer HTTP proxy and UI", serverCMD())
	app.Command("status", "output the status of local services", statusCMD(json))
	app.Command("wait", "wait for a file to exist", waitCMD())

	maybe(app.Run(os.Args))
}
