package util

import (
	"encoding/json"
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os"
)

func Maybe(err error) {
	if err != nil {
		if log.Log != nil {
			log.Log.Error("gaffer encountered an un-recoverable error", zap.Error(err))
		}
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func JSONToStdout(o interface{}) {
	Maybe(json.NewEncoder(os.Stdout).Encode(o))
}

func NewClientConn(cfg config.Config) (*grpc.ClientConn, error) {
	log.Log.Info(fmt.Sprintf("dailing %s", cfg.Address))
	opts, err := cfg.DailOpts()
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(cfg.Address, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
