package curatedpackages

import (
	"bytes"
	"context"
)

type KubectlRunner interface {
	ExecuteCommand(ctx context.Context, opts ...string) (bytes.Buffer, error)
	ExecuteCommandFromBytes(ctx context.Context, data []byte, opts ...string) (bytes.Buffer, error)
}
