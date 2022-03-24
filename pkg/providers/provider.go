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
	SetupAndValidateDeleteCluster(ctx context.Context) error
	SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	UpdateSecrets(ctx context.Context, cluster *types.Cluster) error
	GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error)
	GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currrentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error)
	GenerateStorageClass() []byte
	PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error
	BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error)
	UpdateKubeConfig(content *[]byte, clusterName string) error
	Version(clusterSpec *cluster.Spec) string
	EnvMap(clusterSpec *cluster.Spec) (map[string]string, error)
	GetDeployments() map[string][]string
	GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle
	DatacenterConfig(clusterSpec *cluster.Spec) DatacenterConfig
	DatacenterResourceType() string
	MachineResourceType() string
	MachineConfigs(clusterSpec *cluster.Spec) []MachineConfig
	ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	GenerateMHC() ([]byte, error)
	ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff
	RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error
	UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec) (bool, error)
	DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error
	RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error
	MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string
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
