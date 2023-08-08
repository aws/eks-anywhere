package createvalidations

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

var (
	clusterResourceType      = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)
	fluxConfigResourceType   = fmt.Sprintf("fluxconfigs.%s", v1alpha1.GroupVersion.Group)
	gitOpsConfigResourceType = fmt.Sprintf("gitopsconfigs.%s", v1alpha1.GroupVersion.Group)
)

func ValidateGitOps(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if err := validateGitOpsConfig(ctx, k, cluster, clusterSpec); err != nil {
		return fmt.Errorf("invalid gitOpsConfig: %v", err)
	}

	if err := validateFluxConfig(ctx, k, cluster, clusterSpec); err != nil {
		return fmt.Errorf("invalid fluxConfig: %v", err)
	}

	return nil
}

// validateGitOpsConfig method will be removed in a future release since gitOpsConfig is deprecated in favor of fluxConfig.
func validateGitOpsConfig(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if clusterSpec.GitOpsConfig == nil {
		return nil
	}

	gitOpsConfig := &v1alpha1.GitOpsConfig{}
	err := k.GetObject(ctx, gitOpsConfigResourceType, clusterSpec.GitOpsConfig.Name, clusterSpec.Cluster.Namespace, cluster.KubeconfigFile, gitOpsConfig)
	if err == nil {
		return fmt.Errorf("gitOpsConfig %s already exists", clusterSpec.Cluster.Spec.GitOpsRef.Name)
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("fetching gitOpsConfig in cluster: %v", err)
	}

	mgmtCluster, err := k.GetEksaCluster(ctx, clusterSpec.ManagementCluster, clusterSpec.Cluster.ManagedBy())
	if err != nil {
		return err
	}

	mgmtGitOpsConfig := &v1alpha1.GitOpsConfig{}
	if err := k.GetObject(ctx, gitOpsConfigResourceType, mgmtCluster.Spec.GitOpsRef.Name, mgmtCluster.Namespace, cluster.KubeconfigFile, mgmtGitOpsConfig); err != nil {
		return err
	}

	if !mgmtGitOpsConfig.Spec.Equal(&clusterSpec.GitOpsConfig.Spec) {
		return errors.New("expected gitOpsConfig.spec to be the same between management and its workload clusters")
	}

	return nil
}

func validateFluxConfig(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if clusterSpec.FluxConfig == nil {
		return nil
	}

	// when processing deprecated gitopsConfig, we parse and convert it to fluxConfig.
	// for this case both fluxConfig and gitopsConfig can exist in spec. Skip fluxConfig validation.
	if clusterSpec.GitOpsConfig != nil {
		return nil
	}

	fluxConfig := &v1alpha1.FluxConfig{}
	err := k.GetObject(ctx, fluxConfigResourceType, clusterSpec.FluxConfig.Name, clusterSpec.Cluster.Namespace, cluster.KubeconfigFile, fluxConfig)
	if err == nil {
		return fmt.Errorf("fluxConfig %s already exists", clusterSpec.Cluster.Spec.GitOpsRef.Name)
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("fetching fluxConfig in cluster: %v", err)
	}

	mgmtCluster, err := k.GetEksaCluster(ctx, clusterSpec.ManagementCluster, clusterSpec.Cluster.ManagedBy())
	if err != nil {
		return err
	}

	mgmtFluxConfig := &v1alpha1.FluxConfig{}
	if err := k.GetObject(ctx, fluxConfigResourceType, mgmtCluster.Spec.GitOpsRef.Name, clusterSpec.Cluster.Namespace, cluster.KubeconfigFile, mgmtFluxConfig); err != nil {
		return err
	}

	if !mgmtFluxConfig.Spec.Equal(&clusterSpec.FluxConfig.Spec) {
		return errors.New("expected fluxConfig.spec to be the same between management and its workload clusters")
	}

	return nil
}
