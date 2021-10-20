package upgradevalidations

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

type ValidationsKubectlClient interface {
	ValidateControlPlaneNodes(ctx context.Context, cluster *types.Cluster) error
	ValidateWorkerNodes(ctx context.Context, cluster *types.Cluster) error
	ValidateNodes(ctx context.Context, kubeconfig string) error
	ValidateClustersCRD(ctx context.Context, cluster *types.Cluster) error
	Version(ctx context.Context, cluster *types.Cluster) (*executables.VersionResponse, error)
	GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.GitOpsConfig, error)
	GetEksaOIDCConfig(ctx context.Context, oidcConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.OIDCConfig, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
}

func New(opts *UpgradeValidationOpts) *UpgradeValidations {
	return &UpgradeValidations{Opts: opts}
}

type UpgradeValidations struct {
	Opts *UpgradeValidationOpts
}

type UpgradeValidationOpts struct {
	Kubectl         ValidationsKubectlClient
	Spec            *cluster.Spec
	WorkloadCluster *types.Cluster
	Provider        providers.Provider
}
