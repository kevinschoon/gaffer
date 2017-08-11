package plugin

import (
	"fmt"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
)

// Plugin implements some unit of work
// within Gaffer. All functions of
// Gaffer exist as a plugin interface.
type Plugin interface {
	// Name returns a unique name like gaffer.plugin
	Name() string
	// Run must launch a blocking function call.
	// The implementer can choose to handle
	// events or ignore them.
	Run(*event.EventBus) error
	// Stop stops the blocking run function.
	Stop() error
}

// Registry stores a collection of
// plugins each with a unique name.
type Registry map[string]Plugin

// Registry registers a Plugin within
// the registry.
func (r Registry) Register(p Plugin) error {
	if _, ok := r[p.Name()]; ok {
		return fmt.Errorf("plugin with name %s is already registered", p.Name())
	}
	r[p.Name()] = p
	return nil
}

// NewRegistry builds a new plugin registry.
func NewRegistry() Registry {
	return Registry{}
}

// Run runs registry of plugins each
// in a separate Go routine. It waits until
// all plugins have returned. If any plugin
// returns an error the function returns
// immediately.
func Run(registry Registry) error {
	eb := event.NewEventBus()
	eb.Start()
	defer eb.Stop()
	errCh := make(chan error, len(registry))
	// Launch a routine that listens for a SHUTDOWN
	// event and calls Stop() on each plugin. This
	// avoids each plugin having to implement the
	// same logic.
	go func(errCh chan error, eb *event.EventBus) {
		sub := event.NewSubscriber()
		eb.Subscribe(sub)
		evtCh := sub.Chan()
		for evt := range evtCh {
			if evt.Type() == event.SHUTDOWN {
				for name, plugin := range registry {
					log.Log.Warn(fmt.Sprintf("shutting down plugin %s", name))
					errCh <- plugin.Stop()
				}
			}
		}
	}(errCh, eb)
	// Launch each plugin in the registry
	for name, plugin := range registry {
		log.Log.Info(fmt.Sprintf("launching plugin %s", name))
		go func(plugin Plugin) {
			errCh <- plugin.Run(eb)
		}(plugin)
	}
	// Wait until we recieve the same number
	// of errors or nil as there are registered
	// plugins
	for i := 0; i < len(registry); i++ {
		if err := <-errCh; err != nil {
			log.Log.Error(err.Error())
			// Give up immediately when we encounter
			// a plugin error
			return err
		}
	}
	// All plugins successfully shutdown
	log.Log.Info("all plugins have shutdown")
	return nil
}
