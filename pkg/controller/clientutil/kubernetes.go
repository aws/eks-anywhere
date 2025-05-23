package clientutil

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
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
func (c *KubeClient) List(ctx context.Context, list kubernetes.ObjectList, opts ...kubernetes.ListOption) error {
	o := &kubernetes.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(o)
	}

	clientOptions := &client.ListOptions{}
	clientOptions.Namespace = o.Namespace

	return c.client.List(ctx, list, clientOptions)
}

// Create saves the object obj in the Kubernetes cluster.
func (c *KubeClient) Create(ctx context.Context, obj kubernetes.Object) error {
	return c.client.Create(ctx, obj)
}

// Update updates the given obj in the Kubernetes cluster.
func (c *KubeClient) Update(ctx context.Context, obj kubernetes.Object) error {
	return c.client.Update(ctx, obj)
}

// Patch updates the given obj in the Kubernetes cluster.
// This method is used only in unit tests currently.
func (c *KubeClient) Patch(ctx context.Context, obj kubernetes.Object, patch kubernetes.Patch, opts ...kubernetes.PatchOption) error {
	o := &kubernetes.PatchOptions{}
	for _, opt := range opts {
		opt.ApplyToPatch(o)
	}
	patchOpts := &client.PatchOptions{}
	if o.Force {
		patchOpts.Force = ptr.Bool(true)
	}
	return c.client.Patch(ctx, obj, patch, patchOpts)
}

// ApplyServerSide creates or patches and object using server side logic.
func (c *KubeClient) ApplyServerSide(ctx context.Context, fieldManager string, obj kubernetes.Object, opts ...kubernetes.ApplyServerSideOption) error {
	o := &kubernetes.ApplyServerSideOptions{}
	for _, opt := range opts {
		opt.ApplyToApplyServerSide(o)
	}

	patchOpts := &client.PatchOptions{
		FieldManager: fieldManager,
	}
	if o.ForceOwnership {
		patchOpts.Force = ptr.Bool(true)
	}
	return c.client.Patch(ctx, obj, client.Apply, patchOpts)
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
