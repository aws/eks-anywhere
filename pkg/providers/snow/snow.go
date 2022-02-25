package snow

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	eksaSnowCredentialsFileKey = "EKSA_SNOW_DEVICES_CREDENTIALS_FILE"
	eksaSnowCABundlesFileKey   = "EKSA_SNOW_DEVICES_CA_BUNDLES_FILE"
	snowCredentialsKey         = "AWS_B64ENCODED_CREDENTIALS"
	snowCertsKey               = "AWS_B64ENCODED_CA_BUNDLES"
	maxRetries                 = 30
	backOffPeriod              = 5 * time.Second
)

var requiredEnvs = []string{
	"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", "AWS_SESSION_TOKEN",
}

type snowProvider struct {
	datacenterConfig      *v1alpha1.SnowDatacenterConfig
	machineConfigs        map[string]*v1alpha1.SnowMachineConfig
	clusterConfig         *v1alpha1.Cluster
	providerKubectlClient ProviderKubectlClient
	writer                filewriter.FileWriter
	templateBuilder       *SnowTemplateBuilder
	retrier               *retrier.Retrier
	bootstrapCreds        bootstrapCreds
}

type ProviderKubectlClient interface{}

func NewProvider(datacenterConfig *v1alpha1.SnowDatacenterConfig, machineConfigs map[string]*v1alpha1.SnowMachineConfig, clusterConfig *v1alpha1.Cluster, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, now types.NowFunc) *snowProvider {
	var controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.SnowMachineConfigSpec
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.SnowMachineConfigSpec, len(machineConfigs))

	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}
	retrier := retrier.NewWithMaxRetries(maxRetries, backOffPeriod)
	return &snowProvider{
		datacenterConfig:      datacenterConfig,
		machineConfigs:        machineConfigs,
		clusterConfig:         clusterConfig,
		providerKubectlClient: providerKubectlClient,
		writer:                writer,
		templateBuilder: &SnowTemplateBuilder{
			datacenterSpec:              &datacenterConfig.Spec,
			controlPlaneMachineSpec:     controlPlaneMachineSpec,
			workerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
			etcdMachineSpec:             etcdMachineSpec,
			now:                         now,
		},
		retrier: retrier,
	}
}

type SnowTemplateBuilder struct {
	datacenterSpec              *v1alpha1.SnowDatacenterConfigSpec
	controlPlaneMachineSpec     *v1alpha1.SnowMachineConfigSpec
	workerNodeGroupMachineSpecs map[string]v1alpha1.SnowMachineConfigSpec
	etcdMachineSpec             *v1alpha1.SnowMachineConfigSpec
	now                         types.NowFunc
}

func (p *snowProvider) Name() string {
	return constants.SnowProviderName
}

// TODO: move this to validator
func validateEnvsForEcrRegistry() error {
	// get aws credentials for the private ecr registry
	for _, key := range requiredEnvs {
		if env, ok := os.LookupEnv(key); !ok || len(env) <= 0 {
			return fmt.Errorf("warning required env not set %s", key)
		}
	}
	return nil
}

func (p *snowProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	// TODO: remove this validation when capas image is public.
	if err := validateEnvsForEcrRegistry(); err != nil {
		return fmt.Errorf("failed checking aws credentials for private ecr: %v", err)
	}

	if err := p.setupBootstrapCreds(); err != nil {
		return fmt.Errorf("failed setting up credentials: %v", err)
	}
	return nil
}

func (p *snowProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *snowProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	return nil
}

func (p *snowProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func (p *snowProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return nil, nil, nil
}

func (p *snowProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currrentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return nil, nil, nil
}

func (p *snowProvider) GenerateStorageClass() []byte {
	return nil
}

func (p *snowProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *snowProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	return []bootstrapper.BootstrapClusterOption{bootstrapper.WithExtraDockerMounts()}, nil
}

func (p *snowProvider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

func (p *snowProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.Snow.Version
}

func (p *snowProvider) EnvMap() (map[string]string, error) {
	envMap := make(map[string]string)
	envMap[snowCredentialsKey] = p.bootstrapCreds.snowCredsB64
	envMap[snowCertsKey] = p.bootstrapCreds.snowCertsB64

	// TODO: remove $DEVICE_IPS whenever CAPAS removes it as a required env
	envMap["DEVICE_IPS"] = ""

	// TODO: tmp solution to pull private ECR
	envMap["SNOW_CONTROLLER_IMAGE"] = fmt.Sprintf("%s/cluster-api-provider-aws-snow:latest", constants.EcrRegistry)
	envMap["ECR_CREDS"] = constants.EcrRegistrySecretName

	return envMap, nil
}

func (p *snowProvider) GetDeployments() map[string][]string {
	return nil
}

func (p *snowProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	bundle := clusterSpec.VersionsBundle
	folderName := fmt.Sprintf("infrastructure-snow/%s/", bundle.Snow.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			bundle.Snow.Components,
			bundle.Snow.Metadata,
		},
	}
	return &infraBundle
}

func (p *snowProvider) DatacenterConfig() providers.DatacenterConfig {
	return nil
}

func (p *snowProvider) DatacenterResourceType() string {
	return ""
}

func (p *snowProvider) MachineResourceType() string {
	return ""
}

func (p *snowProvider) MachineConfigs() []providers.MachineConfig {
	return nil
}

func (p *snowProvider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *snowProvider) GenerateMHC() ([]byte, error) {
	return nil, nil
}

func (p *snowProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	return nil
}

func (p *snowProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	return nil
}

func (p *snowProvider) UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec) (bool, error) {
	return false, nil
}

func (p *snowProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *snowProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	return nil
}

func (p *snowProvider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	return nil
}
