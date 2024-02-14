package nutanix

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
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
	// list of env variables required by CAPX to be present and defined beforehand.
	requiredEnvs = []string{nutanixEndpointKey, constants.NutanixUsernameKey, constants.NutanixPasswordKey, expClusterResourceSetKey}
)

// Provider implements the Nutanix Provider.
type Provider struct {
	clusterConfig    *v1alpha1.Cluster
	datacenterConfig *v1alpha1.NutanixDatacenterConfig
	machineConfigs   map[string]*v1alpha1.NutanixMachineConfig
	templateBuilder  *TemplateBuilder
	kubectlClient    ProviderKubectlClient
	validator        *Validator
	writer           filewriter.FileWriter
	ipValidator      IPValidator
	skipIPCheck      bool
}

var _ providers.Provider = &Provider{}

// NewProvider returns a new nutanix provider.
func NewProvider(
	datacenterConfig *v1alpha1.NutanixDatacenterConfig,
	machineConfigs map[string]*v1alpha1.NutanixMachineConfig,
	clusterConfig *v1alpha1.Cluster,
	providerKubectlClient ProviderKubectlClient,
	writer filewriter.FileWriter,
	clientCache *ClientCache,
	ipValidator IPValidator,
	certValidator crypto.TlsValidator,
	httpClient *http.Client,
	now types.NowFunc,
	skipIPCheck bool,
) *Provider {
	datacenterConfig.SetDefaults()
	for _, machineConfig := range machineConfigs {
		machineConfig.SetDefaults()
	}

	var controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.NutanixMachineConfigSpec
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}

	creds := GetCredsFromEnv()
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.NutanixMachineConfigSpec, len(machineConfigs))
	templateBuilder := NewNutanixTemplateBuilder(&datacenterConfig.Spec, controlPlaneMachineSpec, etcdMachineSpec, workerNodeGroupMachineSpecs, creds, now)

	nutanixValidator := NewValidator(clientCache, certValidator, httpClient)
	return &Provider{
		clusterConfig:    clusterConfig,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		templateBuilder:  templateBuilder,
		kubectlClient:    providerKubectlClient,
		validator:        nutanixValidator,
		writer:           writer,
		ipValidator:      ipValidator,
		skipIPCheck:      skipIPCheck,
	}
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

