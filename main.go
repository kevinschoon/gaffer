package main

import (
	"flag"
	"fmt"
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
	default:
		flag.PrintDefaults()
		failOnErr(fmt.Errorf("Invalid server mode %s", *mode))
	}
	/*
			p := process{
				cmd:  exec.Command("/home/kevin/repos/go/src/github.com/vektorlab/gaffer/script.sh"),
				env:  map[string]string{"COOL": "BEANS"},
				err:  make(chan error),
				quit: make(chan struct{}, 1),
				log:  logger,
			}
			if err := p.Start(); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		loop:
			for {
				select {
				case err := <-p.err:
					fmt.Println("Uh oh", err.Error())
					break loop
				case <-p.quit:
					fmt.Println("quittin")
					break loop
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
	*/
}
