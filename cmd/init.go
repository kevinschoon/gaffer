package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/util"
	"github.com/mesanine/ginit"
	"go.uber.org/zap"
	"os"
)

func initCMD(cfg *config.Config) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		recovery := cmd.Bool(cli.BoolOpt{
			Name:   "r recovery",
			Desc:   "launch a recovery shell",
			Value:  false,
			EnvVar: "GAFFER_RECOVERY_SHELL",
		})
		cmd.Spec = "[OPTIONS]"
		cmd.Action = func() {
			// TODO: Need a clean way to initialize the OS
			// in "stages" but I am not sure what the best
			// abstraction is to do this at the moment.
			if !ginit.IsRoot() {
				util.Maybe(fmt.Errorf("init can only be run as root"))
			}
			// Check if we are running on a memory-based
			// file system. If we are we need to switch-root
			// and re-mount as tempfs.
			isMem, err := ginit.IsMemFS("/")
			util.Maybe(err)
			if isMem {
				// Only supporting tempfs for now
				util.Maybe(ginit.Mount(ginit.TmpFS(cfg.Init.NewRoot, 0)))
				opts, err := ginit.NewSwitchOptions(cfg.Init.NewRoot)
				util.Maybe(err)
				log.Log.Info(fmt.Sprintf("calling switch-root (%s)", cfg.Init.NewRoot))
				util.Maybe(ginit.SwitchRoot(*opts))
				log.Log.Info("switch-root completed successfully")
			} else {
				log.Log.Info("rootfs is not memory-based, will not switch-root")
			}
			if *recovery {
				log.Log.Info("dropping into single user mode")
				util.Maybe(ginit.Exec("/bin/sh"))
			}
			// Load the procfs file system
			log.Log.Info("Loading procfs")
			util.Maybe(ginit.Mount(ginit.MountArgs{
				Source: "proc",
				Target: "/proc",
				FSType: "proc",
				Flags:  0,
				Data:   "nodev,nosuid,noexec,relatime"}))
			// Load devfs filesystem
			log.Log.Info("Loading devfs")
			util.Maybe(ginit.Mount(
				ginit.MountArgs{
					Source: "dev",
					Target: "/dev",
					FSType: "devtmpfs",
					Data:   "nosuid,noexec,relatime,size=10m,nr_inodes=248418,mode=755",
				}))
			log.Log.Info(fmt.Sprintf("calling init helper script: %s", cfg.Init.Helper))
			// Call the init helper script which does most of the
			// heavy lifting for now.
			util.Maybe(
				ginit.Call(ginit.ScriptArgs{
					Cmd: cfg.Init.Helper,
					OnStdout: func(out string) {
						log.Log.Info("stdout", zap.String("output", out))
					},
					OnStderr: func(out string) {
						log.Log.Info("stderr", zap.String("output", out))
					}}))
			log.Log.Info("helper script ran successfully")
			// Perform an exec syscall which becomes PID 1
			util.Maybe(ginit.Exec(os.Args[0], "--config=/etc/gaffer.json", "run"))
		}
	}
}
