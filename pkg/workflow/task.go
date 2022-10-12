package workflow

import "context"

// TaskName uniquely identifies a task within a given workflow.
type TaskName string

// Task represents an individual step within a workflow that can be run.
type Task interface {
	// RunTask executes the task. Tasks may return a context that should be used in subsequent task
	// execution.
	RunTask(context.Context) (context.Context, error)
}

// namedTask associates a name with a Task in the context of a Workflow to enable hook lookup.
type namedTask struct {
	Task
	Name TaskName
}
