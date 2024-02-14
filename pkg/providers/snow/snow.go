package snow

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
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
	configManager    *ConfigManager
	ipValidator      *providerValidator.IPValidator
	skipIpCheck      bool
	log              logr.Logger
}

type KubeUnAuthClient interface {
	KubeconfigClient(kubeconfig string) kubernetes.Client
	Apply(ctx context.Context, kubeconfig string, obj runtime.Object) error
}

func NewProvider(kubeUnAuthClient KubeUnAuthClient, configManager *ConfigManager, skipIpCheck bool) *SnowProvider {
	retrier := retrier.NewWithMaxRetries(maxRetries, backOffPeriod)
	return &SnowProvider{
		kubeUnAuthClient: kubeUnAuthClient,
		retrier:          retrier,
		configManager:    configManager,
		ipValidator:      providerValidator.NewIPValidator(),
		skipIpCheck:      skipIpCheck,
		log:              logger.Get(),
	}
}

func (p *SnowProvider) Name() string {
	return constants.SnowProviderName
}

func (p *SnowProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := p.configManager.SetDefaultsAndValidate(ctx, clusterSpec.Config); err != nil {
		return fmt.Errorf("setting defaults and validate snow config: %v", err)
	}
	if !p.skipIpCheck {
		if err := p.ipValidator.ValidateControlPlaneIPUniqueness(clusterSpec.Cluster); err != nil {
			return err
		}
	} else {
		logger.Info("Skipping check for whether control plane ip is in use")
	}
	return nil
}

func (p *SnowProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, _ *cluster.Spec) error {
	if err := p.configManager.SetDefaultsAndValidate(ctx, clusterSpec.Config); err != nil {
		return fmt.Errorf("setting defaults and validate snow config: %v", err)
	}
	return nil
}

// SetupAndValidateUpgradeManagementComponents performs necessary setup for upgrade management components operation.
func (p *SnowProvider) SetupAndValidateUpgradeManagementComponents(_ context.Context, _ *cluster.Spec) error {
	return nil
}

func (p *SnowProvider) SetupAndValidateDeleteCluster(ctx context.Context, _ *types.Cluster, clusterSpec *cluster.Spec) error {
	if err := SetupEksaCredentialsSecret(clusterSpec.Config); err != nil {
		return fmt.Errorf("setting up credentials: %v", err)
	}
	return nil
}

func (p *SnowProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if err := p.kubeUnAuthClient.Apply(ctx, cluster.KubeconfigFile, clusterSpec.SnowCredentialsSecret); err != nil {
		return fmt.Errorf("applying eks-a snow credentials secret in cluster: %v", err)
	}
	return nil
}

// CAPIObjects generates the control plane and worker nodes objects for snow provider from clusterSpec.
func CAPIObjects(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneObjs, err := ControlPlaneObjects(ctx, log, clusterSpec, kubeClient)
	if err != nil {
		return nil, nil, err
	}

	controlPlaneSpec, err = templater.ObjectsToYaml(kubernetesToRuntimeObjects(controlPlaneObjs)...)
	if err != nil {
		return nil, nil, err
	}

	workersObjs, err := WorkersObjects(ctx, log, clusterSpec, kubeClient)
	if err != nil {
		return nil, nil, err
	}

	workersSpec, err = templater.ObjectsToYaml(kubernetesToRuntimeObjects(workersObjs)...)
	if err != nil {
		return nil, nil, err
	}

	return controlPlaneSpec, workersSpec, nil
}

func kubernetesToRuntimeObjects(objs []kubernetes.Object) []runtime.Object {
	runtimeObjs := make([]runtime.Object, 0, len(objs))
	for _, o := range objs {
		runtimeObjs = append(runtimeObjs, o)
	}

	return runtimeObjs
}

func (p *SnowProvider) generateCAPISpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	kubeconfigClient := p.kubeUnAuthClient.KubeconfigClient(cluster.KubeconfigFile)
	return CAPIObjects(ctx, p.log, clusterSpec, kubeconfigClient)
}

func (p *SnowProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return p.generateCAPISpec(ctx, cluster, clusterSpec)
}

func (p *SnowProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, _ *types.Cluster, _ *cluster.Spec, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return p.generateCAPISpec(ctx, bootstrapCluster, clusterSpec)
}

