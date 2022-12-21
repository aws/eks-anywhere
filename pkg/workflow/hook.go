package workflow

// HookBinder is used by hook registrars to bind tasks to be executed among the workflow's
// core task set.
type HookBinder interface {
	// BindPreWorkflowHook binds a task to a workflow that is run _before_ a workflow is executed.
	BindPreWorkflowHook(Task)

	// BindPostWorkflowHook binds a task to a workflow that is run _after_ a workflow is executed.
	BindPostWorkflowHook(Task)

	// BindPreTaskHook binds task to be run _before_ the anchor task is executed.
	BindPreTaskHook(anchor TaskName, task Task)

	// BindPostTaskHook binds Task to be run _after_ the anchor task is executed.
	BindPostTaskHook(anchor TaskName, task Task)
}
