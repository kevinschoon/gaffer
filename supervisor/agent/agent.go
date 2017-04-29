package agent

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/config"
	"go.uber.org/zap"
	"syscall"
	"time"
)

type Opts struct {
	Client  *client.Client
	Logger  *zap.Logger
	Cluster string
}

func Run(opts Opts) error {
	// Main Mesos launch function
	fn := func() error {
		// Request cluster information
		c, err := opts.Client.Cluster(opts.Cluster)
		if err != nil {
			return err
		}
		// Create new Agent configuration
		agent := config.NewAgent(c)
		// Detect the cluster state
		state := c.State()
		opts.Logger.Info(
			"agent",
			zap.String("state", state.String()),
		)
		// Masters are still converging
		if state < config.MASTER_READY {
			return fmt.Errorf("Masters not ready")
		}
		// Create new agent process
		proc, err := agent.Process(opts.Logger)
		if err != nil {
			return err
		}
		// Start the process
		err = proc.Start()
		if err != nil {
			return err
		}
		// Check if the process is running every 2s
		for {
			time.Sleep(2000 * time.Millisecond)
			// kill -n 0 <PID>
			err := proc.Signal(syscall.Signal(0))
			if err != nil {
				// Agent is no longer running
				return err
			}
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
		}
	}
	notify := func(err error, d time.Duration) {
		opts.Logger.Info(
			"agent",
			zap.String("msg", "mesos agent process has died"),
			zap.Duration("duration", d),
			zap.Error(err),
		)
	}
	return backoff.RetryNotify(fn, backoff.NewExponentialBackOff(), notify)
}