// PostBootstrapDeleteForUpgrade runs any provider-specific operations after bootstrap cluster has been deleted.
func (p *Provider) PostBootstrapDeleteForUpgrade(ctx context.Context, cluster *types.Cluster) error {
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

func (p *Provider) generateSSHKeysIfNotSet() error {
	var generatedKey string
	for _, machineConfig := range p.machineConfigs {
		user := machineConfig.Spec.Users[0]
		if user.SshAuthorizedKeys[0] == "" {
			if generatedKey != "" { // use the same key
				user.SshAuthorizedKeys[0] = generatedKey
			} else {
				logger.Info("Provided sshAuthorizedKey is not set or is empty, auto-generating new key pair...", "NutanixMachineConfig", machineConfig.Name)
				var err error
				generatedKey, err = common.GenerateSSHAuthKey(p.writer)
				if err != nil {
					return err
				}
				user.SshAuthorizedKeys[0] = generatedKey
			}
		}
	}

	return nil
}

func (p *Provider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range p.machineConfigs {
		if err := p.kubectlClient.DeleteEksaNutanixMachineConfig(ctx, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.kubectlClient.DeleteEksaNutanixDatacenterConfig(ctx, clusterSpec.NutanixDatacenter.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.NutanixDatacenter.Namespace)
}

func (p *Provider) PostClusterDeleteValidate(ctx context.Context, managementCluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := p.validator.validateUpgradeRolloutStrategy(clusterSpec); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	if err := setupEnvVars(clusterSpec.NutanixDatacenter); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	creds := GetCredsFromEnv()
	if err := p.validator.ValidateClusterSpec(ctx, clusterSpec, creds); err != nil {
		return fmt.Errorf("failed to validate cluster spec: %v", err)
	}

	if err := p.generateSSHKeysIfNotSet(); err != nil {
		return fmt.Errorf("failed to generate ssh key: %v", err)
	}
	clusterSpec.NutanixMachineConfigs = p.machineConfigs

	if !p.skipIPCheck {
		if err := p.ipValidator.ValidateControlPlaneIPUniqueness(clusterSpec.Cluster); err != nil {
			return err
		}
	} else {
		logger.Info("Skipping check for whether control plane ip is in use")
	}

	return nil
}

func (p *Provider) SetupAndValidateDeleteCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if err := p.validator.validateUpgradeRolloutStrategy(clusterSpec); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	return nil
}

// SetupAndValidateUpgradeCluster - Performs necessary setup and validations for upgrade cluster operation.
func (p *Provider) SetupAndValidateUpgradeCluster(ctx context.Context, _ *types.Cluster, clusterSpec *cluster.Spec, _ *cluster.Spec) error {
	if err := p.SetupAndValidateUpgradeManagementComponents(ctx, clusterSpec); err != nil {
		return err
	}

	if err := p.validator.validateUpgradeRolloutStrategy(clusterSpec); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	return nil
}

// SetupAndValidateUpgradeManagementComponents performs necessary setup for upgrade management components operation.
func (p *Provider) SetupAndValidateUpgradeManagementComponents(_ context.Context, _ *cluster.Spec) error {
	// TODO(nutanix): Add validations when this is supported
	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	return nil
}

func (p *Provider) UpdateSecrets(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	// check if CAPI Secret name and EKS-A Secret name are not the same
	// this is to ensure that the EKS-A Secret that is watched and CAPX Secret that is reconciled are not the same
	if CAPXSecretName(clusterSpec) == EKSASecretName(clusterSpec) {
		return fmt.Errorf("NutanixDatacenterConfig CredentialRef name cannot be the same as the NutanixCluster CredentialRef name")
	}

	capxSecretContents, err := p.templateBuilder.GenerateCAPISpecSecret(clusterSpec)
	if err != nil {
		return err
	}

	if err := p.kubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, capxSecretContents); err != nil {
		return fmt.Errorf("loading secrets object: %v", err)
	}

	return p.updateEKSASecrets(ctx, cluster, clusterSpec)
}

// updateEKSASecrets generates and applies the EKSA secret on the cluster.
func (p *Provider) updateEKSASecrets(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	eksaSecretContents, err := p.templateBuilder.GenerateEKSASpecSecret(clusterSpec)
	if err != nil {
		return err
	}

	if err := p.kubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, eksaSecretContents); err != nil {
		return fmt.Errorf("loading secrets object: %v", err)
	}

	return nil
}

func (p *Provider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	if err := p.UpdateSecrets(ctx, cluster, clusterSpec); err != nil {
		return nil, nil, fmt.Errorf("updating Nutanix credentials: %v", err)
	}

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

func NeedsNewControlPlaneTemplate(oldSpec, newSpec *cluster.Spec, oldNmc, newNmc *v1alpha1.NutanixMachineConfig) bool {
	// Another option is to generate MachineTemplates based on the old and new eksa spec,
	// remove the name field and compare them with DeepEqual
	// We plan to approach this way since it's more flexible to add/remove fields and test out for validation
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldNmc, newNmc)
}

func nutanixIdentifierChanged(old, new v1alpha1.NutanixResourceIdentifier) bool {
	if old.Type != new.Type {
		return true
	}
	if old.Type == v1alpha1.NutanixIdentifierName && old.Name != nil && new.Name != nil && *old.Name != *new.Name {
		return true
	}
	if old.Type == v1alpha1.NutanixIdentifierUUID && old.UUID != nil && new.UUID != nil && *old.UUID != *new.UUID {
		return true
	}
	return false
}

