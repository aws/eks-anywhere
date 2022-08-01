package logger

import (
	"os"
	"sync"

	"github.com/go-logr/logr"
)

const (
	maxLogging  = 9
	markPass    = "✅ "
	markSuccess = "🎉 "
	markFailed  = "❌ "
)

var (
	l    logr.Logger = logr.Discard()
	once sync.Once
)

func set(logger logr.Logger) {
	once.Do(func() {
		l = logger
	})
}

// Get returns the logger instance that has been previously set.
// If no logger has been set, it returns a null logger.
func Get() logr.Logger {
	return l
}

func MaxLogging() bool {
	return l.V(maxLogging).Enabled()
}

func MaxLoggingLevel() int {
	return maxLogging
}

// Fatal is equivalent to Get().Error() followed by a call to os.Exit(1).
func Fatal(err error, msg string) {
	l.Error(err, msg)
	os.Exit(1)
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line. The key/value pairs can then be used to add additional
// variable information. The key/value pairs should alternate string
// keys and arbitrary values.
func Info(msg string, keysAndValues ...interface{}) {
	l.Info(msg, keysAndValues...)
}

// V returns an Logger value for a specific verbosity level, relative to
// this Logger. In other words, V values are additive.  V higher verbosity
// level means a log message is less important.  It's illegal to pass a log
// level less than zero.
func V(level int) logr.Logger {
	return l.V(level)
}

func Error(err error, msg string, keysAndValues ...interface{}) {
	l.Error(err, msg, keysAndValues...)
}

func MarkPass(msg string, keysAndValues ...interface{}) {
	l.V(0).Info(markPass+msg, keysAndValues...)
}

func MarkSuccess(msg string, keysAndValues ...interface{}) {
	l.V(0).Info(markSuccess+msg, keysAndValues...)
}

func MarkFail(msg string, keysAndValues ...interface{}) {
	l.V(0).Info(markFailed+msg, keysAndValues...)
}

type LoggerOpt func(logr *logr.Logger)

func WithName(name string) LoggerOpt {
	return func(logr *logr.Logger) {
		*logr = (*logr).WithName(name)
	}
}
