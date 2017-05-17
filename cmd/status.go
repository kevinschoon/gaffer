package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/store"
)

func statusCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			db, err := store.NewSQLStore("gaffer", "./gaffer.db", false)
			maybe(err)
			c := client.NewClient(db)
			hosts, err := c.Hosts()
			maybe(err)
			for _, host := range hosts {
				fmt.Println(host)
			}
			services, err := c.Services()
			maybe(err)
			for _, svc := range services {
				fmt.Println(svc)
			}
			processes, err := c.Processes()
			maybe(err)
			fmt.Println(processes)
		}
	}
}
