package cli

import (
	"github.com/go-logr/logr"
)

const (
	markPass   = "✅ "
	markFailed = "❌ "
)

// ValidationPassed logs a success message for a validation.
func ValidationPassed(log logr.Logger, msg string) {
	log.Info(markPass + msg)
}

// ValidationFailed logs an error message for a validation.
func ValidationFailed(log logr.Logger, msg string) {
	log.Info(markFailed + msg)
}
