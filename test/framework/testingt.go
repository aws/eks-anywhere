package framework

import "testing"

// T defines test support functionality, ala the Go stdlib testing.T.
//
// Being able to change its implementation supports logging functionality for
// test support methods that are executed outside of a test, such as when
// bringing up or tearing down test clusters that will be used and re-used
// throughout multiple tests.
//
// Only those methods currently in use are defined. Add more methods from
// stdlib testing.T as necessary.
type T interface {
	Cleanup(func())
	Error(...any)
	Errorf(string, ...any)
	Fail()
	FailNow()
	Failed() bool
	Fatal(...any)
	Fatalf(string, ...any)
	Helper()
	Log(args ...any)
	Logf(format string, args ...any)
	Name() string
	Parallel()
	Run(string, func(*testing.T)) bool
	Setenv(string, string)
	Skip(...any)
	SkipNow()
	Skipf(string, ...any)
	Skipped() bool
	TempDir() string
}

// T ensures that *testing.T implements T, to detect API drift.
var _ T = (*testing.T)(nil)

// LoggingOnlyT implements select logging and error handling functionality of T.
//
// Most non-logging, non-error reporting methods will simply panic.
type LoggingOnlyT struct{}

// NewLoggingOnlyT creates a LoggingOnlyT, which does what its name implies.
func NewLoggingOnlyT() *LoggingOnlyT { return &LoggingOnlyT{} }

// Cleanup implements T.
func (t LoggingOnlyT) Cleanup(_ func()) {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Error implements T.
func (t LoggingOnlyT) Error(_ ...any) {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Errorf implements T.
func (t LoggingOnlyT) Errorf(_ string, _ ...any) {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Fail implements T.
func (t LoggingOnlyT) Fail() {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// FailNow implements T.
func (t LoggingOnlyT) FailNow() {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Failed implements T.
func (t LoggingOnlyT) Failed() bool {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Fatal implements T.
func (t LoggingOnlyT) Fatal(_ ...any) {
	v := &testing.T{}
	v.Fatal("foo")
	// panic("LoggingOnlyT implements only the logging methods of T")
}

// Fatalf implements T.
func (t LoggingOnlyT) Fatalf(format string, args ...any) {
	v := &testing.T{}
	v.Fatalf(format, args...)
}

// Helper implements T.
func (t LoggingOnlyT) Helper() {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Log implements T.
func (t LoggingOnlyT) Log(args ...any) {
	(&testing.T{}).Log(args...)
}

// Logf implements T.
func (t LoggingOnlyT) Logf(format string, args ...any) {
	(&testing.T{}).Logf(format, args...)
}

// Name implements T.
func (t LoggingOnlyT) Name() string {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Parallel implements T.
func (t LoggingOnlyT) Parallel() {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Run implements T.
func (t LoggingOnlyT) Run(_ string, _ func(*testing.T)) bool {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Setenv implements T.
func (t LoggingOnlyT) Setenv(_ string, _ string) {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Skip implements T.
func (t LoggingOnlyT) Skip(_ ...any) {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// SkipNow implements T.
func (t LoggingOnlyT) SkipNow() {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Skipf implements T.
func (t LoggingOnlyT) Skipf(_ string, _ ...any) {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// Skipped implements T.
func (t LoggingOnlyT) Skipped() bool {
	panic("LoggingOnlyT implements only the logging methods of T")
}

// TempDir implements T.
func (t LoggingOnlyT) TempDir() string {
	panic("LoggingOnlyT implements only the logging methods of T")
}
