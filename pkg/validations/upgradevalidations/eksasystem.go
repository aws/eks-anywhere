package upgradevalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	eksaControllerDeploymentName = "eksa-controller-manager"
)

func ValidateEksaSystemComponents(ctx context.Context, k *executables.Kubectl, cluster *types.Cluster) error {
	deployments, err := k.GetDeployments(ctx, executables.WithCluster(cluster), executables.WithNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return fmt.Errorf("getting deployments in namespace %s: %v", constants.EksaSystemNamespace, err)
	}
	for _, d := range deployments {
		if d.Name == eksaControllerDeploymentName {
			ready := d.Status.ReadyReplicas
			actual := d.Status.Replicas
			if actual == 0 {
				return fmt.Errorf("EKS-A controller deployment %s in namespace %s is scaled to 0 replicas; should be at least one replcias", eksaControllerDeploymentName, constants.EksaSystemNamespace)
			}
			if ready != actual {
				return fmt.Errorf("EKS-A controller deployment %s replicas in namespace %s are not ready; ready=%d, want=%d", eksaControllerDeploymentName, constants.EksaSystemNamespace, ready, actual)
			}
			return nil
		}
	}
	return fmt.Errorf("failed to find EKS-A controller deployment %s in namespace %s", eksaControllerDeploymentName, constants.EksaSystemNamespace)
}
