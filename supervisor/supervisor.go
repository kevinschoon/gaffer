package supervisor

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"time"
)

const PollTime = 2000 * time.Millisecond

func maybeLog(fn func() error) error {
	err := fn()
	if err != nil {
		log.Log.Warn("supervisor", zap.Error(err))
	}
	return err
}

type Opts struct {
	Client    *client.Client
	ClusterID string
	Service   string
}

func Run(opts Opts) error {
	// Main Mesos launch function
	fn := func() error {
		// Request cluster information
		config, err := opts.Client.Cluster(opts.ClusterID)
		if err != nil {
			return err
		}
		var self *cluster.Host
		for _, host := range config.Hosts {
			if err := host.Register(); err != nil {
				self = host
			}
		}
		if self == nil {
			return fmt.Errorf("Could not register self with cluster")
		}
		service, ok := self.Services[opts.Service]
		if !ok {
			return fmt.Errorf("Invalid service %s", opts.Service)
		}
		// Check if the process is running every 2s
		maybeLog(func() error {
			return service.Start()
		})

		for {
			time.Sleep(PollTime)

			if !service.Running() {
				log.Log.Warn("supervisor", zap.String("message", fmt.Sprintf("service %s is not running", opts.Service)))
				err := maybeLog(func() error { return service.Start() })
				if err != nil {
					continue
				}
			}
			err := opts.Client.Update(config)
			if err != nil {
				log.Log.Info("supervisor", zap.String("message", "failed to update remote server"), zap.Error(err))
			}

			/*
				// Refresh the cluster configuraiton prior to update
				c, err = opts.Client.Cluster(opts.Cluster)
				if err != nil {
					// Could not refresh cluster configuration
					opts.Logger.Error(
						"agent",
						zap.Error(err),
					)
					continue
				}
				// Process is still running but server unable to tell server
				if err := opts.Client.Update(c); err != nil {
					opts.Logger.Error(
						"agent",
						zap.Error(err),
					)
				}
			*/
		}
	}
	notify := func(err error, d time.Duration) {
		log.Log.Info(
			"supervisor",
			zap.String("message", fmt.Sprintf("service %s has failed", opts.Service)),
			zap.Duration("duration", d),
			zap.Error(err),
		)
	}
	return backoff.RetryNotify(fn, backoff.NewExponentialBackOff(), notify)
}
