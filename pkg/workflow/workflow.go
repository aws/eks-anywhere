package workflow

import (
	"context"
)

// Config is the configuration for constructing a Workflow instance.
type Config struct {
	// ErrorHandler is handler called when a workflow experiences an error. The error may originate
	// from hook or from a task. The original error is alwasy returned from the workflow's Execute.
	// Optional. Defaults to a no-op handler.
	ErrorHandler ErrorHandler
}

// Workflow defines an abstract workflow that can execute a serialized set of tasks.
type Workflow struct {
	Config

	// tasks are the tasks to be run as part of the core workflow.
	tasks []namedTask

	// taskNames is a map of tasks added with AppendTask. Its used to ensure unique task names so
	// hooks aren't accidentally overwritten.
	taskNames map[TaskName]struct{}

	preWorkflowHooks  []Task
	postWorkflowHooks []Task
	preTaskHooks      map[TaskName][]Task
	postTaskHooks     map[TaskName][]Task
}

// New initializes a Workflow instance without any tasks or hooks.
func New(cfg Config) *Workflow {
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = nopErrorHandler
	}

	wflw := &Workflow{
		Config:        cfg,
		taskNames:     make(map[TaskName]struct{}),
		preTaskHooks:  make(map[TaskName][]Task),
		postTaskHooks: make(map[TaskName][]Task),
	}

	return wflw
}

// AppendTask appends t to the list of workflow tasks. Task names must be unique within a workflow.
// Duplicate names will receive an ErrDuplicateTaskName.
func (w *Workflow) AppendTask(name TaskName, t Task) error {
	if _, found := w.taskNames[name]; found {
		return ErrDuplicateTaskName{name}
	}
	w.tasks = append(w.tasks, namedTask{Task: t, Name: name})
	w.taskNames[name] = struct{}{}
	return nil
}

// Execute executes the workflow running any pre and post hooks registered for each task.
func (w *Workflow) Execute(ctx context.Context) error {
	var err error

	if ctx, err = runHooks(ctx, w.preWorkflowHooks); err != nil {
		return w.handleError(ctx, err)
	}

	for _, task := range w.tasks {
		if ctx, err = w.runPreTaskHooks(ctx, task.Name); err != nil {
			return w.handleError(ctx, err)
		}

		if ctx, err = task.RunTask(ctx); err != nil {
			return w.handleError(ctx, err)
		}

		if ctx, err = w.runPostTaskHooks(ctx, task.Name); err != nil {
			return w.handleError(ctx, err)
		}
	}

	if ctx, err = runHooks(ctx, w.postWorkflowHooks); err != nil {
		return w.handleError(ctx, err)
	}

	return nil
}

// BindPreWorkflowHook implements the HookBinder interface.
func (w *Workflow) BindPreWorkflowHook(t Task) {
	w.preWorkflowHooks = append(w.preWorkflowHooks, t)
}

// BindPostWorkflowHook implements the HookBinder interface.
func (w *Workflow) BindPostWorkflowHook(t Task) {
	w.postWorkflowHooks = append(w.postWorkflowHooks, t)
}

// BindPreTaskHook implements the HookBinder interface.
func (w *Workflow) BindPreTaskHook(id TaskName, t Task) {
	hooks := w.preTaskHooks[id]
	hooks = append(hooks, t)
	w.preTaskHooks[id] = hooks
}

// RunpreTaskHooks executes all pre hooks registered against TaskName in the order they were registered.
func (w *Workflow) runPreTaskHooks(ctx context.Context, id TaskName) (context.Context, error) {
	if hooks, ok := w.preTaskHooks[id]; ok {
		return runHooks(ctx, hooks)
	}
	return ctx, nil
}

// BindPostHook implements the HookBinder interface.
func (w *Workflow) BindPostTaskHook(id TaskName, t Task) {
	hooks := w.postTaskHooks[id]
	hooks = append(hooks, t)
	w.postTaskHooks[id] = hooks
}

// RunpreTaskHooks executes all post hooks registered against TaskName in the order they were registered.
func (w *Workflow) runPostTaskHooks(ctx context.Context, id TaskName) (context.Context, error) {
	if hooks, ok := w.postTaskHooks[id]; ok {
		return runHooks(ctx, hooks)
	}
	return ctx, nil
}

func runHooks(ctx context.Context, hooks []Task) (context.Context, error) {
	var err error
	for _, hook := range hooks {
		if ctx, err = hook.RunTask(ctx); err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (w *Workflow) handleError(ctx context.Context, err error) error {
	w.ErrorHandler(ctx, err)
	return err
}
