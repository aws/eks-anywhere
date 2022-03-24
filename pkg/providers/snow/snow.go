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
	providerKubectlClient ProviderKubectlClient
	writer                filewriter.FileWriter
	retrier               *retrier.Retrier
	bootstrapCreds        bootstrapCreds
}

type ProviderKubectlClient interface {
	DeleteEksaDatacenterConfig(ctx context.Context, snowDatacenterResourceType string, snowDatacenterConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaMachineConfig(ctx context.Context, snowMachineResourceType string, snowMachineConfigName string, kubeconfigFile string, namespace string) error
}

func NewProvider(providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, now types.NowFunc) *snowProvider {
	retrier := retrier.NewWithMaxRetries(maxRetries, backOffPeriod)
	return &snowProvider{
		providerKubectlClient: providerKubectlClient,
		writer:                writer,
		retrier:               retrier,
	}
}

func (p *snowProvider) Name() string {
	return constants.SnowProviderName
}

func (p *snowProvider) setupMachineConfigs(clusterSpec *cluster.Spec) {
	controlPlaneMachineName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.SnowMachineConfigs[controlPlaneMachineName].Annotations = map[string]string{clusterSpec.Cluster.ControlPlaneAnnotation(): "true"}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		clusterSpec.SnowMachineConfigs[etcdMachineName].Annotations = map[string]string{clusterSpec.Cluster.EtcdAnnotation(): "true"}
	}

	if clusterSpec.Cluster.IsManaged() {
		for _, mc := range clusterSpec.SnowMachineConfigs {
			mc.SetManagedBy(clusterSpec.Cluster.ManagedBy())
		}
	}
}

func (p *snowProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := p.setupBootstrapCreds(); err != nil {
		return fmt.Errorf("failed setting up credentials: %v", err)
	}
	p.setupMachineConfigs(clusterSpec)
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

func ControlPlaneObjects(clusterSpec *cluster.Spec, machineConfigs map[string]*v1alpha1.SnowMachineConfig) ([]runtime.Object, error) {
	snowCluster := SnowCluster(clusterSpec)
	controlPlaneMachineTemplate := SnowMachineTemplate(machineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
	kubeadmControlPlane, err := KubeadmControlPlane(clusterSpec, controlPlaneMachineTemplate)
	if err != nil {
		return nil, err
	}
	capiCluster := CAPICluster(clusterSpec, snowCluster, kubeadmControlPlane)

	return []runtime.Object{capiCluster, snowCluster, kubeadmControlPlane, controlPlaneMachineTemplate}, nil
}

func WorkersObjects(clusterSpec *cluster.Spec, machineConfigs map[string]*v1alpha1.SnowMachineConfig) ([]runtime.Object, error) {
	kubeadmConfigTemplates, err := KubeadmConfigTemplates(clusterSpec)
	if err != nil {
		return nil, err
	}
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

	return workersObjs, nil
}

func (p *snowProvider) GenerateCAPISpecForCreate(ctx context.Context, _ *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneObjs, err := ControlPlaneObjects(clusterSpec, clusterSpec.SnowMachineConfigs)
	if err != nil {
		return nil, nil, err
	}

	controlPlaneSpec, err = templater.ObjectsToYaml(controlPlaneObjs...)
	if err != nil {
		return nil, nil, err
	}

	workersObjs, err := WorkersObjects(clusterSpec, clusterSpec.SnowMachineConfigs)
	if err != nil {
		return nil, nil, err
	}

	workersSpec, err = templater.ObjectsToYaml(workersObjs...)
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

func (p *snowProvider) DatacenterConfig(clusterSpec *cluster.Spec) providers.DatacenterConfig {
	return clusterSpec.SnowDatacenter
}

func (p *snowProvider) DatacenterResourceType() string {
	return snowDatacenterResourceType
}

func (p *snowProvider) MachineResourceType() string {
	return snowMachineResourceType
}

func (p *snowProvider) MachineConfigs(clusterSpec *cluster.Spec) []providers.MachineConfig {
	configs := make([]providers.MachineConfig, 0, len(clusterSpec.SnowMachineConfigs))
	for _, mc := range clusterSpec.SnowMachineConfigs {
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
	for _, mc := range clusterSpec.SnowMachineConfigs {
		if err := p.providerKubectlClient.DeleteEksaMachineConfig(ctx, snowMachineResourceType, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaDatacenterConfig(ctx, snowDatacenterResourceType, clusterSpec.SnowDatacenter.GetName(), clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.SnowDatacenter.GetNamespace())
}

func (p *snowProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	return nil
}

func (p *snowProvider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	return nil
}
