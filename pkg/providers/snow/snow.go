package snow

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers"
	providerValidator "github.com/aws/eks-anywhere/pkg/providers/validator"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	eksaSnowCredentialsFileKey = "EKSA_AWS_CREDENTIALS_FILE"
	eksaSnowCABundlesFileKey   = "EKSA_AWS_CA_BUNDLES_FILE"
	snowCredentialsKey         = "AWS_B64ENCODED_CREDENTIALS"
	snowCertsKey               = "AWS_B64ENCODED_CA_BUNDLES"
	maxRetries                 = 30
	backOffPeriod              = 5 * time.Second
)

var (
	snowDatacenterResourceType = fmt.Sprintf("snowdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	snowMachineResourceType    = fmt.Sprintf("snowmachineconfigs.%s", v1alpha1.GroupVersion.Group)
)

type SnowProvider struct {
	kubeUnAuthClient KubeUnAuthClient
	retrier          *retrier.Retrier
	bootstrapCreds   bootstrapCreds
	configManager    *ConfigManager
	skipIpCheck      bool
}

type KubeUnAuthClient interface {
	KubeconfigClient(kubeconfig string) kubernetes.Client
	Delete(ctx context.Context, name, namespace, kubeconfig string, obj runtime.Object) error
}

func NewProvider(kubeUnAuthClient KubeUnAuthClient, configManager *ConfigManager, skipIpCheck bool) *SnowProvider {
	retrier := retrier.NewWithMaxRetries(maxRetries, backOffPeriod)
	return &SnowProvider{
		kubeUnAuthClient: kubeUnAuthClient,
		retrier:          retrier,
		configManager:    configManager,
		skipIpCheck:      skipIpCheck,
	}
}

func (p *SnowProvider) Name() string {
	return constants.SnowProviderName
}

func (p *SnowProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := p.setupBootstrapCreds(); err != nil {
		return fmt.Errorf("setting up credentials: %v", err)
	}
	if err := p.configManager.SetDefaultsAndValidate(ctx, clusterSpec.Config); err != nil {
		return fmt.Errorf("setting defaults and validate snow config: %v", err)
	}
	if !p.skipIpCheck {
		if err := providerValidator.ValidateControlPlaneIpUniqueness(clusterSpec.Cluster, &networkutils.DefaultNetClient{}); err != nil {
			return err
		}
	} else {
		logger.Info("Skipping check for whether control plane ip is in use")
	}
	return nil
}

func (p *SnowProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, _ *cluster.Spec) error {
	if err := p.setupBootstrapCreds(); err != nil {
		return fmt.Errorf("setting up credentials: %v", err)
	}
	if err := p.configManager.SetDefaultsAndValidate(ctx, clusterSpec.Config); err != nil {
		return fmt.Errorf("setting defaults and validate snow config: %v", err)
	}
	return nil
}

func (p *SnowProvider) SetupAndValidateDeleteCluster(ctx context.Context, _ *types.Cluster) error {
	if err := p.setupBootstrapCreds(); err != nil {
		return fmt.Errorf("setting up credentials: %v", err)
	}
	return nil
}

func (p *SnowProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func CAPIObjects(ctx context.Context, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneObjs, err := ControlPlaneObjects(ctx, clusterSpec, kubeClient)
	if err != nil {
		return nil, nil, err
	}

	controlPlaneSpec, err = templater.ObjectsToYaml(controlPlaneObjs...)
	if err != nil {
		return nil, nil, err
	}

	workersObjs, err := WorkersObjects(ctx, clusterSpec, kubeClient)
	if err != nil {
		return nil, nil, err
	}

	workersSpec, err = templater.ObjectsToYaml(workersObjs...)
	if err != nil {
		return nil, nil, err
	}

	return controlPlaneSpec, workersSpec, nil
}

func (p *SnowProvider) generateCAPISpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	kubeconfigClient := p.kubeUnAuthClient.KubeconfigClient(cluster.KubeconfigFile)
	return CAPIObjects(ctx, clusterSpec, kubeconfigClient)
}

func (p *SnowProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return p.generateCAPISpec(ctx, cluster, clusterSpec)
}

func (p *SnowProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, _ *types.Cluster, _ *cluster.Spec, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return p.generateCAPISpec(ctx, bootstrapCluster, clusterSpec)
}

func (p *SnowProvider) GenerateStorageClass() []byte {
	return nil
}

func (p *SnowProvider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *SnowProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *SnowProvider) PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *SnowProvider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *SnowProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	return nil, nil
}

