package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/version"
	"go.uber.org/zap"
	"os"
)

func maybe(err error) {
	if err != nil {
		if log.Log != nil {
			log.Log.Error("gaffer encountered an un-recoverable error", zap.Error(err))
		}
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func Run() {
	app := cli.App("gaffer", "Distributed Init System")
	app.Spec = "[OPTIONS]"
	app.Version("version", fmt.Sprintf("version=%s\ngitsha=%s", version.Version, version.GitSHA))
	var (
		configPath = app.StringOpt("c config", "/etc/gaffer.json", "path to a gaffer.json file")
	)
	cfg := config.New()
	config.SetCLIOpts(app, cfg)
	app.Before = func() {
		if *configPath != "" {
			maybe(config.Load(*configPath, cfg))
		}
		// Initialize the logger
		maybe(log.Setup(*cfg))
	}
	app.Command("init", "launch the operating system", initCMD(cfg))
	app.Command("hosts", "list remote Gaffer hosts", hostsCMD(cfg))
	app.Command("status", "list the status of a remote host", statusCMD())
	app.Command("restart", "restart a remote service", restartCMD())
	// Allow Gaffer to be run in the "multi-call"
	// style of busybox where executables are symlinked
	// like /sbin/init --> /bin/busybox.
	/*
		args := []string{"gaffer"}
		_, exe := filepath.Split(os.Args[0])
		switch exe {
		case "init":
			args = append(args, "init")
			for _, arg := range os.Args[1:] {
				args = append(args, arg)
			}
			maybe(app.Run(args))
			return
		}
	*/
	maybe(app.Run(os.Args))
}
