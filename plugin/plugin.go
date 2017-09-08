package plugin

import (
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	http "github.com/mesanine/gaffer/plugin/http-server"
	reg "github.com/mesanine/gaffer/plugin/registration"
	rpc "github.com/mesanine/gaffer/plugin/rpc-server"
	"github.com/mesanine/gaffer/plugin/supervisor"
)

// Plugin implements some unit of work
// within Gaffer. Everything in Gaffer
// is a plugin.
type Plugin interface {
	// Name returns a unique name like gaffer.plugin
	Name() string
	// Config configures the underlying plugins
	Configure(config.Config) error
	// Run must launch a blocking function call.
	// The implementer can choose to handle
	// events or ignore them.
	Run(*event.EventBus) error
	// Stop stops the blocking run function.
	Stop() error
}

func Find(name string) Plugin {
	switch name {
	case "gaffer.register":
		return &reg.Server{}
	case "gaffer.rpc_server":
		return &rpc.Server{}
	case "gaffer.http_server":
		return &http.Server{}
	case "gaffer.supervisor":
		return &supervisor.Supervisor{}
	}
	panic(fmt.Sprintf("unknown plugin %s", name))
}
