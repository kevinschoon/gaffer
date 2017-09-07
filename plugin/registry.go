package plugin

import (
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	rpc "github.com/mesanine/gaffer/plugin/rpc-server"
	"github.com/mesanine/gaffer/plugin/supervisor"
	"github.com/mesanine/ginit"
	"os"
)

type shutdown struct {
	Name string
	Err  error
}

// Registry stores a collection of
// plugins each with a unique name.
type Registry struct {
	eb      *event.EventBus
	plugins map[string]Plugin
}

func NewRegistry() *Registry {
	return &Registry{
		eb:      event.NewEventBus(),
		plugins: map[string]Plugin{},
	}
}

// Registry registers a Plugin within
// the registry.
func (r Registry) Register(p Plugin) error {
	if _, ok := r.plugins[p.Name()]; ok {
		return fmt.Errorf("plugin with name %s is already registered", p.Name())
	}
	r.plugins[p.Name()] = p
	return nil
}

// Registered checks to see if a plugin
// has been registered.
func (r Registry) Registered(id string) bool {
	_, ok := r.plugins[id]
	return ok
}

// Configure configures all of the underlying plugins.
func (r Registry) Configure(cfg config.Config) error {
	for name, plugin := range r.plugins {
		log.Log.Info(fmt.Sprintf("configuring plugin %s", name))
		if err := plugin.Configure(cfg); err != nil {
			return err
		}
	}
	// If the RPC server and Supervisor are running
	// let the server call runc commands directly.
	if r.Registered("gaffer.rpc-server") && r.Registered("gaffer.supervisor") {
		r.plugins["gaffer.rpc-server"].(*rpc.Server).SetRuncFn(
			r.plugins["gaffer.supervisor"].(*supervisor.Supervisor).Runc,
		)
	}

	return nil
}

// Handle implements the ginit.Handler interface.
func (r Registry) Handle(sig os.Signal) error {
	if ginit.Terminal(sig) {
		for name, plugin := range r.plugins {
			log.Log.Info(fmt.Sprintf("shutting down plugin %s", name))
			err := plugin.Stop()
			if err != nil {
				return err
			}
			log.Log.Info(fmt.Sprintf("shutdown plugin %s", name))
		}
	}
	return nil
}

// Run runs a registry of plugins each
// in a separate Go routine. It waits until
// all plugins have returned. If any plugin
// returns an error the function returns
// immediately.
func (r Registry) Run() error {
	shutdownCh := make(chan shutdown)
	// Launch each plugin in the registry
	for name, plugin := range r.plugins {
		log.Log.Info(fmt.Sprintf("launching plugin %s", name))
		go func(plugin Plugin) {
			err := plugin.Run(r.eb)
			shutdownCh <- shutdown{
				Name: plugin.Name(),
				Err:  err,
			}
		}(plugin)
	}
	// Wait until we recieve the same number
	// of errors or nil as there are registered
	// plugins
	for i := 0; i < len(r.plugins); i++ {
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
