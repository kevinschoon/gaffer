/*
package log implements a rotating file logger
for Gaffer with Zap. The global logger can enable
TeeLogging where output is sent to a rotated file
such as /var/log/gaffer.log and a block device
like /dev/stderr.
*/
package log

import (
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

const LogFileName = "gaffer.log"

var Log *zap.Logger

// Setup configures the global
// logger. This function should
// only be called once when
// the program is initialized.
func Setup(config config.Config) error {
	var (
		encoderConfig zapcore.EncoderConfig
		encoder       zapcore.Encoder
		level         zapcore.Level
	)
	if config.Logger.Debug {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		level = zapcore.DebugLevel
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		level = zapcore.InfoLevel
	}
	if config.Logger.JSON {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		// Pretty colors if logging to console
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}
	cores := []zapcore.Core{}
	// By default this will
	// be /dev/stderr. In Mesanine
	// this will be /dev/ttyS0.
	if config.Logger.Device != "" {
		fp, err := os.OpenFile(config.Logger.Device, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(fp), zap.NewAtomicLevelAt(level)))
	}
	// Create a directory at path
	// if it is missing. Log files
	// will be written and rotated
	// here if configured.
	if config.Logger.LogDir != "" {
		err := os.MkdirAll(config.Logger.LogDir, 0755)
		fmt.Println("RIGHT HERE: ", config.Logger.LogDir, err)
		if err != nil {
			return err
		}
		sync := zapcore.AddSync(&lumberjack.Logger{
			MaxSize:    config.Logger.MaxSize,
			MaxBackups: config.Logger.MaxBackups,
			Filename:   filepath.Join(config.Logger.LogDir, LogFileName),
			Compress:   config.Logger.Compress,
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

func init() {
	// Noop logging by default (useful is importing as library for testing)
	Log = zap.NewNop()
}
