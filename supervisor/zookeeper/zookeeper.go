package zookeeper

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
	// Main ZK launch function
	fn := func() error {
		// Request cluster information
		c, err := opts.Client.Cluster(opts.Cluster)
		if err != nil {
			return err
		}
		// Create new Zookeeper configuration
		zk, err := config.NewZookeeper(c)
		if err != nil {
			return err
		}
		// Zookeeper was disconnected
		if zk.Running {
			opts.Logger.Info(
				"zookeeper",
				zap.String("msg", "re-joining cluster"),
			)
		}
		// Detect the cluster state
		state := c.State()
		opts.Logger.Info(
			"zookeeper",
			zap.String("state", state.String()),
		)
		// Update cluster with this zookeeper configuration
		err = opts.Client.Update(c)
		if err != nil {
			return err
		}
		// Zookeepers are still converging
		if state < config.ZK_CONVERGED {
			return fmt.Errorf("Zookeeper still converging")
		}
		// Create new Zookeeper process
		proc, err := zk.Process(opts.Logger)
		if err != nil {
			return err
		}
		// Start the process
		err = proc.Start()
		if err != nil {
			return err
		}
		// Record that Zookeeper is now running
		zk.Running = true
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
				// Zookeeper is no longer running
				zk.Running = false
				if err := opts.Client.Update(c); err != nil {
					// No longer can update the server, process is dead
					opts.Logger.Error(
						"zookeeper",
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
					"zookeeper",
					zap.Error(err),
				)
				continue
			}
			if err := opts.Client.Update(c); err != nil {
				opts.Logger.Error(
					"zookeeper",
					zap.Error(err),
				)
			}
		}
	}
	notify := func(err error, d time.Duration) {
		opts.Logger.Info(
			"zookeeper",
			zap.String("msg", "zookeeper process has died"),
			zap.Duration("duration", d),
			zap.Error(err),
		)
	}
	return backoff.RetryNotify(fn, backoff.NewExponentialBackOff(), notify)
}
