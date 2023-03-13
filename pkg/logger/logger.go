package logger

import (
	"os"
	"sync"

	"github.com/go-logr/logr"
)

// MaxLogLevel denotes the maximum log level supported by the logger package.
const MaxLogLevel = 9

const (
	markPass    = "✅ "
	markSuccess = "🎉 "
	markFailed  = "❌ "
	markWarning = "⚠️"
)

var (
	pkgLogger    logr.Logger = logr.Discard()
	pkgLoggerMtx sync.RWMutex
)

func setLogger(l logr.Logger) {
	pkgLoggerMtx.Lock()
	defer pkgLoggerMtx.Unlock()
	pkgLogger = l
}

// Get returns the logger instance that has been previously set.
// If no logger has been set, it returns a null logger.
func Get() logr.Logger {
	pkgLoggerMtx.RLock()
	defer pkgLoggerMtx.RUnlock()
	return pkgLogger
}

// MaxLogging determines if the package logger is configured to log at MaxLogLevel.
func MaxLogging() bool {
	return Get().V(MaxLogLevel).Enabled()
}

// Fatal is equivalent to Get().Error() followed by a call to os.Exit(1).
func Fatal(err error, msg string) {
	Get().Error(err, msg)
	os.Exit(1)
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line. The key/value pairs can then be used to add additional
// variable information. The key/value pairs should alternate string
// keys and arbitrary values.
func Info(msg string, keysAndValues ...interface{}) {
	Get().Info(msg, keysAndValues...)
}

// V returns an Logger value for a specific verbosity level, relative to
// this Logger. In other words, V values are additive.  V higher verbosity
// level means a log message is less important.  It's illegal to pass a log
// level less than zero.
func V(level int) logr.Logger {
	return Get().V(level)
}

// Error logs an error message using the package logger.
func Error(err error, msg string, keysAndValues ...interface{}) {
	Get().Error(err, msg, keysAndValues...)
}

// MarkPass logs a message prefixed with a green check emoji.
func MarkPass(msg string, keysAndValues ...interface{}) {
	Get().V(0).Info(markPass+msg, keysAndValues...)
}

// MarkSuccess logs a message prefixed with a popper emoji.
func MarkSuccess(msg string, keysAndValues ...interface{}) {
	Get().V(0).Info(markSuccess+msg, keysAndValues...)
}

// MarkFail logs a message prefixed with a cross emoji.
func MarkFail(msg string, keysAndValues ...interface{}) {
	Get().V(0).Info(markFailed+msg, keysAndValues...)
}

// MarkWarning logs a message prefixed with a warning mark.
func MarkWarning(msg string, keysAndValues ...interface{}) {
	Get().V(0).Info(markWarning+msg, keysAndValues...)
}
