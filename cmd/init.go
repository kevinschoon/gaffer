package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/plugin"
	http "github.com/mesanine/gaffer/plugin/http-server"
	regSrv "github.com/mesanine/gaffer/plugin/registration"
	rpc "github.com/mesanine/gaffer/plugin/rpc-server"
	"github.com/mesanine/gaffer/plugin/supervisor"
	"os"
	"strings"
)

func initCMD() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			path       = cmd.StringArg("PATH", "/containers", "container init path")
			root       = cmd.StringOpt("root", "/run/runc", "runc root path")
			once       = cmd.BoolOpt("once", false, "run the services only once, synchronously")
			httpPort   = cmd.IntOpt("http-port", 9090, "http server port")
			rpcPort    = cmd.IntOpt("rpc-port", 10000, "rpc server port")
			etcdSrvs   = cmd.StringOpt("etcd", "http://localhost:2379", "list of etcd endpoints seperated by ,")
			mount      = cmd.BoolOpt("mount", false, "handle overlay mounts")
			configPath = cmd.StringOpt("config-path", "/var/mesanine", "service configuration path")
		)
		cmd.Spec = "[OPTIONS] [PATH]"
		cmd.Action = func() {
			cfg := config.Config{
				Store: config.Store{
					BasePath:   *path,
					ConfigPath: *configPath,
				},
				Runc: config.Runc{
					Root:  *root,
					Mount: *mount,
				},
				RegistrationServer: config.RegistrationServer{
					EtcdEndpoints: strings.Split(*etcdSrvs, ","),
				},
				RPCServer: config.RPCServer{
					Port: *rpcPort,
				},
				HTTPServer: config.HTTPServer{
					Port: *httpPort,
				},
			}
			if *once {
				maybe(supervisor.Once(cfg))
				os.Exit(0)
			}
			reg := plugin.Registry{}
			maybe(reg.Register(&rpc.Server{}))
			maybe(reg.Register(&http.Server{}))
			maybe(reg.Register(&regSrv.Server{}))
			maybe(reg.Register(&supervisor.Supervisor{}))
			maybe(reg.Configure(cfg))
			maybe(reg.Run())
		}
	}
}
