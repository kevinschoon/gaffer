package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/plugin"
	"github.com/mesanine/gaffer/plugin/logger"
	"github.com/mesanine/gaffer/plugin/metrics"
	"github.com/mesanine/gaffer/plugin/register"
	"github.com/mesanine/gaffer/plugin/supervisor"
	"github.com/mesanine/gaffer/util"
	"github.com/mesanine/gaffer/version"

	"os"
)

func Run() {
	cfg := &config.Config{}
	app := cli.App("gaffer", "Distributed Init System")
	app.Spec = "[OPTIONS]"
	var configPath = app.StringOpt("c config", "", "Path to a gaffer.json configuration file")
	app.Version("version", fmt.Sprintf("version=%s\ngitsha=%s", version.Version, version.GitSHA))
	enabled := app.Strings(cli.StringsOpt{
		Name:   "p plugin",
		Desc:   "Toggled plugins",
		Value:  config.Default.EnabledPlugins,
		EnvVar: "GAFFER_ENABLED_PLUGINS",
	})
	disabled := app.Strings(cli.StringsOpt{
		Name:   "disable",
		Desc:   "Disable the specified plugin",
		Value:  config.Default.DisabledPlugins,
		EnvVar: "GAFFER_DISABLED_PLUGINS",
	})
	device := app.String(cli.StringOpt{
		Name:   "d device",
		Desc:   "Send log output to a block device",
		Value:  config.Default.Logger.Device,
		EnvVar: "GAFFER_LOGGER_DEVICE",
	})
	logDir := app.String(cli.StringOpt{
		Name:   "log-dir",
		Desc:   "Send log output to files in this directory",
		Value:  config.Default.Logger.LogDir,
		EnvVar: "GAFFER_LOGGER_DIRECTORY",
	})
	maxLogSize := app.Int(cli.IntOpt{
		Name:   "max-log-size",
		Desc:   "Maximum log file size in mb",
		Value:  config.Default.Logger.MaxSize,
		EnvVar: "GAFFER_LOGGER_MAX_SIZE",
	})
	maxBackups := app.Int(cli.IntOpt{
		Name:   "max-backups",
		Desc:   "Maximum number of backups to rotate",
		Value:  config.Default.Logger.MaxBackups,
		EnvVar: "GAFFER_LOGGER_MAX_BACKUPS",
	})
	compress := app.Bool(cli.BoolOpt{
		Name:   "compress",
		Desc:   "Compress rotated log files",
		Value:  config.Default.Logger.Compress,
		EnvVar: "GAFFER_LOGGER_COMPRESS",
	})
	debug := app.Bool(cli.BoolOpt{
		Name:   "debug",
		Desc:   "Output debugging information",
		Value:  config.Default.Logger.Debug,
		EnvVar: "GAFFER_LOGGER_DEBUG",
	})
	endpoints := app.Strings(cli.StringsOpt{
		Name:   "e endpoints",
		Desc:   "Etcd endpoint",
		Value:  config.Default.Endpoints,
		EnvVar: "GAFFER_ENDPOINT",
	})
	app.Before = func() {
		if *configPath != "" {
			util.Maybe(config.Load(*configPath, cfg))
		}
		cfg.Endpoints = *endpoints
		cfg.EnabledPlugins = *enabled
		cfg.DisabledPlugins = *disabled
		cfg.Logger.Device = *device
		cfg.Logger.LogDir = *logDir
		cfg.Logger.MaxSize = *maxLogSize
		cfg.Logger.MaxBackups = *maxBackups
		cfg.Logger.Compress = *compress
		cfg.Logger.Debug = *debug
		// Initialize the logger
		util.Maybe(log.Setup(*cfg))
	}
	app.Command("init", "Bootstrap the operating system", initCMD(cfg))
	app.Command("launch", "Launch plugin subsystems", launchCMD(cfg))
	app.Command("hosts", "List remote hosts", hostsCMD(cfg))
	app.Command("remote", "Make RPC calls against a host", remoteCMD(cfg))
	util.Maybe(app.Run(os.Args))
}

func getPlugins(cfg *config.Config) []plugin.Plugin {
	plugins := []plugin.Plugin{}
	for _, p := range cfg.Plugins() {
		switch p {
		case "logger":
			plugins = append(plugins, logger.New())
		case "metrics":
			plugins = append(plugins, metrics.New())
		case "supervisor":
			plugins = append(plugins, supervisor.New())
		case "register":
			plugins = append(plugins, register.New())
		default:
			util.Maybe(fmt.Errorf("unknown plugin: %s", p))
		}
	}
	return plugins
}

func allPlugins() []plugin.Plugin {
	return []plugin.Plugin{logger.New(), metrics.New(), supervisor.New(), register.New()}
}
