package workflow

import (
	"context"
	"fmt"
)

// ErrorHandler is a function called when a workflow experiences an error during execution. The
// error may originate from hook execution or from a task.
type ErrorHandler func(context.Context, error)

func nopErrorHandler(context.Context, error) {}

// ErrDuplicateTaskName indicates 2 tasks with the same TaskName have been added to a workflow.
type ErrDuplicateTaskName struct {
	Name TaskName
}

func (e ErrDuplicateTaskName) Error() string {
	return fmt.Sprintf("duplicate task name: %v", e.Name)
}
