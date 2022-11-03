package test

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// KubeClient implements kubernetes.Client by using client.Client.
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

// NewFakeKubeClient returns a KubeClient that uses a fake client.Client under the hood.
func NewFakeKubeClient(objs ...client.Object) *KubeClient {
	return NewKubeClient(fake.NewClientBuilder().WithObjects(objs...).Build())
}

// NewFakeKubeClientAlwaysError returns a KubeClient that will always fail in any operation
// This is achieved by injecting an empty Scheme, which will make the underlying client.Client
// incapable of determining the resource type for a particular client.Object.
func NewFakeKubeClientAlwaysError(objs ...client.Object) *KubeClient {
	return NewKubeClient(
		fake.NewClientBuilder().WithScheme(runtime.NewScheme()).WithObjects(objs...).Build(),
	)
}
