package plugin

import (
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"google.golang.org/grpc"
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

// RPC returns a grpc.ServiceDesc
// that will be used to expose
// methods via the global Gaffer
// RPC server.
type RPC interface {
	RPC() *grpc.ServiceDesc
}

// CLI returns a CmdInitializer that
// can be used to expose functionality
// to the Gaffer CLI.
type CLI interface {
	CLI(*config.Config) cli.CmdInitializer
}
