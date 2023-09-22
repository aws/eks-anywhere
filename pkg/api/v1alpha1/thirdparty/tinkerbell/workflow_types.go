// Package tinkerbell represents APIs and types copied from the tinkerbell/tink repo.
// +kubebuilder:object:generate=true
package tinkerbell

// Workflow, Task, and Action are copied from https://pkg.go.dev/github.com/tinkerbell/tink@v0.8.0/workflow#pkg-types.
// json tags have been added.

// Workflow represents a workflow to be executed.
type Workflow struct {
	Version       string `json:"version"`
	Name          string `json:"name"`
	ID            string `json:"id"`
	GlobalTimeout int    `json:"global_timeout"`
	Tasks         []Task `json:"tasks"`
}

// Task represents a task to be executed as part of a workflow.
type Task struct {
	Name        string            `json:"name"`
	WorkerAddr  string            `json:"worker"`
	Actions     []Action          `json:"actions"`
	Volumes     []string          `json:"volumes,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// Action is the basic executional unit for a workflow.
type Action struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Timeout     int64             `json:"timeout"`
	Command     []string          `json:"command,omitempty"`
	OnTimeout   []string          `json:"on-timeout,omitempty"`
	OnFailure   []string          `json:"on-failure,omitempty"`
	Volumes     []string          `json:"volumes,omitempty"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Pid         string            `json:"pid,omitempty"`
}
