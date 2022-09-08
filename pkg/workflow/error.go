package workflow

import "context"

// ErrorHandler is a function called when a workflow experiences an error during execution. The
// error may originate from hook execution or from a task.
type ErrorHandler func(context.Context, error)

func nopErrorHandler(context.Context, error) {}
