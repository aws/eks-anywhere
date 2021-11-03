package upgradevalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateClusterObjectExists(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster) error {
	c, err := k.GetClusters(ctx, cluster)
	if err != nil {
		return err
	}
	if len(c) == 0 {
		return fmt.Errorf("no CAPI cluster objects present on workload cluster %s", cluster.Name)
	}
	for _, capiCluster := range c {
		if capiCluster.Metadata.Name == cluster.Name {
			return nil
		}
	}
	return fmt.Errorf("couldn't find CAPI cluster object for cluster with name %s", cluster.Name)
}

func ValidateMachineConfigsNameUnique(ctx context.Context, k validations.KubectlClient, p providers.Provider, cluster *types.Cluster, clusterSpec *cluster.Spec, prevSpec *v1alpha1.Cluster) error {
	cpmc := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	if prevSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name != cpmc {
		exists, err := p.MachineConfigExists(ctx, cpmc, cluster.KubeconfigFile, clusterSpec.GetNamespace())
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("control plane machineconfig %s already exists", cpmc)
		}
	}

	if len(clusterSpec.Spec.WorkerNodeGroupConfigurations) > 0 && clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef != nil &&
		len(prevSpec.Spec.WorkerNodeGroupConfigurations) > 0 && prevSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef != nil {
		wnmc := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
		if prevSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name != wnmc {
			exists, err := p.MachineConfigExists(ctx, wnmc, cluster.KubeconfigFile, clusterSpec.GetNamespace())
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("worker nodes machineconfig %s already exists", wnmc)
			}
		}
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil && clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil &&
		prevSpec.Spec.ExternalEtcdConfiguration != nil && prevSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil {
		etcdmc := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		if prevSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name != etcdmc {
			exists, err := p.MachineConfigExists(ctx, etcdmc, cluster.KubeconfigFile, clusterSpec.GetNamespace())
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("external etcd machineconfig %s already exists", etcdmc)
			}
		}
	}

	return nil
}
