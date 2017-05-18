package log

import (
	"go.uber.org/zap"
)

var (
	Log    *zap.Logger
	Level  zap.AtomicLevel
	config zap.Config
)

// Json toggles JSON logging output
func Json() {
	config.Encoding = "json"
	Log, _ = config.Build()
}

// Debug toggles development mode
func Debug() {
	config = zap.NewDevelopmentConfig()
	config.Level = Level
	Log, _ = config.Build()
}

func init() {
	config = zap.NewProductionConfig()
	// Default to human friendly console output
	config.Encoding = "console"
	Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.Level = Level
	Log, _ = config.Build()
}
