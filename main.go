package main

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"time"
)

func main() {
	logger, _ := zap.NewDevelopment()
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
}
