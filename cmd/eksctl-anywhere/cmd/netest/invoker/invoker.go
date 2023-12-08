package invoker

import (
	"bytes"
	"context"
)

// Invoker executes a command.
type Invoker interface {
	Invoke(ctx context.Context, args ...string) Outcome
}

// Outcome is the result of executing a command using an Invoker.
type Outcome struct {
	Cmd    string
	Stdout bytes.Buffer
	Stderr bytes.Buffer
	Error  error
}
