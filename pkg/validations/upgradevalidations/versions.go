package upgradevalidations

import (
	"context"
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

// ValidateServerVersionSkew validates Kubernetes version skew between upgrades for the CLI.
func ValidateServerVersionSkew(ctx context.Context, newCluster *anywherev1.Cluster, cluster *types.Cluster, mgmtCluster *types.Cluster, kubectl validations.KubectlClient) error {
	managementCluster := cluster
	if !cluster.ExistingManagement {
		managementCluster = mgmtCluster
	}

	eksaCluster, err := kubectl.GetEksaCluster(ctx, managementCluster, newCluster.Name)
	if err != nil {
		return fmt.Errorf("fetching old cluster: %v", err)
	}

	return anywherev1.ValidateKubernetesVersionSkew(newCluster, eksaCluster).ToAggregate()
}
