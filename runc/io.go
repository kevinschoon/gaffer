package runc

import (
	"bufio"
	"fmt"
	"github.com/containerd/go-runc"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"io"
)

type IO struct {
	id    string
	rio   runc.IO
	debug bool
}

func (i *IO) Start() {
	logFn := func(stream string, rc io.ReadCloser) {
		log.Log.Debug(fmt.Sprintf("Monitoring output from service %s", i.id))
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			text := scanner.Text()
			if err := scanner.Err(); err != nil {
				log.Log.Info(
					i.id,
					zap.Error(err),
				)
				break
			}
			log.Log.Info(
				i.id,
				zap.String(stream, text),
			)
		}
	}
	// stdout
	go logFn("stdout", i.rio.Stdout())
	// stderr
	go logFn("stderr", i.rio.Stderr())
}

func (i *IO) Close() error {
	return i.rio.Close()
}

func NewIO(id string) (*IO, error) {
	rio, err := runc.NewPipeIO(0, 0)
	if err != nil {
		return nil, err
	}
	return &IO{
		id:  id,
		rio: rio,
	}, nil
}
