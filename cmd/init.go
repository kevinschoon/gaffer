package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/ginit"
)

func recovery(err error) {
	if err != nil {
		log.Log.Error(fmt.Sprintf("unrecoverable error: %s", err.Error()))
		log.Log.Info("dropping into recovery shell!")
		maybe(ginit.Exec("/bin/sh"))
	}
}

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			single  = cmd.BoolOpt("s single", false, "drop into a shell with single user mode")
			helper  = cmd.StringOpt("helper", "", "path to a helper startup script")
			newRoot = cmd.StringArg("NEW_ROOT", "", "new root path")
		)
		cmd.Spec = "[OPTIONS] NEW_ROOT"
		cmd.Action = func() {
			if !ginit.IsRoot() {
				maybe(fmt.Errorf("init can only be run as root"))
			}
			isMem, err := ginit.IsMemFS("/")
			recovery(err)
			// Only supporting tempfs for now
			if isMem {
				recovery(ginit.Mount(ginit.TmpFS(*newRoot, 0)))
				opts, err := ginit.NewSwitchOptions(*newRoot)
				recovery(err)
				log.Log.Info(fmt.Sprintf("calling switch-root (%s)", *newRoot))
				recovery(ginit.SwitchRoot(*opts))
				log.Log.Info("switch-root completed successfully")
			} else {
				log.Log.Info("rootfs is not memory-based, will not switch-root")
			}
			if *single {
				log.Log.Info("dropping into single user mode")
				maybe(ginit.Exec("/bin/sh"))
			}
			log.Log.Info("Loading procfs")
			recovery(ginit.Mount(ginit.MountArgs{
				Source: "proc",
				Target: "/proc",
				FSType: "proc",
				Flags:  0,
				Data:   "nodev,nosuid,noexec,relatime"}))
			log.Log.Info("Loading devfs")
			recovery(ginit.Mount(
				ginit.MountArgs{
					Source: "dev",
					Target: "/dev",
					FSType: "devtmpfs",
					Data:   "nosuid,noexec,relatime,size=10m,nr_inodes=248418,mode=755",
				},
			))
			if *helper != "" {
				log.Log.Info(fmt.Sprintf("calling init helper script: %s", *helper))
				recovery(ginit.Call(ginit.ScriptArgs{Cmd: *helper}))
				log.Log.Info("helper script ran successfully")
			}
			log.Log.Info("launching new PID 1")
			recovery(ginit.Exec("/bin/gaffer", "--device=/dev/console", "launch", "--mount=true", "/containers/services"))
		}
	}
}
