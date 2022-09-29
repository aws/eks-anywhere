package curatedpackages

import (
	"bytes"
	"context"
)

type KubectlRunner interface {
	ExecuteCommand(ctx context.Context, opts ...string) (bytes.Buffer, error)
	ExecuteFromYaml(ctx context.Context, yaml []byte, opts ...string) (bytes.Buffer, error)
	// HasResource is true if the resource can be retrieved from the API and has length > 0.
	HasResource(ctx context.Context, resourceType string, name string, kubeconfig string, namespace string) (bool, error)
}
