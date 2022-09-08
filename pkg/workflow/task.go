package workflow

import "context"

// TaskName is a string that uniquely identifies a task.
type TaskName string

// Task represents an individual step within a workflow that can be run.
type Task interface {
	// GeGetNametID returns a unique identifier for the task.
	GetName() TaskName

	// RunTask executes the task. Tasks may return a context that should be used in subsequent task
	// execution.
	RunTask(context.Context) (context.Context, error)
}
