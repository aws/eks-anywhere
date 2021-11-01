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

	return validateWorkloadFields(ctx, k, cluster, spec)
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

	mg := mgmtGitOps.Spec.Flux.Github
	wg := spec.GitOpsConfig.Spec.Flux.Github

	if mg.ClusterRootPath() != wg.ClusterRootPath() {
		return fmt.Errorf("gitOpsConfig.spec.flux.clusterConfigPath: %s is invalid: expected workload cluster's GitOps clusterConfigPath to share the same parent directory as management cluster's", wg.ClusterConfigPath)
	}
	if mg.Branch != wg.Branch {
		return fmt.Errorf("gitOpsConfig.spec.flux.branch must be same as management cluster's. want: %s, got: %s", mg.Branch, wg.Branch)
	}
	if mg.Owner != wg.Owner {
		return fmt.Errorf("gitOpsConfig.spec.flux.owner must be same as management cluster's. want: %s, got: %s", mg.Owner, wg.Owner)
	}
	if mg.Repository != wg.Repository {
		return fmt.Errorf("gitOpsConfig.spec.flux.repository must be same as management cluster's. want: %s, got: %s", mg.Repository, wg.Repository)
	}
	if mg.FluxSystemNamespace != wg.FluxSystemNamespace {
		return fmt.Errorf("gitOpsConfig.spec.flux.fluxSystemNamespace must be same as management cluster's. want: %s, got: %s", mg.FluxSystemNamespace, wg.FluxSystemNamespace)
	}
	if mg.Personal != wg.Personal {
		return fmt.Errorf("gitOpsConfig.spec.flux.personal must be same as management cluster's. want: %v, got: %v", mg.Personal, wg.Personal)
	}
	return nil
}