func (p *SnowProvider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *SnowProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

// PostBootstrapDeleteForUpgrade runs any provider-specific operations after bootstrap cluster has been deleted.
func (p *SnowProvider) PostBootstrapDeleteForUpgrade(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func (p *SnowProvider) PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *SnowProvider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *SnowProvider) BootstrapClusterOpts(_ *cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error) {
	return nil, nil
}

func (p *SnowProvider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

// Version returns the snow version from the management components.
func (p *SnowProvider) Version(components *cluster.ManagementComponents) string {
	return components.Snow.Version
}

// EnvMap returns the environment variables for the snow provider.
func (p *SnowProvider) EnvMap(managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec) (map[string]string, error) {
	envMap := make(map[string]string)
	envMap[snowCredentialsKey] = string(clusterSpec.SnowCredentialsSecret.Data[v1alpha1.SnowCredentialsKey])
	envMap[snowCertsKey] = string(clusterSpec.SnowCredentialsSecret.Data[v1alpha1.SnowCertificatesKey])

	envMap["SNOW_CONTROLLER_IMAGE"] = managementComponents.Snow.Manager.VersionedImage()

	return envMap, nil
}

func (p *SnowProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		constants.CapasSystemNamespace: {"capas-controller-manager"},
	}
}

// GetInfrastructureBundle returns the infrastructure bundle from the management components.
func (p *SnowProvider) GetInfrastructureBundle(components *cluster.ManagementComponents) *types.InfrastructureBundle {
	folderName := fmt.Sprintf("infrastructure-snow/%s/", components.Snow.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			components.Snow.Components,
			components.Snow.Metadata,
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

// ChangeDiff returns the change diff from the management components.
func (p *SnowProvider) ChangeDiff(currentComponents, newComponents *cluster.ManagementComponents) *types.ComponentChangeDiff {
	if currentComponents.Snow.Version == newComponents.Snow.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.SnowProviderName,
		NewVersion:    newComponents.Snow.Version,
		OldVersion:    currentComponents.Snow.Version,
	}
}

func (p *SnowProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	return nil
}

func bundleImagesEqual(new, old releasev1alpha1.SnowBundle) bool {
	return new.Manager.ImageDigest == old.Manager.ImageDigest && new.KubeVip.ImageDigest == old.KubeVip.ImageDigest
}

func (p *SnowProvider) machineConfigsChanged(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec) (bool, error) {
	client := p.kubeUnAuthClient.KubeconfigClient(cluster.KubeconfigFile)

	for _, new := range spec.SnowMachineConfigs {
		old := &v1alpha1.SnowMachineConfig{}
		err := client.Get(ctx, new.Name, namespaceOrDefault(new), old)
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		if len(new.Spec.Devices) != len(old.Spec.Devices) || !equality.Semantic.DeepDerivative(new.Spec, old.Spec) {
			return true, nil
		}
	}

	return false, nil
}

func (p *SnowProvider) datacenterChanged(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec) (bool, error) {
	client := p.kubeUnAuthClient.KubeconfigClient(cluster.KubeconfigFile)
	new := spec.SnowDatacenter
	old := &v1alpha1.SnowDatacenterConfig{}
	err := client.Get(ctx, new.Name, namespaceOrDefault(new), old)
	if apierrors.IsNotFound(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return !equality.Semantic.DeepDerivative(new.Spec, old.Spec), nil
}

// namespaceOrDefault return the object namespace or default if it's empty.
func namespaceOrDefault(obj client.Object) string {
	ns := obj.GetNamespace()
	if ns == "" {
		ns = "default"
	}

	return ns
}

// UpgradeNeeded compares the new snow version bundle and objects with the existing ones in the cluster and decides whether
// to trigger a cluster upgrade or not.
// TODO: revert the change once cluster.BuildSpec is used in cluster_manager to replace the deprecated cluster.BuildSpecForCluster
func (p *SnowProvider) UpgradeNeeded(ctx context.Context, newSpec, oldSpec *cluster.Spec, c *types.Cluster) (bool, error) {
	oldVersionBundle := oldSpec.RootVersionsBundle()
	newVersionsBundle := newSpec.RootVersionsBundle()
	if !bundleImagesEqual(newVersionsBundle.Snow, oldVersionBundle.Snow) {
		return true, nil
	}

	datacenterChanged, err := p.datacenterChanged(ctx, c, newSpec)
	if err != nil {
		return false, err
	}
	if datacenterChanged {
		return true, nil
	}

	return p.machineConfigsChanged(ctx, c, newSpec)
}

func (p *SnowProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	client := p.kubeUnAuthClient.KubeconfigClient(clusterSpec.ManagementCluster.KubeconfigFile)

	for _, mc := range clusterSpec.SnowMachineConfigs {
		mc.Namespace = namespaceOrDefault(mc)
		if err := client.Delete(ctx, mc); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	clusterSpec.SnowDatacenter.Namespace = namespaceOrDefault(clusterSpec.SnowDatacenter)
	if err := client.Delete(ctx, clusterSpec.SnowDatacenter); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting snow datacenter: %v", err)
	}

	return nil
}

func (p *SnowProvider) PostClusterDeleteValidate(_ context.Context, _ *types.Cluster) error {
	// No validations
	return nil
}

func (p *SnowProvider) PostMoveManagementToBootstrap(_ context.Context, _ *types.Cluster) error {
	// NOOP
	return nil
}

func (p *SnowProvider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	return nil
}

// PreCoreComponentsUpgrade staisfies the Provider interface.
func (p *SnowProvider) PreCoreComponentsUpgrade(
	ctx context.Context,
	cluster *types.Cluster,
	managementComponents *cluster.ManagementComponents,
	clusterSpec *cluster.Spec,
) error {
	return nil
}
