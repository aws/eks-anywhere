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
	SetupAndValidateUpgradeCluster(ctx context.Context, clusterSpec *cluster.Spec) error
	UpdateSecrets(ctx context.Context, cluster *types.Cluster) error
	GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error)
	GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error)
	GenerateStorageClass() []byte
	BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error
	CleanupProviderInfrastructure(ctx context.Context) error
	BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error)
	UpdateKubeConfig(content *[]byte, clusterName string) error
	Version(clusterSpec *cluster.Spec) string
	EnvMap() (map[string]string, error)
	GetDeployments() map[string][]string
	GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle
	DatacenterConfig() DatacenterConfig
	DatacenterResourceType() string
	MachineResourceType() string
	MachineConfigs() []MachineConfig
	ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error
	GenerateMHC() ([]byte, error)
	ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff
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
	GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, buildOptions ...BuildMapOption) (content []byte, err error)
	WorkerMachineTemplateName(clusterName string) string
	CPMachineTemplateName(clusterName string) string
}

type MachineConfig interface {
	OSFamily() v1alpha1.OSFamily
	Marshallable() v1alpha1.Marshallable
}
