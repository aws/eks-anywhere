package logger

import (
	"fmt"
	"time"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitZap creates a zap logger with the provided verbosity level
// and sets it as the package logger.
// 0 is the least verbose and 10 the most verbose.
// The package logger can only be init once, so subsequent calls to this method
// won't have any effect
func InitZap(level int, opts ...LoggerOpt) error {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.Level(-1 * level))
	cfg.EncoderConfig.EncodeLevel = nil
	cfg.EncoderConfig.EncodeTime = NullTimeEncoder
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true

	// Only enabling this at level 4 because that's when
	// our debugging levels start. Ref: doc.go
	if level >= 4 {
		cfg.EncoderConfig.EncodeLevel = VLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	zapLog, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("creating zap logger: %v", err)
	}

	logr := zapr.NewLogger(zapLog)
	for _, opt := range opts {
		opt(&logr)
	}

	set(logr)
	l.V(4).Info("Logger init completed", "vlevel", level)

	return nil
}

// VLevelEncoder serializes a Level to V + v-level number,
func VLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("V%d", -1*int(l)))
}

// NullTimeEncoder skips time serialization
func NullTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
