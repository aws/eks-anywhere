package nutanix

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed config/cp-template.yaml
var defaultCAPIConfigCP string

//go:embed config/md-template.yaml
var defaultClusterConfigMD string

//go:embed config/secret-template.yaml
var secretTemplate string

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

var (
	eksaNutanixDatacenterResourceType = fmt.Sprintf("nutanixdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaNutanixMachineResourceType    = fmt.Sprintf("nutanixmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	requiredEnvs                      = []string{nutanixEndpointKey, nutanixUsernameKey, nutanixPasswordKey}
)

type basicAuthCreds struct {
	username string
	password string
}

// Provider implements the Nutanix Provider
type Provider struct {
	clusterConfig    *v1alpha1.Cluster
	datacenterConfig *v1alpha1.NutanixDatacenterConfig
	machineConfigs   map[string]*v1alpha1.NutanixMachineConfig
	templateBuilder  *TemplateBuilder
	kubectlClient    ProviderKubectlClient
	validator        *Validator
}

var _ providers.Provider = &Provider{}

// NewProvider returns a new nutanix provider
func NewProvider(
	datacenterConfig *v1alpha1.NutanixDatacenterConfig,
	machineConfigs map[string]*v1alpha1.NutanixMachineConfig,
	clusterConfig *v1alpha1.Cluster,
	providerKubectlClient ProviderKubectlClient,
	now types.NowFunc,
) (*Provider, error) {
	var controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.NutanixMachineConfigSpec
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}

	creds := getCredsFromEnv()
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.NutanixMachineConfigSpec, len(machineConfigs))
	templateBuilder := NewNutanixTemplateBuilder(&datacenterConfig.Spec, controlPlaneMachineSpec, etcdMachineSpec, workerNodeGroupMachineSpecs, creds, now)

	url := fmt.Sprintf("%s:%d", datacenterConfig.Spec.Endpoint, datacenterConfig.Spec.Port)
	nutanixCreds := prismgoclient.Credentials{
		URL:      url,
		Username: creds.username,
		Password: creds.password,
		Endpoint: datacenterConfig.Spec.Endpoint,
		Port:     fmt.Sprintf("%d", datacenterConfig.Spec.Port),
	}
	client, err := v3.NewV3Client(nutanixCreds)
	if err != nil {
		return nil, fmt.Errorf("error creating nutanix client: %v", err)
	}

	validator, err := NewValidator(client.V3)
	if err != nil {
		return nil, err
	}

	return &Provider{
		clusterConfig:    clusterConfig,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		templateBuilder:  templateBuilder,
		kubectlClient:    providerKubectlClient,
		validator:        validator,
	}, nil
}

func (p *Provider) BootstrapClusterOpts(_ *cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error) {
	// TODO(nutanix): figure out if we need something else here
	return nil, nil
}

func (p *Provider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) PostBootstrapDeleteForUpgrade(ctx context.Context) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) Name() string {
	return constants.NutanixProviderName
}

func (p *Provider) DatacenterResourceType() string {
	return eksaNutanixDatacenterResourceType
}

func (p *Provider) MachineResourceType() string {
	return eksaNutanixMachineResourceType
}

func (p *Provider) DeleteResources(_ context.Context, _ *cluster.Spec) error {
	// TODO(nutanix): Add delete resource logic
	return nil
}

func (p *Provider) PostClusterDeleteValidate(ctx context.Context, managementCluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The nutanix infrastructure provider is still in development and should not be used in production")
	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	for _, conf := range clusterSpec.NutanixMachineConfigs {
		if err := p.validator.ValidateMachineConfig(ctx, conf); err != nil {
			return fmt.Errorf("failed to validate machine config: %v", err)
		}
	}

	return nil
}

func (p *Provider) SetupAndValidateDeleteCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	return nil
}

func (p *Provider) SetupAndValidateUpgradeCluster(ctx context.Context, _ *types.Cluster, _ *cluster.Spec, _ *cluster.Spec) error {
	// TODO(nutanix): Add validations when this is supported
	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	return errors.New("upgrade for nutanix provider isn't currently supported")
}

func (p *Provider) UpdateSecrets(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	contents, err := p.templateBuilder.GenerateCAPISpecSecret(clusterSpec)
	if err != nil {
		return err
	}

	if err := p.kubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, contents); err != nil {
		return fmt.Errorf("loading secrets object: %v", err)
	}
	return nil
}

