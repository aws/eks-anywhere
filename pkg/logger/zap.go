package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapOpts represents a set of arguments for initializing the zap logger.
type ZapOpts struct {
	Level          int      // indicates the log level of the logger.
	OutputFilePath string   // if specified, the logger will output to file at this path.
	WithNames      []string // specified name elements are added to the logger's name.
}

// InitZap creates a zap logger with the provided verbosity level
// and sets it as the package logger.
// 0 is the least verbose and 10 the most verbose.
// The package logger can only be init once, so subsequent calls to this method
// won't have any effect.
func InitZap(args ZapOpts) error {
	logr, err := newZap(args)
	if err != nil {
		return err
	}
	set(logr, args.OutputFilePath)
	l.V(4).Info("Logger init completed", "vlevel", args.Level)

	return nil
}

// VLevelEncoder serializes a Level to V + v-level number,.
func VLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("V%d", -1*int(l)))
}

// NullTimeEncoder skips time serialization.
func NullTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}

func newZap(args ZapOpts) (logr.Logger, error) {
	outputPaths := []string{}
	if args.OutputFilePath != "" {
		outputPaths = append(outputPaths, args.OutputFilePath)
	}

	cfg := config{
		level:         args.Level,
		encoderConfig: zap.NewDevelopmentEncoderConfig(),
		outputPaths:   outputPaths,
	}

	cfg.encoderConfig.EncodeLevel = nil
	cfg.encoderConfig.EncodeTime = NullTimeEncoder

	// Only enabling this at level 4 because that's when
	// our debugging levels start. Ref: doc.go
	if args.Level >= 4 {
		cfg.encoderConfig.EncodeLevel = VLevelEncoder
		cfg.encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	zapLog, err := build(cfg)
	if err != nil {
		return logr.Discard(), fmt.Errorf("creating zap logger: %v", err)
	}

	logr := zapr.NewLogger(zapLog)

	for _, name := range args.WithNames {
		logr = logr.WithName(name)
	}

	return logr, err
}

// newAtomicLevelAt returns an appropriate zap.AtomicLevel given an integer representing the log level.
func newAtomicLevelAt(level int) zap.AtomicLevel {
	return zap.NewAtomicLevelAt(zapcore.Level(-1 * level))
}

// build constructs a logger and returns a logger.
func build(cfg config) (*zap.Logger, error) {
	sink, err := cfg.openSinks()
	if err != nil {
		return nil, err
	}

	logger := zap.New(cfg.buildCore(sink))
	return logger, nil
}

func (cfg config) openSinks() (zapcore.WriteSyncer, error) {
	sink, _, err := zap.Open(cfg.outputPaths...)
	return sink, err
}

// config helps to construct a customized zap logger.
type config struct {
	outputPaths   []string
	level         int
	encoderConfig zapcore.EncoderConfig
}

func (cfg config) buildCore(sink zapcore.WriteSyncer) zapcore.Core {
	fileEncoder := zapcore.NewJSONEncoder(cfg.encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(cfg.encoderConfig)

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), newAtomicLevelAt(cfg.level)),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(sink), newAtomicLevelAt(9)),
	)
}
