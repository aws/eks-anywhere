package snow

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
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

var (
	snowDatacenterResourceType = fmt.Sprintf("snowdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	snowMachineResourceType    = fmt.Sprintf("snowmachineconfigs.%s", v1alpha1.GroupVersion.Group)
)

type snowProvider struct {
	// TODO: once cluster.config is available, remove below objs
	datacenterConfig      *v1alpha1.SnowDatacenterConfig
	machineConfigs        map[string]*v1alpha1.SnowMachineConfig
	clusterConfig         *v1alpha1.Cluster
	providerKubectlClient ProviderKubectlClient
	writer                filewriter.FileWriter
	retrier               *retrier.Retrier
	bootstrapCreds        bootstrapCreds
}

type ProviderKubectlClient interface {
	DeleteEksaDatacenterConfig(ctx context.Context, snowDatacenterResourceType string, snowDatacenterConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaMachineConfig(ctx context.Context, snowMachineResourceType string, snowMachineConfigName string, kubeconfigFile string, namespace string) error
}

func NewProvider(datacenterConfig *v1alpha1.SnowDatacenterConfig, machineConfigs map[string]*v1alpha1.SnowMachineConfig, clusterConfig *v1alpha1.Cluster, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, now types.NowFunc) *snowProvider {
	retrier := retrier.NewWithMaxRetries(maxRetries, backOffPeriod)
	return &snowProvider{
		datacenterConfig:      datacenterConfig,
		machineConfigs:        machineConfigs,
		clusterConfig:         clusterConfig,
		providerKubectlClient: providerKubectlClient,
		writer:                writer,
		retrier:               retrier,
	}
}

func (p *snowProvider) Name() string {
	return constants.SnowProviderName
}

func (p *snowProvider) setupMachineConfigs() {
	controlPlaneMachineName := p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	p.machineConfigs[controlPlaneMachineName].Annotations = map[string]string{p.clusterConfig.ControlPlaneAnnotation(): "true"}

	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		p.machineConfigs[etcdMachineName].Annotations = map[string]string{p.clusterConfig.EtcdAnnotation(): "true"}
	}

	if p.clusterConfig.IsManaged() {
		for _, mc := range p.machineConfigs {
			mc.SetManagedBy(p.clusterConfig.ManagedBy())
		}
	}
}

func (p *snowProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := p.setupBootstrapCreds(); err != nil {
		return fmt.Errorf("failed setting up credentials: %v", err)
	}
	p.setupMachineConfigs()
	return nil
}

func (p *snowProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *snowProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	if err := p.setupBootstrapCreds(); err != nil {
		return fmt.Errorf("failed setting up credentials: %v", err)
	}
	return nil
}

func (p *snowProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func ControlPlaneObjects(clusterSpec *cluster.Spec, machineConfigs map[string]*v1alpha1.SnowMachineConfig) []runtime.Object {
	snowCluster := SnowCluster(clusterSpec)
	controlPlaneMachineTemplate := SnowMachineTemplate(machineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
	kubeadmControlPlane := KubeadmControlPlane(clusterSpec, controlPlaneMachineTemplate)
	capiCluster := CAPICluster(clusterSpec, snowCluster, kubeadmControlPlane)

	return []runtime.Object{capiCluster, snowCluster, kubeadmControlPlane, controlPlaneMachineTemplate}
}

func WorkersObjects(clusterSpec *cluster.Spec, machineConfigs map[string]*v1alpha1.SnowMachineConfig) []runtime.Object {
	kubeadmConfigTemplates := KubeadmConfigTemplates(clusterSpec)
	workerMachineTemplates := SnowMachineTemplates(clusterSpec, machineConfigs)
	machineDeployments := MachineDeployments(clusterSpec, kubeadmConfigTemplates, workerMachineTemplates)

	workersObjs := make([]runtime.Object, 0, len(machineDeployments)+len(kubeadmConfigTemplates)+len(workerMachineTemplates))
	for _, item := range machineDeployments {
		workersObjs = append(workersObjs, item)
	}
	for _, item := range kubeadmConfigTemplates {
		workersObjs = append(workersObjs, item)
	}
	for _, item := range workerMachineTemplates {
		workersObjs = append(workersObjs, item)
	}

	return workersObjs
}

func (p *snowProvider) GenerateCAPISpecForCreate(ctx context.Context, _ *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, err = templater.ObjectsToYaml(ControlPlaneObjects(clusterSpec, p.machineConfigs)...)
	if err != nil {
		return nil, nil, err
	}

	workersSpec, err = templater.ObjectsToYaml(WorkersObjects(clusterSpec, p.machineConfigs)...)
	if err != nil {
		return nil, nil, err
	}

	return controlPlaneSpec, workersSpec, nil
}

func (p *snowProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currrentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return nil, nil, nil
}

func (p *snowProvider) GenerateStorageClass() []byte {
	return nil
}

func (p *snowProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *snowProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	return nil, nil
}

func (p *snowProvider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

func (p *snowProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.Snow.Version
}

func (p *snowProvider) EnvMap(clusterSpec *cluster.Spec) (map[string]string, error) {
	envMap := make(map[string]string)
	envMap[snowCredentialsKey] = p.bootstrapCreds.snowCredsB64
	envMap[snowCertsKey] = p.bootstrapCreds.snowCertsB64

	envMap["SNOW_CONTROLLER_IMAGE"] = clusterSpec.VersionsBundle.Snow.Manager.VersionedImage()

	return envMap, nil
}

func (p *snowProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capas-system": {"capas-controller-manager"},
	}
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
	return p.datacenterConfig
}

func (p *snowProvider) DatacenterResourceType() string {
	return snowDatacenterResourceType
}

func (p *snowProvider) MachineResourceType() string {
	return snowMachineResourceType
}

func (p *snowProvider) MachineConfigs() []providers.MachineConfig {
	configs := make([]providers.MachineConfig, 0, len(p.machineConfigs))
	for _, mc := range p.machineConfigs {
		configs = append(configs, mc)
	}
	return configs
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
	for _, mc := range p.machineConfigs {
		if err := p.providerKubectlClient.DeleteEksaMachineConfig(ctx, snowMachineResourceType, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaDatacenterConfig(ctx, snowDatacenterResourceType, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, p.datacenterConfig.Namespace)
}

func (p *snowProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	return nil
}

func (p *snowProvider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	return nil
}
