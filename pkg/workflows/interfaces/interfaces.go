package interfaces

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type Bootstrapper interface {
	CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...bootstrapper.BootstrapClusterOption) (*types.Cluster, error)
	DeleteBootstrapCluster(context.Context, *types.Cluster, constants.Operation, bool) error
}

type ClusterManager interface {
	BackupCAPI(ctx context.Context, cluster *types.Cluster, managementStatePath string) error
	MoveCAPI(ctx context.Context, from, to *types.Cluster, clusterName string, clusterSpec *cluster.Spec, checkers ...types.NodeReadyChecker) error
	CreateWorkloadCluster(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) (*types.Cluster, error)
	PauseCAPIWorkloadClusters(ctx context.Context, managementCluster *types.Cluster) error
	ResumeCAPIWorkloadClusters(ctx context.Context, managementCluster *types.Cluster) error
	RunPostCreateWorkloadCluster(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) error
	UpgradeCluster(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	DeleteCluster(ctx context.Context, managementCluster, clusterToDelete *types.Cluster, provider providers.Provider, clusterSpec *cluster.Spec) error
	InstallCAPI(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error
	InstallNetworking(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	UpgradeNetworking(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, provider providers.Provider) (*types.ChangeDiff, error)
	SaveLogsManagementCluster(ctx context.Context, spec *cluster.Spec, cluster *types.Cluster) error
	SaveLogsWorkloadCluster(ctx context.Context, provider providers.Provider, spec *cluster.Spec, cluster *types.Cluster) error
	InstallCustomComponents(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error
	CreateEKSANamespace(ctx context.Context, cluster *types.Cluster) error
	CreateEKSAResources(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error
	ApplyBundles(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	ApplyReleases(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	PauseEKSAControllerReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	ResumeEKSAControllerReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	EKSAClusterSpecChanged(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (bool, error)
	InstallMachineHealthChecks(ctx context.Context, clusterSpec *cluster.Spec, workloadCluster *types.Cluster) error
	GetCurrentClusterSpec(ctx context.Context, cluster *types.Cluster, clusterName string) (*cluster.Spec, error)
	Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error)
	InstallAwsIamAuth(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) error
	CreateAwsIamAuthCaSecret(ctx context.Context, bootstrapCluster *types.Cluster, workloadClusterName string) error
	DeletePackageResources(ctx context.Context, managementCluster *types.Cluster, clusterName string) error
}

type GitOpsManager interface {
	InstallGitOps(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error
	PauseClusterResourcesReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	ResumeClusterResourcesReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	UpdateGitEksaSpec(ctx context.Context, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error
	ForceReconcileGitRepo(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	Validations(ctx context.Context, clusterSpec *cluster.Spec) []validations.Validation
	CleanupGitRepo(ctx context.Context, clusterSpec *cluster.Spec) error
	Install(ctx context.Context, cluster *types.Cluster, oldSpec, newSpec *cluster.Spec) error
	Upgrade(ctx context.Context, cluster *types.Cluster, oldSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error)
}

type Validator interface {
	PreflightValidations(ctx context.Context) []validations.Validation
}

type CAPIManager interface {
	Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error)
	EnsureEtcdProvidersInstallation(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currSpec *cluster.Spec) error
}

type EksdInstaller interface {
	InstallEksdCRDs(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	InstallEksdManifest(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
}

type EksdUpgrader interface {
	Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error)
}

type PackageInstaller interface {
	InstallCuratedPackages(ctx context.Context)
}

// ClusterUpgrader prepares the cluster for an upgrade.
type ClusterUpgrader interface {
	PrepareUpgrade(ctx context.Context, spec *cluster.Spec, managementClusterKubeconfigPath, workloadClusterKubeconfigPath string) error
	CleanupAfterUpgrade(ctx context.Context, spec *cluster.Spec, managementClusterKubeconfigPath, workloadClusterKubeconfigPath string) error
}