func (p *SnowProvider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

func (p *SnowProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.Snow.Version
}

func (p *SnowProvider) EnvMap(clusterSpec *cluster.Spec) (map[string]string, error) {
	envMap := make(map[string]string)
	envMap[snowCredentialsKey] = p.bootstrapCreds.snowCredsB64
	envMap[snowCertsKey] = p.bootstrapCreds.snowCertsB64

	envMap["SNOW_CONTROLLER_IMAGE"] = clusterSpec.VersionsBundle.Snow.Manager.VersionedImage()

	return envMap, nil
}

func (p *SnowProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		constants.CapasSystemNamespace: {"capas-controller-manager"},
	}
}

func (p *SnowProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
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

func (p *SnowProvider) DatacenterConfig(clusterSpec *cluster.Spec) providers.DatacenterConfig {
	return clusterSpec.SnowDatacenter
}

func (p *SnowProvider) DatacenterResourceType() string {
	return snowDatacenterResourceType
}

func (p *SnowProvider) MachineResourceType() string {
	return snowMachineResourceType
}

func (p *SnowProvider) MachineConfigs(clusterSpec *cluster.Spec) []providers.MachineConfig {
	configs := make([]providers.MachineConfig, 0, len(clusterSpec.SnowMachineConfigs))
	for _, mc := range clusterSpec.SnowMachineConfigs {
		configs = append(configs, mc)
	}
	return configs
}

func (p *SnowProvider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *SnowProvider) GenerateMHC(clusterSpec *cluster.Spec) ([]byte, error) {
	return templater.ObjectsToYaml(clusterapi.MachineHealthCheckObjects(clusterSpec)...)
}

func (p *SnowProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	if currentSpec.VersionsBundle.Snow.Version == newSpec.VersionsBundle.Snow.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.SnowProviderName,
		NewVersion:    newSpec.VersionsBundle.Snow.Version,
		OldVersion:    currentSpec.VersionsBundle.Snow.Version,
	}
}

func (p *SnowProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	return nil
}

func bundleImagesEqual(new, old releasev1alpha1.SnowBundle) bool {
	return new.Manager.ImageDigest == old.Manager.ImageDigest && new.KubeVip.ImageDigest == old.KubeVip.ImageDigest
}

func machineConfigsEqual(new, old map[string]*v1alpha1.SnowMachineConfig) bool {
	if len(new) != len(old) {
		return false
	}

	for name, newConfig := range new {
		oldConfig, ok := old[name]
		if !ok || !equality.Semantic.DeepDerivative(newConfig.Spec, oldConfig.Spec) {
			return false
		}
	}

	return true
}

func (p *SnowProvider) UpgradeNeeded(ctx context.Context, newSpec, oldSpec *cluster.Spec, _ *types.Cluster) (bool, error) {
	return !bundleImagesEqual(newSpec.VersionsBundle.Snow, oldSpec.VersionsBundle.Snow) ||
		!machineConfigsEqual(newSpec.SnowMachineConfigs, oldSpec.SnowMachineConfigs), nil
}

func (p *SnowProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range clusterSpec.SnowMachineConfigs {
		if err := p.kubeUnAuthClient.Delete(ctx, mc.Name, mc.Namespace, clusterSpec.ManagementCluster.KubeconfigFile, mc); err != nil {
			return err
		}
	}
	return p.kubeUnAuthClient.Delete(ctx, clusterSpec.SnowDatacenter.GetName(), clusterSpec.SnowDatacenter.GetNamespace(), clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.SnowDatacenter)
}

func (p *SnowProvider) PostClusterDeleteValidate(_ context.Context, _ *types.Cluster) error {
	// No validations
	return nil
}

func (p *SnowProvider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	return nil
}

func (p *SnowProvider) PostClusterDeleteForUpgrade(ctx context.Context, managementCluster *types.Cluster) error {
	return nil
}
