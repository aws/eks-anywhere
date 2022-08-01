package kubernetes

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	Get(ctx context.Context, name, namespace string, obj Object) error
}

type Object client.Object

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
// and unmarshalls the response into the provided Object
func (c *KubeconfigClient) Get(ctx context.Context, name, namespace string, obj Object) error {
	return c.client.Get(ctx, name, namespace, c.kubeconfig, obj)
}