func AnyImmutableFieldChanged(oldNmc, newNmc *v1alpha1.NutanixMachineConfig) bool {
	if oldNmc.Spec.MemorySize != newNmc.Spec.MemorySize {
		return true
	}
	if oldNmc.Spec.SystemDiskSize != newNmc.Spec.SystemDiskSize {
		return true
	}
	if oldNmc.Spec.VCPUSockets != newNmc.Spec.VCPUSockets {
		return true
	}
	if oldNmc.Spec.VCPUsPerSocket != newNmc.Spec.VCPUsPerSocket {
		return true
	}
	if oldNmc.Spec.OSFamily != newNmc.Spec.OSFamily {
		return true
	}
	if nutanixIdentifierChanged(oldNmc.Spec.Image, newNmc.Spec.Image) {
		return true
	}
	if nutanixIdentifierChanged(oldNmc.Spec.Cluster, newNmc.Spec.Cluster) {
		return true
	}
	if nutanixIdentifierChanged(oldNmc.Spec.Subnet, newNmc.Spec.Subnet) {
		return true
	}

	return false
}

func (p *Provider) getWorkerNodeMachineConfigs(ctx context.Context, workloadCluster *types.Cluster, newClusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (*v1alpha1.NutanixMachineConfig, *v1alpha1.NutanixMachineConfig, error) {
	if _, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		oldWorkerMachineConfig := p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		newWorkerMachineConfig, err := p.kubectlClient.GetEksaNutanixMachineConfig(ctx, workerNodeGroupConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		if err != nil {
			return oldWorkerMachineConfig, nil, err
		}
		return oldWorkerMachineConfig, newWorkerMachineConfig, nil
	}
	return nil, nil, nil
}

func (p *Provider) needsNewMachineTemplate(currentSpec, newClusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, ndc *v1alpha1.NutanixDatacenterConfig, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration, oldWorkerMachineConfig *v1alpha1.NutanixMachineConfig, newWorkerMachineConfig *v1alpha1.NutanixMachineConfig) (bool, error) {
	if prevWorkerGroup, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		needsNewWorkloadTemplate := NeedsNewWorkloadTemplate(currentSpec, newClusterSpec, oldWorkerMachineConfig, newWorkerMachineConfig, prevWorkerGroup, workerNodeGroupConfiguration)
		return needsNewWorkloadTemplate, nil
	}
	return true, nil
}

func (p *Provider) needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration, oldWorkerNodeNmc *v1alpha1.NutanixMachineConfig, newWorkerNodeNmc *v1alpha1.NutanixMachineConfig) (bool, error) {
	if _, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		existingWorkerNodeGroupConfig := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]
		return NeedsNewKubeadmConfigTemplate(&workerNodeGroupConfiguration, &existingWorkerNodeGroupConfig, oldWorkerNodeNmc, newWorkerNodeNmc), nil
	}
	return true, nil
}

// NeedsNewWorkloadTemplate determines if a new workload template is needed.
func NeedsNewWorkloadTemplate(oldSpec, newSpec *cluster.Spec, oldNmc, newNmc *v1alpha1.NutanixMachineConfig, oldWorker, newWorker v1alpha1.WorkerNodeGroupConfiguration) bool {
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	if !v1alpha1.TaintsSliceEqual(oldWorker.Taints, newWorker.Taints) ||
		!v1alpha1.MapEqual(oldWorker.Labels, newWorker.Labels) ||
		!v1alpha1.WorkerNodeGroupConfigurationKubeVersionUnchanged(&oldWorker, &newWorker, oldSpec.Cluster, newSpec.Cluster) {
		return true
	}
	return AnyImmutableFieldChanged(oldNmc, newNmc)
}

func NeedsNewKubeadmConfigTemplate(newWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration, oldWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration, oldWorkerNodeNmc *v1alpha1.NutanixMachineConfig, newWorkerNodeNmc *v1alpha1.NutanixMachineConfig) bool {
	return !v1alpha1.TaintsSliceEqual(newWorkerNodeGroup.Taints, oldWorkerNodeGroup.Taints) || !v1alpha1.MapEqual(newWorkerNodeGroup.Labels, oldWorkerNodeGroup.Labels) ||
		!v1alpha1.UsersSliceEqual(oldWorkerNodeNmc.Spec.Users, newWorkerNodeNmc.Spec.Users)
}

