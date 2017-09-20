package supervisor

import (
	"bufio"
	"github.com/containerd/go-runc"
	"github.com/mesanine/gaffer/log"
	"go.uber.org/zap"
	"io"
	"os"
)

type IO struct {
	id    string
	rio   runc.IO
	debug bool
}

func (i *IO) Start() {
	logFn := func(stream string, rc io.ReadCloser) {
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
			switch stream {
			case "stdout":
				log.Log.Debug(i.id, zap.String("stdout", text))
			case "stderr":
				log.Log.Debug(i.id, zap.String("stderr", text))
			}
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
	rio, err := runc.NewPipeIO(os.Getuid(), os.Getgid())
	if err != nil {
		return nil, err
	}
	return &IO{
		id:  id,
		rio: rio,
	}, nil
}
