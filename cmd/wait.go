package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/log"
	"os"
	"time"
)

const WaitInterval = 100 * time.Millisecond

func waitCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			path    = cmd.StringArg("PATH", "", "path to wait for")
			timeout = cmd.StringOpt("t timeout", "1m", "timeout")
		)
		cmd.Action = func() {
			max, err := time.ParseDuration(*timeout)
			maybe(err)
			started := time.Now()
			for {
				if time.Since(started) >= max {
					maybe(fmt.Errorf("timeout exceeded: %s", *timeout))
				}
				_, err := os.Stat(*path)
				if os.IsNotExist(err) {
					log.Log.Info(fmt.Sprintf("waiting for %s", *path))
					time.Sleep(WaitInterval)
					continue
				}
				maybe(err)
				break
			}
		}
	}
}
