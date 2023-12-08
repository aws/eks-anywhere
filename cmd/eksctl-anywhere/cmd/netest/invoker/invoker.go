package invoker

import (
	"bytes"
	"context"
)

type Invoker interface {
	Invoke(ctx context.Context, args ...string) Outcome
}

type Outcome struct {
	Cmd    string
	Stdout bytes.Buffer
	Stderr bytes.Buffer
	Error  error
}
