package plugin

import (
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	rpc "github.com/mesanine/gaffer/plugin/rpc-server"
	"github.com/mesanine/gaffer/plugin/supervisor"
	"os"
	"os/signal"
)

type shutdown struct {
	Name string
	Err  error
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

// Registered checks to see if a plugin
// has been registered.
func (r Registry) Registered(id string) bool {
	_, ok := r[id]
	return ok
}

// Configure configures all of the underlying plugins.
func (registry Registry) Configure(cfg config.Config) error {
	for name, plugin := range registry {
		log.Log.Info(fmt.Sprintf("configuring plugin %s", name))
		if err := plugin.Configure(cfg); err != nil {
			return err
		}
	}
	// If the RPC server and Supervisor are running
	// let the server call runc commands directly.
	if registry.Registered("gaffer.rpc-server") && registry.Registered("gaffer.supervisor") {
		registry["gaffer.rpc-server"].(*rpc.Server).SetRuncFn(
			registry["gaffer.supervisor"].(*supervisor.Supervisor).Runc,
		)
	}

	return nil
}

// Run runs a registry of plugins each
// in a separate Go routine. It waits until
// all plugins have returned. If any plugin
// returns an error the function returns
// immediately.
func (registry Registry) Run() error {
	eb := event.NewEventBus()
	eb.Start()
	defer eb.Stop()
	shutdownCh := make(chan shutdown)
	// Launch a routine that listens for a SHUTDOWN
	// event and calls Stop() on each plugin. This
	// avoids each plugin having to implement the
	// same logic.
	go func(shutdownCh chan shutdown, eb *event.EventBus) {
		sub := event.NewSubscriber()
		eb.Subscribe(sub)
		defer eb.Unsubscribe(sub)
		for {
			if evt := sub.Next(); evt != nil {
				switch {
				case event.Is(event.REQUEST_SHUTDOWN)(*evt):
					for name, plugin := range registry {
						log.Log.Warn(fmt.Sprintf("shutting down plugin: %s", name))
						if err := plugin.Stop(); err != nil {
							log.Log.Error(fmt.Sprintf("encountered error shutting down plugin %s: %s", name, err.Error()))
						}
					}
					return
				}
			}
		}
	}(shutdownCh, eb)
	sigCh := make(chan os.Signal, 1)
	// TODO: Which other signals might we encounter as init?
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	go func(eb *event.EventBus) {
		sig := <-sigCh
		log.Log.Warn(fmt.Sprintf("caught signal %s", sig.String()))
		// Signal we are shutting down
		eb.Push(event.New(event.REQUEST_SHUTDOWN))
	}(eb)
	// Launch each plugin in the registry
	for name, plugin := range registry {
		log.Log.Info(fmt.Sprintf("launching plugin %s", name))
		go func(plugin Plugin) {
			shutdownCh <- shutdown{
				Name: name,
				Err:  plugin.Run(eb),
			}
		}(plugin)
	}
	// Wait until we recieve the same number
	// of errors or nil as there are registered
	// plugins
	for i := 0; i < len(registry); i++ {
		msg := <-shutdownCh
		if msg.Err != nil {
			log.Log.Error(fmt.Sprintf("plugin %s encountered an error: %s", msg.Name, msg.Err.Error()))
			// Give up immediately when we encounter
			// a plugin error
			return msg.Err
		}
		log.Log.Info(fmt.Sprintf("plugin %s has shutdown", msg.Name))
	}
	// All plugins successfully shutdown
	log.Log.Info("all plugins have shutdown")
	return nil
}
