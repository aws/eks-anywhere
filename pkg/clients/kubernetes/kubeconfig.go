package kubernetes

import (
	"context"
)

// KubeconfigClient is an authenticated kubernetes API client
// it authenticates using the credentials of a kubeconfig file.
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
// and unmarshalls the response into the provided Object.
func (c *KubeconfigClient) Get(ctx context.Context, name, namespace string, obj Object) error {
	return c.client.Get(ctx, name, namespace, c.kubeconfig, obj)
}

// List retrieves list of objects. On a successful call, Items field
// in the list will be populated with the result returned from the server.
func (c *KubeconfigClient) List(ctx context.Context, list ObjectList) error {
	return c.client.List(ctx, c.kubeconfig, list)
}

// Create saves the object obj in the Kubernetes cluster.
func (c *KubeconfigClient) Create(ctx context.Context, obj Object) error {
	return c.client.Create(ctx, c.kubeconfig, obj)
}

// Update updates the given obj in the Kubernetes cluster.
func (c *KubeconfigClient) Update(ctx context.Context, obj Object) error {
	return c.client.Update(ctx, c.kubeconfig, obj)
}

// ApplyServerSide creates or patches and object using server side logic.
func (c *KubeconfigClient) ApplyServerSide(ctx context.Context, fieldManager string, obj Object, opts ...ApplyServerSideOption) error {
	return c.client.ApplyServerSide(ctx, c.kubeconfig, fieldManager, obj, opts...)
}

// Delete deletes the given obj from Kubernetes cluster.
func (c *KubeconfigClient) Delete(ctx context.Context, obj Object) error {
	return c.client.Delete(ctx, c.kubeconfig, obj)
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (c *KubeconfigClient) DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error {
	return c.client.DeleteAllOf(ctx, c.kubeconfig, obj, opts...)
}
