package createvalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateGitOps(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, spec *cluster.Spec) error {
	if spec.GitOpsConfig == nil || spec.Cluster.IsSelfManaged() {
		logger.V(5).Info("skipping ValidateGitOps")
		return nil
	}

	existingGitOps, err := k.SearchEksaGitOpsConfig(ctx, spec.Cluster.Spec.GitOpsRef.Name, cluster.KubeconfigFile, spec.Cluster.Namespace)
	if err != nil {
		return err
	}
	if len(existingGitOps) > 0 {
		return fmt.Errorf("gitOpsConfig %s already exists", spec.Cluster.Spec.GitOpsRef.Name)
	}

	err = validateWorkloadFields(ctx, k, cluster, spec)
	if err != nil {
		return fmt.Errorf("workload cluster gitOpsConfig is invalid: %v", err)
	}
	return nil
}

func validateWorkloadFields(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, spec *cluster.Spec) error {
	mgmtCluster, err := k.GetEksaCluster(ctx, cluster, cluster.Name)
	if err != nil {
		return err
	}
	mgmtGitOps, err := k.GetEksaGitOpsConfig(ctx, mgmtCluster.Spec.GitOpsRef.Name, cluster.KubeconfigFile, mgmtCluster.Namespace)
	if err != nil {
		return err
	}

	if !mgmtGitOps.Spec.Equal(&spec.GitOpsConfig.Spec) {
		return fmt.Errorf("expected gitOpsConfig to be the same between management and its workload clusters")
	}

	return nil
}
