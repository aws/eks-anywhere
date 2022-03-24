package interfaces

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type Bootstrapper interface {
	CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...bootstrapper.BootstrapClusterOption) (*types.Cluster, error)
	DeleteBootstrapCluster(context.Context, *types.Cluster, bool) error
}

type ClusterManager interface {
	MoveCAPI(ctx context.Context, from, to *types.Cluster, clusterName string, clusterSpec *cluster.Spec, checkers ...types.NodeReadyChecker) error
	CreateWorkloadCluster(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) (*types.Cluster, error)
	UpgradeCluster(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	DeleteCluster(ctx context.Context, managementCluster, clusterToDelete *types.Cluster, provider providers.Provider, clusterSpec *cluster.Spec) error
	InstallCAPI(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error
	InstallNetworking(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	UpgradeNetworking(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, provider providers.Provider) (*types.ChangeDiff, error)
	InstallStorageClass(ctx context.Context, cluster *types.Cluster, provider providers.Provider) error
	SaveLogsManagementCluster(ctx context.Context, cluster *types.Cluster) error
	SaveLogsWorkloadCluster(ctx context.Context, provider providers.Provider, spec *cluster.Spec, cluster *types.Cluster) error
	InstallCustomComponents(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	InstallEksdComponents(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	CreateEKSAResources(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error
	ApplyBundles(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	PauseEKSAControllerReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	ResumeEKSAControllerReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	EKSAClusterSpecChanged(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) (bool, error)
	InstallMachineHealthChecks(ctx context.Context, workloadCluster *types.Cluster, provider providers.Provider) error
	GetCurrentClusterSpec(ctx context.Context, cluster *types.Cluster, clusterName string) (*cluster.Spec, error)
	Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error)
	InstallAwsIamAuth(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) error
	CreateAwsIamAuthCaSecret(ctx context.Context, cluster *types.Cluster) error
}

type AddonManager interface {
	InstallGitOps(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error
	PauseGitOpsKustomization(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	ResumeGitOpsKustomization(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	UpdateGitEksaSpec(ctx context.Context, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error
	ForceReconcileGitRepo(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	Validations(ctx context.Context, clusterSpec *cluster.Spec) []validations.Validation
	CleanupGitRepo(ctx context.Context, clusterSpec *cluster.Spec) error
	Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec *cluster.Spec, newSpec *cluster.Spec) (*types.ChangeDiff, error)
	UpdateLegacyFileStructure(ctx context.Context, currentSpec, newSpec *cluster.Spec) error
}

type Validator interface {
	PreflightValidations(ctx context.Context) error
}

type CAPIManager interface {
	Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error)
	EnsureEtcdProvidersInstallation(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currSpec *cluster.Spec) error
}
