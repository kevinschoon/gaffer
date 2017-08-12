package plugin

import (
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
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
