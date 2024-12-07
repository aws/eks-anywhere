package interfaces

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

// ClientFactory builds Kubernetes clients.
type ClientFactory interface {
	// BuildClientFromKubeconfig builds a Kubernetes client from a kubeconfig file.
	BuildClientFromKubeconfig(kubeconfigPath string) (kubernetes.Client, error)
}

type Bootstrapper interface {
	CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...bootstrapper.BootstrapClusterOption) (*types.Cluster, error)
	DeleteBootstrapCluster(context.Context, *types.Cluster, constants.Operation, bool) error
}

type ClusterManager interface {
	BackupCAPI(ctx context.Context, cluster *types.Cluster, managementStatePath, clusterName string) error
	BackupCAPIWaitForInfrastructure(ctx context.Context, cluster *types.Cluster, managementStatePath, clusterName string) error
	MoveCAPI(ctx context.Context, from, to *types.Cluster, clusterName string, clusterSpec *cluster.Spec, checkers ...types.NodeReadyChecker) error
	PauseCAPIWorkloadClusters(ctx context.Context, managementCluster *types.Cluster) error
	ResumeCAPIWorkloadClusters(ctx context.Context, managementCluster *types.Cluster) error
	InstallCAPI(ctx context.Context, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error
	SaveLogsManagementCluster(ctx context.Context, spec *cluster.Spec, cluster *types.Cluster) error
	SaveLogsWorkloadCluster(ctx context.Context, provider providers.Provider, spec *cluster.Spec, cluster *types.Cluster) error
	CreateEKSANamespace(ctx context.Context, cluster *types.Cluster) error
	ApplyBundles(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	ApplyReleases(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	PauseEKSAControllerReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	GetCurrentClusterSpec(ctx context.Context, cluster *types.Cluster, clusterName string) (*cluster.Spec, error)
	Upgrade(ctx context.Context, cluster *types.Cluster, currentManagementComponents, newManagementComponents *cluster.ManagementComponents, newSpec *cluster.Spec) (*types.ChangeDiff, error)
	CreateRegistryCredSecret(ctx context.Context, mgmt *types.Cluster) error
	ResumeEKSAControllerReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	AllowDeleteWhilePaused(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
}

type GitOpsManager interface {
	InstallGitOps(ctx context.Context, cluster *types.Cluster, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error
	PauseClusterResourcesReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	ResumeClusterResourcesReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error
	UpdateGitEksaSpec(ctx context.Context, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error
	ForceReconcileGitRepo(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	Validations(ctx context.Context, clusterSpec *cluster.Spec) []validations.Validation
	CleanupGitRepo(ctx context.Context, clusterSpec *cluster.Spec) error
	Install(ctx context.Context, cluster *types.Cluster, managementComponents *cluster.ManagementComponents, oldSpec, newSpec *cluster.Spec) error
	Upgrade(ctx context.Context, cluster *types.Cluster, currentManagementComponents, newManagementComponents *cluster.ManagementComponents, oldSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error)
}

type Validator interface {
	PreflightValidations(ctx context.Context) []validations.Validation
}

type CAPIManager interface {
	Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currentManagementComponents, newManagementComponents *cluster.ManagementComponents, newSpec *cluster.Spec) (*types.ChangeDiff, error)
	EnsureEtcdProvidersInstallation(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, managementComponents *cluster.ManagementComponents, currSpec *cluster.Spec) error
}

type EksdInstaller interface {
	InstallEksdCRDs(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	InstallEksdManifest(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
}

type EksdUpgrader interface {
	Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) error
}

// PackageManager handles installation and upgrades of curated packages.
type PackageManager interface {
	InstallCuratedPackages(ctx context.Context)
	UpgradeCuratedPackages(ctx context.Context)
}

// ClusterUpgrader upgrades the cluster and waits until it's ready.
type ClusterUpgrader interface {
	Run(ctx context.Context, spec *cluster.Spec, managementCluster types.Cluster) error
}

// ClusterCreator creates the cluster and waits until it's ready.
type ClusterCreator interface {
	Run(ctx context.Context, spec *cluster.Spec, managementCluster types.Cluster) error
	CreateSync(ctx context.Context, spec *cluster.Spec, managementCluster *types.Cluster) (*types.Cluster, error)
}

// EksaInstaller installs the EKS-A controllers and CRDs.
type EksaInstaller interface {
	Install(ctx context.Context, log logr.Logger, cluster *types.Cluster, managementComponents *cluster.ManagementComponents, spec *cluster.Spec) error
}

// ClusterDeleter deletes the cluster.
type ClusterDeleter interface {
	Run(ctx context.Context, spec *cluster.Spec, managementCluster types.Cluster) error
}

// ClusterMover moves the EKS-A cluster.
type ClusterMover interface {
	Move(ctx context.Context, spec *cluster.Spec, srcClient, dstClient kubernetes.Client) error
}

// AwsIamAuth is responsible for generating iam kubeconfigs.
type AwsIamAuth interface {
	GenerateWorkloadKubeconfig(ctx context.Context, management, workload *types.Cluster, spec *cluster.Spec) error
	GenerateManagementKubeconfig(ctx context.Context, cluster *types.Cluster) error
}
