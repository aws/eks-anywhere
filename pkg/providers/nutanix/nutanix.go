package nutanix

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"

	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	nutanixUsernameKey = "NUTANIX_USER"
	nutanixPasswordKey = "NUTANIX_PASSWORD"
	nutanixEndpointKey = "NUTANIX_ENDPOINT"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

var (
	eksaNutanixDatacenterResourceType = fmt.Sprintf("nutanixdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaNutanixMachineResourceType    = fmt.Sprintf("nutanixmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	requiredEnvs                      = []string{nutanixEndpointKey, nutanixUsernameKey, nutanixPasswordKey}
)

type nutanixProvider struct {
	clusterConfig    *v1alpha1.Cluster
	datacenterConfig *v1alpha1.NutanixDatacenterConfig
	machineConfigs   map[string]*v1alpha1.NutanixMachineConfig
	// providerKubectlClient ProviderKubectlClient
	templateBuilder *NutanixTemplateBuilder
}

// type ProviderKubectlClient interface {
// 	// TODO: Add necessary kubectl functions here
// 	foo() error
// }

func NewProvider(
	datacenterConfig *v1alpha1.NutanixDatacenterConfig,
	machineConfigs map[string]*v1alpha1.NutanixMachineConfig,
	clusterConfig *v1alpha1.Cluster,
	now types.NowFunc,
) *nutanixProvider {
	var controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.NutanixMachineConfigSpec
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.NutanixMachineConfigSpec, len(machineConfigs))

	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}
	return &nutanixProvider{
		clusterConfig:    clusterConfig,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		templateBuilder: &NutanixTemplateBuilder{
			datacenterSpec:              &datacenterConfig.Spec,
			controlPlaneMachineSpec:     controlPlaneMachineSpec,
			etcdMachineSpec:             etcdMachineSpec,
			workerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
			now:                         now,
		},
	}
}
func (p *nutanixProvider) BootstrapClusterOpts(_ *cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error) {
	return nil, nil
}

func (p *nutanixProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO: figure out if we need something else here
	return nil
}

func (p *nutanixProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO: figure out if we need something else here
	return nil
}

func (p *nutanixProvider) PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO: figure out if we need something else here
	return nil
}

func (p *nutanixProvider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *nutanixProvider) Name() string {
	return constants.NutanixProviderName
}

func (p *nutanixProvider) DatacenterResourceType() string {
	return eksaNutanixDatacenterResourceType
}

func (p *nutanixProvider) MachineResourceType() string {
	return eksaNutanixMachineResourceType
}

func (p *nutanixProvider) DeleteResources(_ context.Context, _ *cluster.Spec) error {
	// TODO: Add delete resource logic
	return nil
}

func (p *nutanixProvider) PostClusterDeleteValidate(ctx context.Context, managementCluster *types.Cluster) error {
	// TODO:
	return nil
}

func (p *nutanixProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The nutanix infrastructure provider is still in development and should not be used in production")
	if err := SetupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	// TODO
	return nil
}

func (p *nutanixProvider) SetupAndValidateDeleteCluster(ctx context.Context, cluster *types.Cluster) error {
	if err := SetupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	// TODO
	return nil
}

func (p *nutanixProvider) SetupAndValidateUpgradeCluster(ctx context.Context, _ *types.Cluster, _ *cluster.Spec, _ *cluster.Spec) error {
	// TODO: Add validations when this is supported
	if err := SetupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	// TODO
	return errors.New("upgrade for nutanix provider isn't currently supported")
}

func (p *nutanixProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	// TODO: implement
	return nil
}

type NutanixTemplateBuilder struct {
	datacenterSpec              *v1alpha1.NutanixDatacenterConfigSpec
	controlPlaneMachineSpec     *v1alpha1.NutanixMachineConfigSpec
	etcdMachineSpec             *v1alpha1.NutanixMachineConfigSpec
	workerNodeGroupMachineSpecs map[string]v1alpha1.NutanixMachineConfigSpec
	now                         types.NowFunc
}

func NewNutanixTemplateBuilder(datacenterSpec *v1alpha1.NutanixDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.NutanixMachineConfigSpec, workerNodeGroupMachineSpecs map[string]v1alpha1.NutanixMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &NutanixTemplateBuilder{
		datacenterSpec:              datacenterSpec,
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		etcdMachineSpec:             etcdMachineSpec,
		workerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		now:                         now,
	}
}

func (ntb *NutanixTemplateBuilder) WorkerMachineTemplateName(clusterName string) string {
	t := ntb.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-worker-node-template-%d", clusterName, t)
}

func (ntb *NutanixTemplateBuilder) CPMachineTemplateName(clusterName string) string {
	t := ntb.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-control-plane-template-%d", clusterName, t)
}

func (ntb *NutanixTemplateBuilder) EtcdMachineTemplateName(clusterName string) string {
	t := ntb.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-etcd-template-%d", clusterName, t)
}

func (ntb *NutanixTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	var etcdMachineSpec v1alpha1.NutanixMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *ntb.etcdMachineSpec
	}
	values := buildTemplateMapCP(ntb.datacenterSpec, clusterSpec, *ntb.controlPlaneMachineSpec, etcdMachineSpec)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (ntb *NutanixTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		values := buildTemplateMapMD(clusterSpec, ntb.workerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration)
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]

		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

