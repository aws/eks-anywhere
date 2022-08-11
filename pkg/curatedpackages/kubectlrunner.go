package curatedpackages

import (
	"bytes"
	"context"
)

type KubectlRunner interface {
	ExecuteCommand(ctx context.Context, opts ...string) (bytes.Buffer, error)
	ExecuteFromYaml(ctx context.Context, yaml []byte, opts ...string) (bytes.Buffer, error)
	GetResource(ctx context.Context, resourceType string, name string, kubeconfig string, namespace string) (bool, error)
}
