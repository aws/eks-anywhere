package clustermanager

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

// RetrierClient wraps around a ClusterClient, offering retry functionality for some operations.
type RetrierClient struct {
	*client
	*retrier.Retrier
}

// NewRetrierClient constructs a new RetrierClient.
func NewRetrierClient(client ClusterClient, retrier *retrier.Retrier) *RetrierClient {
	return &RetrierClient{
		client:  NewClient(client),
		Retrier: retrier,
	}
}

// ApplyKubeSpecFromBytes creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
func (c *RetrierClient) ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

// Apply creates/updates an object against the api server following a client side apply mechanism.
func (c *RetrierClient) Apply(ctx context.Context, kubeconfigPath string, obj runtime.Object) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.Apply(ctx, kubeconfigPath, obj)
		},
	)
}

// ApplyKubeSpecFromBytesForce creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
// It forces the operation, so if api validation failed, it will delete and re-create the object.
func (c *RetrierClient) ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytesForce(ctx, cluster, data)
		},
	)
}

// ApplyKubeSpecFromBytesWithNamespace creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
// It applies all objects in the given namespace.
func (c *RetrierClient) ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace)
		},
	)
}

// UpdateAnnotationInNamespace adds/updates an annotation for the given kubernetes resource.
func (c *RetrierClient) UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.UpdateAnnotationInNamespace(ctx, resourceType, objectName, annotations, cluster, namespace)
		},
	)
}

// RemoveAnnotationInNamespace deletes an annotation for the given kubernetes resource if present.
func (c *RetrierClient) RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.RemoveAnnotationInNamespace(ctx, resourceType, objectName, key, cluster, namespace)
		},
	)
}
