package clustermanager

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

// RetrierClient wraps around a ClusterClient, offering retry functionality for some operations.
type RetrierClient struct {
	*clusterManagerClient
	retrier *retrier.Retrier
}

// NewRetrierClient constructs a new RetrierClient.
func NewRetrierClient(client ClusterClient, retrier *retrier.Retrier) *RetrierClient {
	return &RetrierClient{
		clusterManagerClient: newClient(client),
		retrier:              retrier,
	}
}

// ApplyKubeSpecFromBytes creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
func (c *RetrierClient) ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

// Apply creates/updates an object against the api server following a client side apply mechanism.
func (c *RetrierClient) Apply(ctx context.Context, kubeconfigPath string, obj runtime.Object) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.Apply(ctx, kubeconfigPath, obj)
		},
	)
}

// PauseCAPICluster adds a `spec.Paused: true` to the CAPI cluster resource. This will cause all
// downstream CAPI + provider controllers to skip reconciling on the paused cluster's objects.
func (c *RetrierClient) PauseCAPICluster(ctx context.Context, cluster, kubeconfig string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.PauseCAPICluster(ctx, cluster, kubeconfig)
		},
	)
}

// ResumeCAPICluster removes the `spec.Paused` on the CAPI cluster resource. This will cause all
// downstream CAPI + provider controllers to resume reconciling on the paused cluster's objects.
func (c *RetrierClient) ResumeCAPICluster(ctx context.Context, cluster, kubeconfig string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.ResumeCAPICluster(ctx, cluster, kubeconfig)
		},
	)
}

// ApplyKubeSpecFromBytesForce creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
// It forces the operation, so if api validation failed, it will delete and re-create the object.
func (c *RetrierClient) ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytesForce(ctx, cluster, data)
		},
	)
}

// ApplyKubeSpecFromBytesWithNamespace creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
// It applies all objects in the given namespace.
func (c *RetrierClient) ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace)
		},
	)
}

// UpdateAnnotationInNamespace adds/updates an annotation for the given kubernetes resource.
func (c *RetrierClient) UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.UpdateAnnotationInNamespace(ctx, resourceType, objectName, annotations, cluster, namespace)
		},
	)
}

// RemoveAnnotationInNamespace deletes an annotation for the given kubernetes resource if present.
func (c *RetrierClient) RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.RemoveAnnotationInNamespace(ctx, resourceType, objectName, key, cluster, namespace)
		},
	)
}

// ListObjects reads all Objects of a particular resource type in a namespace.
func (c *RetrierClient) ListObjects(ctx context.Context, resourceType, namespace, kubeconfig string, list kubernetes.ObjectList) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.ListObjects(ctx, resourceType, namespace, kubeconfig, list)
		},
	)
}

// DeleteGitOpsConfig deletes a GitOpsConfigObject from the cluster.
func (c *RetrierClient) DeleteGitOpsConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.DeleteGitOpsConfig(ctx, cluster, name, namespace)
		},
	)
}

// DeleteEKSACluster deletes an EKSA Cluster object from the cluster.
func (c *RetrierClient) DeleteEKSACluster(ctx context.Context, cluster *types.Cluster, name string, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.DeleteEKSACluster(ctx, cluster, name, namespace)
		},
	)
}

// DeleteAWSIamConfig deletes an AWSIamConfig object from the cluster.
func (c *RetrierClient) DeleteAWSIamConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.DeleteAWSIamConfig(ctx, cluster, name, namespace)
		},
	)
}

// DeleteOIDCConfig deletes a OIDCConfig object from the cluster.
func (c *RetrierClient) DeleteOIDCConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.DeleteOIDCConfig(ctx, cluster, name, namespace)
		},
	)
}

// DeleteCluster deletes a CAPI Cluster from the cluster.
func (c *RetrierClient) DeleteCluster(ctx context.Context, cluster, clusterToDelete *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.ClusterClient.DeleteCluster(ctx, cluster, clusterToDelete)
		},
	)
}
