package clients

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// Kubeclient implements kubernetes.Client interface using a
// client.Client as the underlying implementation
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
