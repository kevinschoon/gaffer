package logger

import (
	"context"
	"github.com/jawher/mow.cli"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/util"
	"google.golang.org/grpc"
)

type Logger struct {
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

func (l Logger) Configure(cfg config.Config) error { return nil }

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

func (l Logger) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	return &ReadResponse{}, nil
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
			cmd.Spec = "[OPTIONS]"
			var req *ReadRequest
			cmd.Before = func() {
				req = &ReadRequest{}
			}
			cmd.Action = func() {
				resp, err := client.Read(context.Background(), req, cfg.CallOpts()...)
				util.Maybe(err)
				util.JSONToStdout(resp)
			}
		})
	}
}
