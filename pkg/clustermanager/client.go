package clustermanager

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/types"
)

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
