package log

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func Debug() {
	Log, _ = zap.NewDevelopment()
}

func init() {
	Log, _ = zap.NewProduction()
}
