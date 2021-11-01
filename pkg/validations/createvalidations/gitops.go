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
	if spec.GitOpsConfig == nil || spec.IsSelfManaged() {
		logger.V(5).Info("skipping ValidateGitOps")
		return nil
	}

	existingGitOps, err := k.SearchEksaGitOpsConfig(ctx, spec.Spec.GitOpsRef.Name, cluster.KubeconfigFile, spec.Namespace)
	if err != nil {
		return err
	}
	if len(existingGitOps) > 0 {
		return fmt.Errorf("gitOpsConfig %s already exists", spec.Spec.GitOpsRef.Name)
	}

	mgmtCluster, err := k.GetEksaCluster(ctx, cluster, cluster.Name)
	if err != nil {
		return err
	}
	mgmtGitOps, err := k.GetEksaGitOpsConfig(ctx, mgmtCluster.Spec.GitOpsRef.Name, cluster.KubeconfigFile, mgmtCluster.Namespace)
	if err != nil {
		return err
	}
	if mgmtGitOps.Spec.Flux.Github.ClusterRootPath() != spec.GitOpsConfig.Spec.Flux.Github.ClusterRootPath() {
		return fmt.Errorf("gitOpsConfig.Spec.Flux.ClusterConfigPath: %s is invalid: expect workload cluster's GitOps clusterConfigPath to share the same parent directory as managaement cluster's", spec.GitOpsConfig.Spec.Flux.Github.ClusterConfigPath)
	}

	return nil
}
