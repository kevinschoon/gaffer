package cmd

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/supervisor"
	"github.com/vektorlab/gaffer/user"
	"strings"
	"sync"
)

func superviseCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			service  = cmd.StringOpt("s service", "", "service names to supervise, e.g. svc1,svc2")
			endpoint = cmd.StringOpt("e endpoint", "http://localhost:9090", "gaffer API server")
			auth     = cmd.StringOpt("u user", "", "user:pass basic auth string")
		)
		cmd.Action = func() {
			var usr *user.User
			if *auth != "" {
				u, err := user.FromString(*auth)
				maybe(err)
				usr = u
			}
			if *service == "" {
				maybe(fmt.Errorf("must specify service name"))
			}
			var wg sync.WaitGroup
			services := make(chan string)
			go func() {
				for name := range services {
					wg.Add(1)
					go func() {
						maybe(supervisor.Run(
							supervisor.Opts{
								Store:   store.NewHTTPStore(*endpoint, usr),
								Service: name,
							},
						))
					}()
				}
			}()
			for _, svc := range strings.Split(*service, ",") {
				services <- svc
			}
			close(services)
			wg.Wait()
		}
	}
}
