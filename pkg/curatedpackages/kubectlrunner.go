package curatedpackages

import (
	"bytes"
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

type KubectlRunner interface {
	ExecuteCommand(ctx context.Context, opts ...string) (bytes.Buffer, error)
	ExecuteFromYaml(ctx context.Context, yaml []byte, opts ...string) (bytes.Buffer, error)
	// GetObject performs a GET call to the kube API server authenticating with a kubeconfig file
	// and unmarshalls the response into the provdied Object
	// If the object is not found, it returns an error implementing apimachinery errors.APIStatus
	GetObject(ctx context.Context, resourceType, name, namespece, kubeconfig string, obj runtime.Object) error
	// HasResource is true if the resource can be retrieved from the API and has length > 0.
	HasResource(ctx context.Context, resourceType string, name string, kubeconfig string, namespace string) (bool, error)
}
