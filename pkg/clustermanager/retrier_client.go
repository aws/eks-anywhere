package clustermanager

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

// KubernetesRetrierClient wraps around a KubernetesClient, offering retry functionality for some operations.
type KubernetesRetrierClient struct {
	KubernetesClient
	retrier *retrier.Retrier
}

// NewRetrierClient constructs a new RetrierClient.
func NewRetrierClient(client KubernetesClient, retrier *retrier.Retrier) *KubernetesRetrierClient {
	return &KubernetesRetrierClient{
		KubernetesClient: client,
		retrier:          retrier,
	}
}

// ApplyKubeSpecFromBytes creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
func (c *KubernetesRetrierClient) ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

// Apply creates/updates an object against the api server following a client side apply mechanism.
func (c *KubernetesRetrierClient) Apply(ctx context.Context, kubeconfigPath string, obj runtime.Object, opts ...kubernetes.KubectlApplyOption) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.Apply(ctx, kubeconfigPath, obj, opts...)
		},
	)
}

// PauseCAPICluster adds a `spec.Paused: true` to the CAPI cluster resource. This will cause all
// downstream CAPI + provider controllers to skip reconciling on the paused cluster's objects.
func (c *KubernetesRetrierClient) PauseCAPICluster(ctx context.Context, cluster, kubeconfig string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.PauseCAPICluster(ctx, cluster, kubeconfig)
		},
	)
}

// ResumeCAPICluster removes the `spec.Paused` on the CAPI cluster resource. This will cause all
// downstream CAPI + provider controllers to resume reconciling on the paused cluster's objects.
func (c *KubernetesRetrierClient) ResumeCAPICluster(ctx context.Context, cluster, kubeconfig string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.ResumeCAPICluster(ctx, cluster, kubeconfig)
		},
	)
}

// ApplyKubeSpecFromBytesForce creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
// It forces the operation, so if api validation failed, it will delete and re-create the object.
func (c *KubernetesRetrierClient) ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.ApplyKubeSpecFromBytesForce(ctx, cluster, data)
		},
	)
}

// ApplyKubeSpecFromBytesWithNamespace creates/updates the objects defined in a yaml manifest against the api server following a client side apply mechanism.
// It applies all objects in the given namespace.
func (c *KubernetesRetrierClient) ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace)
		},
	)
}

// UpdateAnnotationInNamespace adds/updates an annotation for the given kubernetes resource.
func (c *KubernetesRetrierClient) UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.UpdateAnnotationInNamespace(ctx, resourceType, objectName, annotations, cluster, namespace)
		},
	)
}

// RemoveAnnotationInNamespace deletes an annotation for the given kubernetes resource if present.
func (c *KubernetesRetrierClient) RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.RemoveAnnotationInNamespace(ctx, resourceType, objectName, key, cluster, namespace)
		},
	)
}

// ListObjects reads all Objects of a particular resource type in a namespace.
func (c *KubernetesRetrierClient) ListObjects(ctx context.Context, resourceType, namespace, kubeconfig string, list kubernetes.ObjectList) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.ListObjects(ctx, resourceType, namespace, kubeconfig, list)
		},
	)
}

// DeleteGitOpsConfig deletes a GitOpsConfigObject from the cluster.
func (c *KubernetesRetrierClient) DeleteGitOpsConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.DeleteGitOpsConfig(ctx, cluster, name, namespace)
		},
	)
}

// DeleteEKSACluster deletes an EKSA Cluster object from the cluster.
func (c *KubernetesRetrierClient) DeleteEKSACluster(ctx context.Context, cluster *types.Cluster, name string, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.DeleteEKSACluster(ctx, cluster, name, namespace)
		},
	)
}

// DeleteAWSIamConfig deletes an AWSIamConfig object from the cluster.
func (c *KubernetesRetrierClient) DeleteAWSIamConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.DeleteAWSIamConfig(ctx, cluster, name, namespace)
		},
	)
}

// DeleteOIDCConfig deletes a OIDCConfig object from the cluster.
func (c *KubernetesRetrierClient) DeleteOIDCConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.DeleteOIDCConfig(ctx, cluster, name, namespace)
		},
	)
}

// DeleteCluster deletes a CAPI Cluster from the cluster.
func (c *KubernetesRetrierClient) DeleteCluster(ctx context.Context, cluster, clusterToDelete *types.Cluster) error {
	return c.retrier.Retry(
		func() error {
			return c.KubernetesClient.DeleteCluster(ctx, cluster, clusterToDelete)
		},
	)
}
