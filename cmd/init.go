package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/ginit"
)

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			newRoot = cmd.StringArg("NEW_ROOT", "", "new root path")
			newInit = cmd.StringArg("NEW_INIT", "", "new init pid 1")
		)
		cmd.Spec = "[OPTIONS] NEW_ROOT NEW_INIT"
		cmd.Action = func() {
			if !ginit.IsRoot() {
				maybe(fmt.Errorf("init can only be run as root"))
			}
			isMem, err := ginit.IsMemFS("/")
			maybe(err)
			if !isMem {
				maybe(fmt.Errorf("current root is not ramfs or tempfs, refusing to switch_root"))
			}
			// Only supporting tempfs for now
			maybe(ginit.TmpFS(*newRoot, 0).Call())
			opts, err := ginit.NewSwitchOptions(*newRoot)
			maybe(err)
			maybe(ginit.SwitchRoot(*opts))
			maybe(ginit.Exec(*newInit))
		}
	}
}
