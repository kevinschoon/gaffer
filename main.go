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
		c, err := client.Cluster(*cluster)
		failOnErr(err)
		failOnErr(client.UntilZKReady(c))
	case "agent":
		client := NewClient(*endpoint, *token, logger)
		c, err := client.Cluster(*cluster)
		failOnErr(err)
		failOnErr(client.UntilMasterReady(c))
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
			// Create new Zookeeper process
			proc := zk.Process(logger)
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
				// Process is still running but server unable to tell server
				if err := client.Update(c); err != nil {
					logger.Error(
						"zookeeper",
						zap.Error(err),
					)
				}
				time.Sleep(2000 * time.Millisecond)
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
