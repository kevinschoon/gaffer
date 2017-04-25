package main

import (
	"flag"
	"fmt"
	"github.com/cenkalti/backoff"
	"go.uber.org/zap"
	"os"
	"syscall"
	"time"
)

func failOnErr(err error) {
	if err != nil {
		fmt.Println("error: ", err.Error())
		os.Exit(1)
	}
}

func main() {
	logger, _ := zap.NewDevelopment()
	var (
		mode      = flag.String("mode", "", "server mode [server, master, agent, zookeeper]")
		anonymous = flag.Bool("anonymous", false, "allow anonymous access")
		dbStr     = flag.String("db", "./gaffer.db", "database connection string")
		endpoint  = flag.String("endpoint", "http://127.0.0.1:8080", "gaffer HTTP endpoint")
		token     = flag.String("token", "", "gaffer HTTP API Token")
		cluster   = flag.String("cluster", "", "cluster ID")
	)
	flag.Parse()
	// TODO DRY
	switch *mode {
	case "server":
		db, err := NewSQLStore(*dbStr, logger)
		failOnErr(err)
		defer db.Close()
		server := NewServer(db, logger)
		server.Anonymous = *anonymous
		failOnErr(server.Serve())
	case "master":
		client := NewClient(*endpoint, *token, logger)
		// Main Mesos launch function
		fn := func() error {
			// Request cluster information
			c, err := client.Cluster(*cluster)
			if err != nil {
				return err
			}
			// Create new Master configuration
			master, err := NewMaster(c)
			if err != nil {
				return err
			}
			// Mesos master was disconnected
			if master.Running {
				logger.Info(
					"mesos master",
					zap.String("msg", "re-joining cluster"),
				)
			}
			// Detect the cluster state
			state := c.State()
			logger.Info(
				"mesos master",
				zap.String("state", state.String()),
			)
			// Update cluster with this master configuration
			err = client.Update(c)
			if err != nil {
				return err
			}
			// Masters are still converging
			if state < MASTER_CONVERGED {
				return fmt.Errorf("Masters still converging")
			}
			// Create new master process
			proc, err := master.Process(logger)
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
			err = client.Update(c)
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
					if err := client.Update(c); err != nil {
						// No longer can update the server, process is dead
						logger.Error(
							"mesos master",
							zap.Error(err),
						)
					}
					// Process is dead
					return err
				}
				// Refresh the cluster configuraiton prior to update
				c, err = client.Cluster(*cluster)
				if err != nil {
					// Could not refresh cluster configuration
					logger.Error(
						"mesos master",
						zap.Error(err),
					)
					continue
				}
				// Process is still running but server unable to tell server
				if err := client.Update(c); err != nil {
					logger.Error(
						"mesos master",
						zap.Error(err),
					)
				}
			}
		}
		notify := func(err error, d time.Duration) {
			logger.Info(
				"mesos master",
				zap.String("msg", "mesos master process has died"),
				zap.Duration("duration", d),
				zap.Error(err),
			)
		}
		failOnErr(backoff.RetryNotify(fn, backoff.NewExponentialBackOff(), notify))

	case "agent":
		client := NewClient(*endpoint, *token, logger)
		// Main Mesos launch function
		fn := func() error {
			// Request cluster information
			c, err := client.Cluster(*cluster)
			if err != nil {
				return err
			}
			// Create new Agent configuration
			agent := NewAgent(c)
			// Detect the cluster state
			state := c.State()
			logger.Info(
				"agent",
				zap.String("state", state.String()),
			)
			// Masters are still converging
			if state < MASTER_READY {
				return fmt.Errorf("Masters not ready")
			}
			// Create new agent process
			proc, err := agent.Process(logger)
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
				c, err = client.Cluster(*cluster)
				if err != nil {
					// Could not refresh cluster configuration
					logger.Error(
						"agent",
						zap.Error(err),
					)
					continue
				}
				// Process is still running but server unable to tell server
				if err := client.Update(c); err != nil {
					logger.Error(
						"agent",
						zap.Error(err),
					)
				}
			}
		}
		notify := func(err error, d time.Duration) {
			logger.Info(
				"agent",
				zap.String("msg", "mesos agent process has died"),
				zap.Duration("duration", d),
				zap.Error(err),
			)
		}
		failOnErr(backoff.RetryNotify(fn, backoff.NewExponentialBackOff(), notify))
	case "zookeeper":
		client := NewClient(*endpoint, *token, logger)
		// Main ZK launch function
		fn := func() error {
			// Request cluster information
			c, err := client.Cluster(*cluster)
			if err != nil {
				return err
			}
			// Create new Zookeeper configuration
			zk, err := NewZookeeper(c)
			if err != nil {
				return err
			}
			// Zookeeper was disconnected
			if zk.Running {
				logger.Info(
					"zookeeper",
					zap.String("msg", "re-joining cluster"),
				)
			}
			// Detect the cluster state
			state := c.State()
			logger.Info(
				"zookeeper",
				zap.String("state", state.String()),
			)
			// Update cluster with this zookeeper configuration
			err = client.Update(c)
			if err != nil {
				return err
			}
			// Zookeepers are still converging
			if state < ZK_CONVERGED {
				return fmt.Errorf("Zookeeper still converging")
			}
			// Create new Zookeeper process
			proc, err := zk.Process(logger)
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
			err = client.Update(c)
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
					if err := client.Update(c); err != nil {
						// No longer can update the server, process is dead
						logger.Error(
							"zookeeper",
							zap.Error(err),
						)
					}
					// Process is dead
					return err
				}
				// Refresh the cluster configuraiton prior to update
				c, err = client.Cluster(*cluster)
				if err != nil {
					// Could not refresh cluster configuration
					logger.Error(
						"zookeeper",
						zap.Error(err),
					)
					continue
				}
				if err := client.Update(c); err != nil {
					logger.Error(
						"zookeeper",
						zap.Error(err),
					)
				}
			}
		}
		notify := func(err error, d time.Duration) {
			logger.Info(
				"zookeeper",
				zap.String("msg", "zookeeper process has died"),
				zap.Duration("duration", d),
				zap.Error(err),
			)
		}
		failOnErr(backoff.RetryNotify(fn, backoff.NewExponentialBackOff(), notify))
	default:
		flag.PrintDefaults()
		failOnErr(fmt.Errorf("Invalid server mode %s", *mode))
	}
}
