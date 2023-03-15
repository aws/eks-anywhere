package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes the package logger. Repeat calls will overwrite the package logger which may
// result in unexpected behavior.
func Init(opts Options) error {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = nil
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}

	// Level 4 and above are used for debugging and we want a different log structure for debug
	// logs.
	if opts.Level >= 4 {
		encoderCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			// Because we use negated levels it is necessary to negate the level again so the
			// output appears in a V0 format.
			//
			// See logrAtomicLevel().
			enc.AppendString(fmt.Sprintf("V%d", -int(l)))
		}
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// zapcore.Open creates a noop logger if no paths are passed. Using a slice ensures we expand
	// the slice to nothing when opts.OutputFilePath is unset.

	var logPath []string
	if opts.OutputFilePath != "" {
		logPath = append(logPath, opts.OutputFilePath)
	}

	logFile, _, err := zap.Open(logPath...)
	if err != nil {
		return err
	}

	// Build the encoders and logger.

	fileEncoder := zapcore.NewJSONEncoder(encoderCfg)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logrAtomicLevel(opts.Level)),
		zapcore.NewCore(fileEncoder, logFile, logrAtomicLevel(MaxLogLevel)),
	)
	logger := zap.New(core)

	// Configure package state so the logger can be used by other packages.

	setLogger(zapr.NewLogger(logger))
	setOutputFilePath(opts.OutputFilePath)

	return nil
}

// Options represents a set of arguments for initializing the zap logger.
type Options struct {
	// Level is the log level at which to configure the logger from 0 to 9.
	Level int

	// OutputFilePath is an absolute file path. The file will be created if it doesn't exist.
	// All logs available at level 9 will be written to the file.
	OutputFilePath string
}

// logrAtomicLevel creates a zapcore.AtomicLevel compatible with go-logr.
func logrAtomicLevel(level int) zap.AtomicLevel {
	// The go-logr wrapper uses custom Zap log levels. To represent this in Zap, its
	// necessary to negate the level to circumvent Zap level constraints.
	//
	// See https://github.com/go-logr/zapr/blob/master/zapr.go#L50.
	return zap.NewAtomicLevelAt(zapcore.Level(-level))
}
