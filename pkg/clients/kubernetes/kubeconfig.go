package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

// KubeconfigClient is an authenticated kubernetes API client
// it authenticates using the credentials of a kubeconfig file
type KubeconfigClient struct {
	client     *UnAuthClient
	kubeconfig string
}

func NewKubeconfigClient(client *UnAuthClient, kubeconfig string) *KubeconfigClient {
	return &KubeconfigClient{
		client:     client,
		kubeconfig: kubeconfig,
	}
}

// Get performs a GET call to the kube API server
// and unmarshalls the response into the provdied Object
func (c *KubeconfigClient) Get(ctx context.Context, name, namespace string, obj runtime.Object) error {
	return c.client.Get(ctx, name, namespace, c.kubeconfig, obj)
}