func (p *Provider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	if err := p.UpdateSecrets(ctx, bootstrapCluster, newClusterSpec); err != nil {
		return nil, nil, fmt.Errorf("updating Nutanix credentials: %v", err)
	}

	clusterName := newClusterSpec.Cluster.Name
	var controlPlaneTemplateName, workloadTemplateName, kubeadmconfigTemplateName, etcdTemplateName string

	// Get existing EKSA Cluster
	eksaCluster, err := p.kubectlClient.GetEksaCluster(ctx, workloadCluster, newClusterSpec.Cluster.Name)
	if err != nil {
		return nil, nil, err
	}

	// Get current Nutanix Datacenter Config
	ndc, err := p.kubectlClient.GetEksaNutanixDatacenterConfig(ctx, p.datacenterConfig.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return nil, nil, err
	}

	// Get current Nutanix Machine Config
	controlPlaneMachineConfig := p.machineConfigs[newClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	controlPlaneNutanixMachineConfig, err := p.kubectlClient.GetEksaNutanixMachineConfig(ctx,
		eksaCluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name,
		workloadCluster.KubeconfigFile,
		newClusterSpec.Cluster.Namespace)
	if err != nil {
		return nil, nil, err
	}
	needsNewControlPlaneTemplate := NeedsNewControlPlaneTemplate(currentSpec, newClusterSpec, controlPlaneNutanixMachineConfig, controlPlaneMachineConfig)
	if !needsNewControlPlaneTemplate {
		cp, err := p.kubectlClient.GetKubeadmControlPlane(ctx, workloadCluster, eksaCluster.Name, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return nil, nil, err
		}
		controlPlaneTemplateName = cp.Spec.MachineTemplate.InfrastructureRef.Name
	} else {
		controlPlaneTemplateName = common.CPMachineTemplateName(clusterName, p.templateBuilder.now)
	}

	previousWorkerNodeGroupConfigs := cluster.BuildMapForWorkerNodeGroupsByName(currentSpec.Cluster.Spec.WorkerNodeGroupConfigurations)

	workloadTemplateNames := make(map[string]string, len(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {

		oldWorkerNodeNmc, newWorkerNodeNmc, err := p.getWorkerNodeMachineConfigs(ctx, workloadCluster, newClusterSpec, workerNodeGroupConfiguration, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, nil, err
		}
		needsNewWorkloadTemplate, err := p.needsNewMachineTemplate(currentSpec, newClusterSpec, workerNodeGroupConfiguration, ndc, previousWorkerNodeGroupConfigs, oldWorkerNodeNmc, newWorkerNodeNmc)
		if err != nil {
			return nil, nil, err
		}
		needsNewKubeadmConfigTemplate, err := p.needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration, previousWorkerNodeGroupConfigs, oldWorkerNodeNmc, newWorkerNodeNmc)
		if err != nil {
			return nil, nil, err
		}
		if !needsNewKubeadmConfigTemplate {
			mdName := machineDeploymentName(newClusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name)
			md, err := p.kubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			kubeadmconfigTemplateName = md.Spec.Template.Spec.Bootstrap.ConfigRef.Name
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = kubeadmconfigTemplateName
		} else {
			kubeadmconfigTemplateName = common.KubeadmConfigTemplateName(clusterName, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = kubeadmconfigTemplateName
		}

		if !needsNewWorkloadTemplate {
			mdName := machineDeploymentName(newClusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name)
			md, err := p.kubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			workloadTemplateName = md.Spec.Template.Spec.InfrastructureRef.Name
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		} else {
			workloadTemplateName = common.WorkerMachineTemplateName(clusterName, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		}
		p.templateBuilder.workerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["etcdTemplateName"] = etcdTemplateName
	}
	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(newClusterSpec, cpOpt)
	if err != nil {
		return nil, nil, err
	}

	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(newClusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *Provider) GenerateStorageClass() []byte {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

// GenerateMHC returns MachineHealthCheck for the cluster in yaml format.
func (p *Provider) GenerateMHC(_ *cluster.Spec) ([]byte, error) {
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

// Version returns the version of the provider.
func (p *Provider) Version(components *cluster.ManagementComponents) string {
	return components.Nutanix.Version
}

// EnvMap returns the environment variables for the provider.
func (p *Provider) EnvMap(_ *cluster.ManagementComponents, _ *cluster.Spec) (map[string]string, error) {
	// TODO(nutanix): determine if any env vars are needed and add them to requiredEnvs
	envMap := make(map[string]string)
	for _, key := range requiredEnvs {
		if env, ok := os.LookupEnv(key); ok && len(env) > 0 {
			envMap[key] = env
		} else {
			return nil, fmt.Errorf("required env not set %s", key)
		}
	}
	return envMap, nil
}

func (p *Provider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capx-system": {"capx-controller-manager"},
	}
}

// GetInfrastructureBundle returns the infrastructure bundle for the provider.
func (p *Provider) GetInfrastructureBundle(components *cluster.ManagementComponents) *types.InfrastructureBundle {
	manifests := []releasev1alpha1.Manifest{
		components.Nutanix.Components,
		components.Nutanix.Metadata,
		components.Nutanix.ClusterTemplate,
	}
	folderName := fmt.Sprintf("infrastructure-nutanix/%s/", components.Nutanix.Version)
	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests:  manifests,
	}
	return &infraBundle
}

func (p *Provider) DatacenterConfig(_ *cluster.Spec) providers.DatacenterConfig {
	return p.datacenterConfig
}

// MachineConfigs returns a MachineConfig slice.
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

// ChangeDiff returns the component change diff for the provider.
func (p *Provider) ChangeDiff(currentComponents, newComponents *cluster.ManagementComponents) *types.ComponentChangeDiff {
	if currentComponents.Nutanix.Version == newComponents.Nutanix.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.NutanixProviderName,
		NewVersion:    newComponents.Nutanix.Version,
		OldVersion:    currentComponents.Nutanix.Version,
	}
}

func (p *Provider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

func (p *Provider) UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec, cluster *types.Cluster) (bool, error) {
	cc := currentSpec.Cluster
	existingVdc, err := p.kubectlClient.GetEksaNutanixDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, newSpec.Cluster.Namespace)
	if err != nil {
		return false, err
	}
	if !reflect.DeepEqual(existingVdc.Spec, p.datacenterConfig.Spec) {
		logger.V(3).Info("New provider spec is different from the new spec")
		return true, nil
	}

	machineConfigsSpecChanged, err := p.machineConfigsSpecChanged(ctx, cc, cluster, newSpec)
	if err != nil {
		return false, err
	}
	return machineConfigsSpecChanged, nil
}

func (p *Provider) machineConfigsSpecChanged(ctx context.Context, cc *v1alpha1.Cluster, cluster *types.Cluster, newClusterSpec *cluster.Spec) (bool, error) {
	machineConfigMap := make(map[string]*v1alpha1.NutanixMachineConfig)
	for _, config := range p.MachineConfigs(nil) {
		mc := config.(*v1alpha1.NutanixMachineConfig)
		machineConfigMap[mc.Name] = mc
	}

	for _, oldMcRef := range cc.MachineConfigRefs() {
		existingVmc, err := p.kubectlClient.GetEksaNutanixMachineConfig(ctx, oldMcRef.Name, cluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		if err != nil {
			return false, err
		}
		csmc, ok := machineConfigMap[oldMcRef.Name]
		if !ok {
			logger.V(3).Info(fmt.Sprintf("Old machine config spec %s not found in the existing spec", oldMcRef.Name))
			return true, nil
		}
		if !reflect.DeepEqual(existingVmc.Spec, csmc.Spec) {
			logger.V(3).Info(fmt.Sprintf("New machine config spec %s is different from the existing spec", oldMcRef.Name))
			return true, nil
		}
	}

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
	return nil
}

func (p *Provider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return p.updateEKSASecrets(ctx, cluster, clusterSpec)
}

func (p *Provider) PostMoveManagementToBootstrap(ctx context.Context, bootstrapCluster *types.Cluster) error {
	// TODO(nutanix): figure out if we need something else here
	return nil
}

// PreCoreComponentsUpgrade staisfies the Provider interface.
func (p *Provider) PreCoreComponentsUpgrade(
	ctx context.Context,
	cluster *types.Cluster,
	managementComponents *cluster.ManagementComponents,
	clusterSpec *cluster.Spec,
) error {
	return nil
}
