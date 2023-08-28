package clustermanager

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/types"
)

// KubernetesClient allows to interact with the k8s api server.
type KubernetesClient interface {
	Apply(ctx context.Context, kubeconfigPath string, obj runtime.Object, opts ...kubernetes.KubectlApplyOption) error
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error
	ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error
	WaitForDeployment(ctx context.Context, cluster *types.Cluster, timeout string, condition string, target string, namespace string) error
	UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error
	RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error
}

type clusterManagerClient struct {
	ClusterClient
}

func newClient(clusterClient ClusterClient) *clusterManagerClient {
	return &clusterManagerClient{ClusterClient: clusterClient}
}

func (c *clusterManagerClient) waitForDeployments(ctx context.Context, deploymentsByNamespace map[string][]string, cluster *types.Cluster, timeout string) error {
	for namespace, deployments := range deploymentsByNamespace {
		for _, deployment := range deployments {
			err := c.WaitForDeployment(ctx, cluster, timeout, "Available", deployment, namespace)
			if err != nil {
				return fmt.Errorf("waiting for %s in namespace %s: %v", deployment, namespace, err)
			}
		}
	}
	return nil
}
