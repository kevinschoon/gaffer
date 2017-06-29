package runc

import (
	"bufio"
	"fmt"
	"github.com/containerd/go-runc"
	"github.com/vektorlab/gaffer/log"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// TODO: Catch write errors
// TODO: Close log files
// TODO: Propigate errors back
type IO struct {
	id   string
	path string
	rio  runc.IO
	tee  bool
}

func (i *IO) Start() error {
	// stdout
	stdoutP := filepath.Join(i.path, fmt.Sprintf("%s_stdout", i.id))
	log.Log.Info(fmt.Sprintf("recording output for %s to %s", i.id, stdoutP))
	if err := Tee("stdout", i.id, stdoutP, i.rio.Stdout()); err != nil {
		return err
	}
	// stderr
	stderrP := filepath.Join(i.path, fmt.Sprintf("%s_stderr", i.id))
	log.Log.Info(fmt.Sprintf("recording output for %s to %s", i.id, stderrP))
	if err := Tee("stderr", i.id, stderrP, i.rio.Stderr()); err != nil {
		return err
	}

	return nil
}

func (i *IO) Close() error {
	return i.rio.Close()
}

func NewIO(id, path string) (*IO, error) {
	rio, err := runc.NewPipeIO(0, 0)
	if err != nil {
		return nil, err
	}
	return &IO{
		id:   id,
		path: path,
		rio:  rio,
		tee:  true,
	}, nil
}

func Tee(name, service, path string, rc io.ReadCloser) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := ioutil.WriteFile(path, []byte{}, 0666)
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	go func(fd *os.File) {
		defer fd.Close()
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			text := scanner.Text()
			log.Log.Info(
				name,
				zap.String(service, text),
			)
			// TODO
			fd.WriteString(text)
		}
	}(file)
	return nil
}
