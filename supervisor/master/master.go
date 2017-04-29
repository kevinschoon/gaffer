package master

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
	fn := func() error {
		// Request cluster information
		c, err := opts.Client.Cluster(opts.Cluster)
		if err != nil {
			return err
		}
		// Create new Master configuration
		master, err := config.NewMaster(c)
		if err != nil {
			return err
		}
		// Mesos master was disconnected
		if master.Running {
			opts.Logger.Info(
				"mesos master",
				zap.String("msg", "re-joining cluster"),
			)
		}
		// Detect the cluster state
		state := c.State()
		opts.Logger.Info(
			"mesos master",
			zap.String("state", state.String()),
		)
		// Update cluster with this master configuration
		err = opts.Client.Update(c)
		if err != nil {
			return err
		}
		// Masters are still converging
		if state < config.MASTER_CONVERGED {
			return fmt.Errorf("Masters still converging")
		}
		// Create new master process
		proc, err := master.Process(opts.Logger)
		if err != nil {
			return err
		}
		// Start the process
		err = proc.Start()
		if err != nil {
			return err
		}
		// Record that master is now running
		master.Running = true
		// Update the remote cluster configuration
		err = opts.Client.Update(c)
		if err != nil {
			return err
		}
		// Check if the process is running every 2s
		for {
			time.Sleep(2000 * time.Millisecond)
			// kill -n 0 <PID>
			err := proc.Signal(syscall.Signal(0))
			if err != nil {
				// Master is no longer running
				master.Running = false
				if err := opts.Client.Update(c); err != nil {
					// No longer can update the server, process is dead
					opts.Logger.Error(
						"mesos master",
						zap.Error(err),
					)
				}
				// Process is dead
				return err
			}
			// Refresh the cluster configuraiton prior to update
			c, err = opts.Client.Cluster(opts.Cluster)
			if err != nil {
				// Could not refresh cluster configuration
				opts.Logger.Error(
					"mesos master",
					zap.Error(err),
				)
				continue
			}
			// Process is still running but server unable to tell server
			if err := opts.Client.Update(c); err != nil {
				opts.Logger.Error(
					"mesos master",
					zap.Error(err),
				)
			}
		}
	}
	notify := func(err error, d time.Duration) {
		opts.Logger.Info(
			"mesos master",
			zap.String("msg", "mesos master process has died"),
			zap.Duration("duration", d),
			zap.Error(err),
		)
	}
	return backoff.RetryNotify(fn, backoff.NewExponentialBackOff(), notify)
}
