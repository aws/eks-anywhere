package providers

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Provider interface {
	Name() string
	SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error
	SetupAndValidateDeleteCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, currentSpec *cluster.Spec) error
	SetupAndValidateUpgradeManagementComponents(ctx context.Context, clusterSpec *cluster.Spec) error
	UpdateSecrets(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	GenerateCAPISpecForCreate(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error)
	GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currrentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error)
	// PreCAPIInstallOnBootstrap is called after the bootstrap cluster is setup but before CAPI resources are installed on it. This allows us to do provider specific configuration on the bootstrap cluster.
	PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error
	PostBootstrapDeleteForUpgrade(ctx context.Context, cluster *types.Cluster) error
	PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error
	// PostWorkloadInit is called after the workload cluster is created and initialized with a CNI. This allows us to do provider specific configuration on the workload cluster.
	PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	BootstrapClusterOpts(clusterSpec *cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error)
	UpdateKubeConfig(content *[]byte, clusterName string) error
	Version(components *cluster.ManagementComponents) string
	EnvMap(managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec) (map[string]string, error)
	GetDeployments() map[string][]string
	GetInfrastructureBundle(components *cluster.ManagementComponents) *types.InfrastructureBundle
	DatacenterConfig(clusterSpec *cluster.Spec) DatacenterConfig
	DatacenterResourceType() string
	MachineResourceType() string
	MachineConfigs(clusterSpec *cluster.Spec) []MachineConfig
	ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	ChangeDiff(currentComponents, newComponents *cluster.ManagementComponents) *types.ComponentChangeDiff
	RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error
	UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec, cluster *types.Cluster) (bool, error)
	DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error
	InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error
	PostClusterDeleteValidate(ctx context.Context, managementCluster *types.Cluster) error
	// PostMoveManagementToBootstrap is called after the CAPI management is moved back to the bootstrap cluster.
	PostMoveManagementToBootstrap(ctx context.Context, bootstrapCluster *types.Cluster) error
	PreCoreComponentsUpgrade(ctx context.Context, cluster *types.Cluster, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec) error
}

type DatacenterConfig interface {
	Kind() string
	PauseReconcile()
	ClearPauseAnnotation()
	Marshallable() v1alpha1.Marshallable
}

type BuildMapOption func(map[string]interface{})

type TemplateBuilder interface {
	GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...BuildMapOption) (content []byte, err error)
	GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error)
}

type MachineConfig interface {
	OSFamily() v1alpha1.OSFamily
	Marshallable() v1alpha1.Marshallable
	GetNamespace() string
	GetName() string
}
