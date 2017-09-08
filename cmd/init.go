package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/plugin"
	"github.com/mesanine/gaffer/plugin/supervisor"
	"github.com/mesanine/gaffer/store"
	"github.com/mesanine/ginit"
	"go.uber.org/zap"
	"os"
)

const (
	PRE_INIT Stage = "PRE_INIT"
	INIT     Stage = "INIT"
)

type Stage string

func (s *Stage) Set(v string) error {
	switch v {
	case "PRE_INIT":
		*s = Stage(v)
		return nil
	case "INIT":
		*s = Stage(v)
		return nil
	}
	return fmt.Errorf("unknown boot stage: %s", v)
}

func (s Stage) String() string { return string(s) }

func initCMD(cfg *config.Config) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		stage := Stage("PRE_INIT")
		var (
			recovery = cmd.BoolOpt("r recovery", false, "launch a recovery shell")
		)
		cmd.VarOpt("stage", &stage, "init boot stage [PRE_INIT, INIT]")
		cmd.Spec = "[OPTIONS]"
		config.SetInitOpts(cmd, cfg)
		cmd.Action = func() {
			// TODO: Need a clean way to initialize the OS
			// in "stages" but I am not sure what the best
			// abstraction is to do this at the moment.
			switch stage {
			case PRE_INIT:
				if !ginit.IsRoot() {
					maybe(fmt.Errorf("init can only be run as root"))
				}
				// Check if we are running on a memory-based
				// file system. If we are we need to switch-root
				// and re-mount as tempfs.
				isMem, err := ginit.IsMemFS("/")
				maybe(err)
				if isMem {
					// Only supporting tempfs for now
					maybe(ginit.Mount(ginit.TmpFS(cfg.Init.NewRoot, 0)))
					opts, err := ginit.NewSwitchOptions(cfg.Init.NewRoot)
					maybe(err)
					log.Log.Info(fmt.Sprintf("calling switch-root (%s)", cfg.Init.NewRoot))
					maybe(ginit.SwitchRoot(*opts))
					log.Log.Info("switch-root completed successfully")
				} else {
					log.Log.Info("rootfs is not memory-based, will not switch-root")
				}
				if *recovery {
					log.Log.Info("dropping into single user mode")
					maybe(ginit.Exec("/bin/sh"))
				}
				// Load the procfs file system
				log.Log.Info("Loading procfs")
				maybe(ginit.Mount(ginit.MountArgs{
					Source: "proc",
					Target: "/proc",
					FSType: "proc",
					Flags:  0,
					Data:   "nodev,nosuid,noexec,relatime"}))
				// Load devfs filesystem
				log.Log.Info("Loading devfs")
				maybe(ginit.Mount(
					ginit.MountArgs{
						Source: "dev",
						Target: "/dev",
						FSType: "devtmpfs",
						Data:   "nosuid,noexec,relatime,size=10m,nr_inodes=248418,mode=755",
					}))
				log.Log.Info(fmt.Sprintf("calling init helper script: %s", cfg.Init.Helper))
				// Call the init helper script which does most of the
				// heavy lifting for now.
				maybe(
					ginit.Call(ginit.ScriptArgs{
						Cmd: cfg.Init.Helper,
						OnStdout: func(out string) {
							log.Log.Info("stdout", zap.String("output", out))
						},
						OnStderr: func(out string) {
							log.Log.Info("stderr", zap.String("output", out))
						}}))
				log.Log.Info("helper script ran successfully")
				log.Log.Info("PRE_INIT finished, launching INIT")
				// Perform an exec syscall which becomes PID 1
				maybe(ginit.Exec(os.Args[0], "--config=/etc/gaffer.json", "init", "--stage=INIT"))
			case INIT:
				log.Log.Info("starting onboot services")
				// Launch any containers synchronously
				// that exist in directory "onboot" in
				// the store root.
				db := store.New(*cfg, "onboot")
				maybe(db.Init())
				maybe(supervisor.Once(*cfg, db))
				maybe(db.Close())
				log.Log.Info("onboot services finished")
				// Launch all system services from the
				// "services" path in the store root.
				db = store.New(*cfg, "services")
				maybe(db.Init())
				reg := plugin.NewRegistry()
				for _, name := range cfg.Enabled {
					maybe(reg.Register(plugin.Find(name)))
				}
				maybe(reg.Configure(*cfg))
				go func() {
					maybe(reg.Run())
				}()
				maybe(ginit.Init(reg))
				maybe(db.Close())
			}
		}
	}
}