func (p *nutanixProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
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

func (p *nutanixProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	// TODO: implement
	return nil, nil, nil
}

func (p *nutanixProvider) GenerateStorageClass() []byte {
	// TODO: determine if we need something else here
	return nil
}

func (p *nutanixProvider) GenerateMHC(clusterSpec *cluster.Spec) ([]byte, error) {
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

func (p *nutanixProvider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *nutanixProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.Nutanix.Version
}

func (p *nutanixProvider) EnvMap(_ *cluster.Spec) (map[string]string, error) {
	// TODO: determine if any env vars are needed and add them to requiredEnvs
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

func (p *nutanixProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capx-system": {"controller-manager"},
	}
}

func (p *nutanixProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	bundle := clusterSpec.VersionsBundle
	folderName := fmt.Sprintf("infrastructure-nutanix/%s/", bundle.Nutanix.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			bundle.Nutanix.Components,
			bundle.Nutanix.Metadata,
			bundle.Nutanix.ClusterTemplate,
		},
	}
	return &infraBundle
}

func (p *nutanixProvider) DatacenterConfig(_ *cluster.Spec) providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *nutanixProvider) MachineConfigs(_ *cluster.Spec) []providers.MachineConfig {
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

func (p *nutanixProvider) ValidateNewSpec(_ context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *nutanixProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	// TODO: implement
	return nil
}

func (p *nutanixProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *nutanixProvider) UpgradeNeeded(_ context.Context, _, _ *cluster.Spec, _ *types.Cluster) (bool, error) {
	// TODO: Figure out if something is needed here
	return false, nil
}

func (p *nutanixProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	// TODO: Figure out if something is needed here
	return nil
}

func machineDeploymentName(clusterName, nodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, nodeGroupName)
}

func buildTemplateMapCP(datacenterSpec *v1alpha1.NutanixDatacenterConfigSpec, clusterSpec *cluster.Spec, controlPlaneMachineSpec v1alpha1.NutanixMachineConfigSpec, etcdMachineSpec v1alpha1.NutanixMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":                  clusterSpec.Cluster.Name,
		"controlPlaneEndpointIp":       clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":         clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"controlPlaneSshAuthorizedKey": controlPlaneMachineSpec.Users[0].SshAuthorizedKeys[0],
		"controlPlaneSshUsername":      controlPlaneMachineSpec.Users[0].Name,
		"eksaSystemNamespace":          constants.EksaSystemNamespace,
		"format":                       format,
		"podCidrs":                     clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                 clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"kubernetesVersion":            bundle.KubeDistro.Kubernetes.Tag,
		"kubernetesRepository":         bundle.KubeDistro.Kubernetes.Repository,
		"corednsRepository":            bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":               bundle.KubeDistro.CoreDNS.Tag,
		"etcdRepository":               bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                 bundle.KubeDistro.Etcd.Tag,
		"kubeVipImage":                 "ghcr.io/kube-vip/kube-vip:latest",
		"externalEtcdVersion":          bundle.KubeDistro.EtcdVersion,
		"etcdCipherSuites":             crypto.SecureCipherSuitesString(),
		"nutanixEndpoint":              datacenterSpec.NutanixEndpoint,
		"nutanixPort":                  datacenterSpec.NutanixPort,
		"nutanixInsecure":              datacenterSpec.NutanixInsecure,
		"nutanixUser":                  datacenterSpec.NutanixUser,
		"nutanixPassword":              datacenterSpec.NutanixPassword,
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
	}

	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.NutanixMachineConfigSpec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":            clusterSpec.Cluster.Name,
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
		"format":                 format,
		"kubernetesVersion":      bundle.KubeDistro.Kubernetes.Tag,
		"workerReplicas":         workerNodeGroupConfiguration.Count,
		"workerPoolName":         "md-0",
		"workerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys[0],
		"workerSshUsername":      workerNodeGroupMachineSpec.Users[0].Name,
		"vcpusPerSocket":         workerNodeGroupMachineSpec.VCPUsPerSocket,
		"vcpuSockets":            workerNodeGroupMachineSpec.VCPUSockets,
		"memorySize":             workerNodeGroupMachineSpec.MemorySize.String(),
		"systemDiskSize":         workerNodeGroupMachineSpec.SystemDiskSize.String(),
		"imageName":              workerNodeGroupMachineSpec.Image.Name,   //TODO pass name or uuid based on type of identifier
		"nutanixPEClusterName":   workerNodeGroupMachineSpec.Cluster.Name, //TODO pass name or uuid based on type of identifier
		"subnetName":             workerNodeGroupMachineSpec.Subnet.Name,  //TODO pass name or uuid based on type of identifier
	}
	return values
}

func (p *nutanixProvider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	nodeGroupsToDelete := cluster.NodeGroupsToDelete(currentSpec, newSpec)
	machineDeployments := make([]string, 0, len(nodeGroupsToDelete))
	for _, nodeGroup := range nodeGroupsToDelete {
		mdName := machineDeploymentName(workloadCluster.Name, nodeGroup.Name)
		machineDeployments = append(machineDeployments, mdName)
	}
	return machineDeployments
}

func (p *nutanixProvider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	return nil
}

func (p *nutanixProvider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}
