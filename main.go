package main

import (
	"flag"
	"fmt"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/server"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/supervisor/agent"
	"github.com/vektorlab/gaffer/supervisor/master"
	"github.com/vektorlab/gaffer/supervisor/zookeeper"
	"go.uber.org/zap"
	"os"
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
		mode      = flag.String("mode", "", "server mode [server, supervisor]")
		process   = flag.String("process", "", "process to supervise [zookeeper, master, agent]")
		anonymous = flag.Bool("anonymous", false, "allow anonymous access")
		dbStr     = flag.String("db", "./gaffer.db", "database connection string")
		endpoint  = flag.String("endpoint", "http://127.0.0.1:8080", "gaffer HTTP endpoint")
		token     = flag.String("token", "", "gaffer HTTP API Token")
		cluster   = flag.String("cluster", "", "cluster ID")
	)
	flag.Parse()
	switch *mode {
	case "server":
		db, err := store.NewSQLStore(*dbStr, logger)
		failOnErr(err)
		defer db.Close()
		server := server.NewServer(db, logger)
		server.Anonymous = *anonymous
		failOnErr(server.Serve())
	case "supervisor":
		client := client.NewClient(*endpoint, *token, logger)
		switch *process {
		case "zookeeper":
			failOnErr(
				zookeeper.Run(
					zookeeper.Opts{
						Client:  client,
						Logger:  logger,
						Cluster: *cluster,
					},
				))
		case "master":
			failOnErr(
				master.Run(
					master.Opts{
						Client:  client,
						Logger:  logger,
						Cluster: *cluster,
					},
				))
		case "agent":
			failOnErr(
				agent.Run(
					agent.Opts{
						Client:  client,
						Logger:  logger,
						Cluster: *cluster,
					},
				))
		}
	default:
		flag.PrintDefaults()
		failOnErr(fmt.Errorf("Invalid server mode %s", *mode))
	}
}
