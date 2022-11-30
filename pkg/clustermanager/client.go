package clustermanager

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/types"
)

// KubernetesClient allows to interact with the k8s api server.
type KubernetesClient interface {
	Apply(ctx context.Context, kubeconfigPath string, obj runtime.Object) error
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error
	ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error
	WaitForDeployment(ctx context.Context, cluster *types.Cluster, timeout string, condition string, target string, namespace string) error
	UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error
	RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error
}

type client struct {
	ClusterClient
}

func NewClient(clusterClient ClusterClient) *client {
	return &client{ClusterClient: clusterClient}
}

func (c *client) waitForDeployments(ctx context.Context, deploymentsByNamespace map[string][]string, cluster *types.Cluster) error {
	for namespace, deployments := range deploymentsByNamespace {
		for _, deployment := range deployments {
			err := c.WaitForDeployment(ctx, cluster, deploymentWaitStr, "Available", deployment, namespace)
			if err != nil {
				return fmt.Errorf("waiting for %s in namespace %s: %v", deployment, namespace, err)
			}
		}
	}
	return nil
}
