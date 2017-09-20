package logger

import (
	"bufio"
	"context"
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"os"
	"time"
)

// Logger is an RPC service for
// reading and writing to the
// configured log device.
type Logger struct {
	path string
	err  chan error
	stop chan bool
}

func New() *Logger {
	return &Logger{
		err:  make(chan error, 1),
		stop: make(chan bool, 1),
	}
}

func (l Logger) Name() string { return "logger" }

func (l *Logger) Configure(cfg config.Config) error {
	l.path = fmt.Sprintf("%s/%s", cfg.Logger.LogDir, "gaffer.log")
	return nil
}

func (l Logger) Run(*event.EventBus) error {
	select {
	case err := <-l.err:
		return err
	case <-l.stop:
		return nil
	}
}

func (l Logger) Stop() error {
	l.stop <- true
	return nil
}

func (l Logger) RPC() *grpc.ServiceDesc { return &_RPC_serviceDesc }

// Write triggers a logging event on this server causing
// it to be recorded in the log file if configured or
// just to be displayed to stdout.
func (l Logger) Write(stream RPC_WriteServer) error {
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Log.Info(string(data.Content), zap.Int64("offset", data.Offset))
	}
}

// Read reads log data line by line from the Gaffer
// configured log directory if it exists.
func (l Logger) Read(req *ReadRequest, stream RPC_ReadServer) error {
	fd, err := os.Open(l.path)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.Seek(req.Offset, 0)
	if err != nil {
		return err
	}
	var offset int
	reader := bufio.NewReader(fd)
	for i := 0; i < int(req.Lines) || int(req.Lines) == 0; i++ {
		raw, err := reader.ReadBytes('\n')
		// BUG: If the log file rotates
		// while following the remote client
		// won't pickup the new changes until
		// they re-establish the RPC connection.
		if err == io.EOF && req.Follow {
			// avoid busy loop
			time.Sleep(50 * time.Millisecond)
			continue
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		offset += len(raw)
		err = stream.Send(&LogData{
			Content: raw,
			Offset:  int64(offset),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (l Logger) CLI(cfg *config.Config) cli.CmdInitializer {
	return func(cmd *cli.Cmd) {
		var client RPCClient
		cmd.Before = func() {
			conn, err := util.NewClientConn(*cfg)
			util.Maybe(err)
			client = NewRPCClient(conn)
		}
		cmd.Command("read", "Read from the server log", func(cmd *cli.Cmd) {
			var (
				follow = cmd.BoolOpt("f follow", false, "follow log output")
				lines  = cmd.IntOpt("n lines", 0, "number of lines to read")
			)
			cmd.Spec = "[OPTIONS]"
			var req *ReadRequest
			cmd.Before = func() {
				req = &ReadRequest{
					Follow: *follow,
					Lines:  int64(*lines),
				}
			}
			cmd.Action = func() {
				stream, err := client.Read(context.Background(), req, cfg.CallOpts()...)
				util.Maybe(err)
				for {
					data, err := stream.Recv()
					if err == io.EOF {
						break
					}
					util.Maybe(err)
					//util.JSONToStdout(data)
					fmt.Fprintln(os.Stdout, string(data.Content))
				}
			}
		})
		cmd.Command("write", "Write to the remote server log", func(cmd *cli.Cmd) {
			cmd.Action = func() {
				stream, err := client.Write(context.Background(), cfg.CallOpts()...)
				util.Maybe(err)
				scanner := bufio.NewScanner(os.Stdin)
				var offset int
				for scanner.Scan() {
					offset += len(scanner.Bytes())
					err = stream.Send(&LogData{Content: scanner.Bytes(), Offset: int64(offset)})
					util.Maybe(err)
				}
			}
		})
	}
}