func (p *Provider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := clusterSpec.Cluster.Name

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = common.CPMachineTemplateName(clusterName, p.templateBuilder.now)
		values["etcdTemplateName"] = common.EtcdMachineTemplateName(clusterName, p.templateBuilder.now)
	}
	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	if err != nil {
		return nil, nil, err
	}

	workloadTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workloadTemplateNames[workerNodeGroupConfiguration.Name] = common.WorkerMachineTemplateName(clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
		kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = common.KubeadmConfigTemplateName(clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
		p.templateBuilder.workerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}
	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *Provider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	// TODO(nutanix): implement
	return nil, nil, nil
}

func (p *Provider) GenerateStorageClass() []byte {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) GenerateMHC(clusterSpec *cluster.Spec) ([]byte, error) {
	data := map[string]string{
		"clusterName":         p.clusterConfig.Name,
		"eksaSystemNamespace": constants.EksaSystemNamespace,
	}
	mhc, err := templater.Execute(string(mhcTemplate), data)
	if err != nil {
		return nil, err
	}

	return mhc, nil
}

func (p *Provider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.Nutanix.Version
}

func (p *Provider) EnvMap(_ *cluster.Spec) (map[string]string, error) {
	// TODO(nutanix): determine if any env vars are needed and add them to requiredEnvs
	envMap := make(map[string]string)
	for _, key := range requiredEnvs {
		if env, ok := os.LookupEnv(key); ok && len(env) > 0 {
			envMap[key] = env
		} else {
			return envMap, fmt.Errorf("warning required env not set %s", key)
		}
	}
	return envMap, nil
}

func (p *Provider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capx-system": {"controller-manager"},
	}
}

func (p *Provider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	bundle := clusterSpec.VersionsBundle
	manifests := []releasev1alpha1.Manifest{
		bundle.Nutanix.Components,
		bundle.Nutanix.Metadata,
		bundle.Nutanix.ClusterTemplate,
	}
	folderName := fmt.Sprintf("infrastructure-nutanix/%s/", p.Version(clusterSpec))
	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests:  manifests,
	}
	return &infraBundle
}

func (p *Provider) DatacenterConfig(_ *cluster.Spec) providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *Provider) MachineConfigs(_ *cluster.Spec) []providers.MachineConfig {
	configs := make(map[string]providers.MachineConfig, len(p.machineConfigs))
	controlPlaneMachineName := p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	p.machineConfigs[controlPlaneMachineName].Annotations = map[string]string{p.clusterConfig.ControlPlaneAnnotation(): "true"}
	if p.clusterConfig.IsManaged() {
		p.machineConfigs[controlPlaneMachineName].SetManagedBy(p.clusterConfig.ManagedBy())
	}
	configs[controlPlaneMachineName] = p.machineConfigs[controlPlaneMachineName]

	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		p.machineConfigs[etcdMachineName].Annotations = map[string]string{p.clusterConfig.EtcdAnnotation(): "true"}
		if etcdMachineName != controlPlaneMachineName {
			configs[etcdMachineName] = p.machineConfigs[etcdMachineName]
			if p.clusterConfig.IsManaged() {
				p.machineConfigs[etcdMachineName].SetManagedBy(p.clusterConfig.ManagedBy())
			}
		}
	}

	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		workerMachineName := workerNodeGroupConfiguration.MachineGroupRef.Name
		if _, ok := configs[workerMachineName]; !ok {
			configs[workerMachineName] = p.machineConfigs[workerMachineName]
			if p.clusterConfig.IsManaged() {
				p.machineConfigs[workerMachineName].SetManagedBy(p.clusterConfig.ManagedBy())
			}
		}
	}
	return configsMapToSlice(configs)
}

func configsMapToSlice(c map[string]providers.MachineConfig) []providers.MachineConfig {
	configs := make([]providers.MachineConfig, 0, len(c))
	for _, config := range c {
		configs = append(configs, config)
	}

	return configs
}

func (p *Provider) ValidateNewSpec(_ context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) UpgradeNeeded(_ context.Context, _, _ *cluster.Spec, _ *types.Cluster) (bool, error) {
	// TODO(nutanix): figure out if we need something else here
	return false, nil
}

func (p *Provider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	nodeGroupsToDelete := cluster.NodeGroupsToDelete(currentSpec, newSpec)
	machineDeployments := make([]string, 0, len(nodeGroupsToDelete))
	for _, nodeGroup := range nodeGroupsToDelete {
		mdName := machineDeploymentName(workloadCluster.Name, nodeGroup.Name)
		machineDeployments = append(machineDeployments, mdName)
	}
	return machineDeployments
}

func (p *Provider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	return p.kubectlClient.SetEksaControllerEnvVar(ctx, features.NutanixProviderEnvVar, "true", kubeconfigFile)
}

func (p *Provider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) PostMoveManagementToBootstrap(ctx context.Context, bootstrapCluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) Validator() *Validator {
	// TODO(nutanix): figure out if we need something else here
	return p.validator
}
