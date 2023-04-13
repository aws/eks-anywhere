package clientutil

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// Kubeclient implements kubernetes.Client interface using a
// client.Client as the underlying implementation.
type KubeClient struct {
	client client.Client
}

func NewKubeClient(client client.Client) *KubeClient {
	return &KubeClient{
		client: client,
	}
}

// Get retrieves an obj for the given name and namespace from the Kubernetes Cluster.
func (c *KubeClient) Get(ctx context.Context, name, namespace string, obj kubernetes.Object) error {
	return c.client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, obj)
}

// List retrieves list of objects. On a successful call, Items field
// in the list will be populated with the result returned from the server.
func (c *KubeClient) List(ctx context.Context, list kubernetes.ObjectList) error {
	return c.client.List(ctx, list)
}

// Create saves the object obj in the Kubernetes cluster.
func (c *KubeClient) Create(ctx context.Context, obj kubernetes.Object) error {
	return c.client.Create(ctx, obj)
}

// Update updates the given obj in the Kubernetes cluster.
func (c *KubeClient) Update(ctx context.Context, obj kubernetes.Object) error {
	return c.client.Update(ctx, obj)
}

// Delete deletes the given obj from Kubernetes cluster.
func (c *KubeClient) Delete(ctx context.Context, obj kubernetes.Object) error {
	return c.client.Delete(ctx, obj)
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (c *KubeClient) DeleteAllOf(ctx context.Context, obj kubernetes.Object, opts ...kubernetes.DeleteAllOfOption) error {
	o := &kubernetes.DeleteAllOfOptions{}
	for _, opt := range opts {
		opt.ApplyToDeleteAllOf(o)
	}

	clientOptions := &client.DeleteAllOfOptions{}
	clientOptions.LabelSelector = labels.SelectorFromValidatedSet(o.HasLabels)
	clientOptions.Namespace = o.Namespace

	return c.client.DeleteAllOf(ctx, obj, clientOptions)
}
