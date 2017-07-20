/*
package log implements a rotating file logger
for Gaffer with Zap. The global logger can enable
TeeLogging where output is sent to a rotated file
such as /var/log/gaffer.log and a block device
like /dev/stderr.
*/
package log

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

const LogFileName = "gaffer"

var Log *zap.Logger

// Config is used to configure
// the global Gaffer logger.
type Config struct {
	// Device is the path to a
	// block device like /dev/stdout
	Device string
	// LogDir is a path to a directory
	// where log files will be
	// written to and rotated.
	LogDir string
	// Debug toggles debug logging.
	Debug bool
	// JSON configures the logger
	// to encode log output with JSON.
	JSON bool
	// MaxSize specifies
	// the maximum size (mb) of a
	// log before it is rotated. Since
	// Mesanine may operate only in
	// system memory this should be
	// very low by default.
	MaxSize int
	// MaxBackups is the number
	// of backups to retain after
	// rotation. This number should
	// also be very low by default
	MaxBackups int
	// Compress indicates if
	// rotated log files should be
	// compressed
	Compress bool
}

// Setup configures the global
// logger. This function should
// only be called once when
// the program is initialized.
func Setup(config Config) error {
	var (
		encoderConfig zapcore.EncoderConfig
		encoder       zapcore.Encoder
		level         zapcore.Level
	)
	if config.Debug {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		level = zapcore.DebugLevel
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		level = zapcore.InfoLevel
	}
	// Pretty colors if logging to console
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	if config.JSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}
	cores := []zapcore.Core{}
	// By default this will
	// be /dev/stderr. In Mesanine
	// this will be /dev/ttyS0.
	if config.Device != "" {
		fp, err := os.OpenFile(config.Device, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(fp), zap.NewAtomicLevelAt(level)))
	}
	// Create a directory at path
	// if it is missing. Log files
	// will be written and rotated
	// here if configured.
	if config.LogDir != "" {
		err := os.MkdirAll(config.LogDir, 0755)
		if err != nil {
			return err
		}
		sync := zapcore.AddSync(&lumberjack.Logger{
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			Filename:   filepath.Join(config.LogDir, LogFileName),
			Compress:   config.Compress,
		})
		cores = append(cores, zapcore.NewCore(encoder, sync, zap.NewAtomicLevelAt(level)))
	}
	switch len(cores) {
	case 1:
		Log = zap.New(cores[0])
	case 2:
		Log = zap.New(zapcore.NewTee(cores...))
	default:
		// Logging is completely disabled
		Log = zap.NewNop()
	}
	return nil
}
