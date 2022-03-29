package cloudstack

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"

	etcdv1beta1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	eksaLicense                = "EKSA_LICENSE"
	controlEndpointDefaultPort = "6443"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

var requiredEnvs = []string{decoder.CloudStackCloudConfigB64SecretKey}

var (
	eksaCloudStackDatacenterResourceType = fmt.Sprintf("cloudstackdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaCloudStackMachineResourceType    = fmt.Sprintf("cloudstackmachineconfigs.%s", v1alpha1.GroupVersion.Group)
)

type cloudstackProvider struct {
	datacenterConfig       *v1alpha1.CloudStackDatacenterConfig
	machineConfigs         map[string]*v1alpha1.CloudStackMachineConfig
	clusterConfig          *v1alpha1.Cluster
	providerKubectlClient  ProviderKubectlClient
	writer                 filewriter.FileWriter
	selfSigned             bool
	controlPlaneSshAuthKey string
	workerSshAuthKey       string
	etcdSshAuthKey         string
	templateBuilder        *CloudStackTemplateBuilder
	skipIpCheck            bool
	validator              *Validator
}

func (p *cloudstackProvider) PreBootstrapSetup(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func machineRefSliceToMap(machineRefs []v1alpha1.Ref) map[string]v1alpha1.Ref {
	refMap := make(map[string]v1alpha1.Ref, len(machineRefs))
	for _, ref := range machineRefs {
		refMap[ref.Name] = ref
	}
	return refMap
}

func (p *cloudstackProvider) validateMachineConfigImmutability(ctx context.Context, cluster *types.Cluster, newConfig *v1alpha1.CloudStackMachineConfig, clusterSpec *cluster.Spec) error {
	// TODO: for GA, we need to decide which fields are immutable.
	return nil
}

func (p *cloudstackProvider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name)
	if err != nil {
		return err
	}

	prevDatacenter, err := p.providerKubectlClient.GetEksaCloudStackDatacenterConfig(ctx, prevSpec.Spec.DatacenterRef.Name, cluster.KubeconfigFile, prevSpec.Namespace)
	if err != nil {
		return err
	}

	datacenter := p.datacenterConfig

	oSpec := prevDatacenter.Spec
	nSpec := datacenter.Spec

	prevMachineConfigRefs := machineRefSliceToMap(prevSpec.MachineConfigRefs())

	for _, machineConfigRef := range clusterSpec.Cluster.MachineConfigRefs() {
		machineConfig, ok := p.machineConfigs[machineConfigRef.Name]
		if !ok {
			return fmt.Errorf("cannot find machine config %s in cloudstack provider machine configs", machineConfigRef.Name)
		}

		if _, ok = prevMachineConfigRefs[machineConfig.Name]; ok {
			err = p.validateMachineConfigImmutability(ctx, cluster, machineConfig, clusterSpec)
			if err != nil {
				return err
			}
		}
	}

	if nSpec.Domain != oSpec.Domain {
		return fmt.Errorf("spec.domain is immutable. Previous value %s, new value %s", oSpec.Domain, nSpec.Domain)
	}
	if nSpec.Account != oSpec.Account {
		return fmt.Errorf("spec.account is immutable. Previous value %s, new value %s", oSpec.Account, nSpec.Account)
	}

	if len(nSpec.Zones) != len(oSpec.Zones) {
		return fmt.Errorf("spec.zones is immutable. Previous value %s, new value %s", oSpec.Zones, nSpec.Zones)
	} else {
		for i, zone := range nSpec.Zones {
			if !zone.Equals(&oSpec.Zones[i]) {
				return fmt.Errorf("spec.zones is immutable. Previous value %s, new value %s", zone, oSpec.Zones[i])
			}
		}
	}

	return nil
}

func (p *cloudstackProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	if currentSpec.VersionsBundle.CloudStack.Version == newSpec.VersionsBundle.CloudStack.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.CloudStackProviderName,
		NewVersion:    newSpec.VersionsBundle.CloudStack.Version,
		OldVersion:    currentSpec.VersionsBundle.CloudStack.Version,
	}
}

func (p *cloudstackProvider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	nodeGroupsToDelete := cluster.NodeGroupsToDelete(currentSpec, newSpec)
	machineDeployments := make([]string, 0, len(nodeGroupsToDelete))
	for _, group := range nodeGroupsToDelete {
		mdName := machineDeploymentName(workloadCluster.Name, group.Name)
		machineDeployments = append(machineDeployments, mdName)
	}
	return machineDeployments
}

func (p *cloudstackProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// Nothing to do
	return nil
}

func (p *cloudstackProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	// Nothing to do
	return nil
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	CreateNamespace(ctx context.Context, kubeconfig string, namespace string) error
	LoadSecret(ctx context.Context, secretObject string, secretObjType string, secretObjectName string, kubeConfFile string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackDatacenterConfig, error)
	GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*kubeadmv1beta1.KubeadmControlPlane, error)
	GetMachineDeployment(ctx context.Context, workerNodeGroupName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1beta1.EtcdadmCluster, error)
	GetSecret(ctx context.Context, secretObjectName string, opts ...executables.KubectlOpt) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
	SearchCloudStackMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackMachineConfig, error)
	SearchCloudStackDatacenterConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackDatacenterConfig, error)
	DeleteEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) error
}

func NewProvider(datacenterConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerKubectlClient ProviderKubectlClient, providerCmkClient ProviderCmkClient, writer filewriter.FileWriter, now types.NowFunc, skipIpCheck bool) *cloudstackProvider {
	return NewProviderCustomNet(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		providerKubectlClient,
		providerCmkClient,
		writer,
		now,
		skipIpCheck,
	)
}

func NewProviderCustomNet(datacenterConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerKubectlClient ProviderKubectlClient, providerCmkClient ProviderCmkClient, writer filewriter.FileWriter, now types.NowFunc, skipIpCheck bool) *cloudstackProvider {
	var controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec)
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}
	if len(clusterConfig.Spec.WorkerNodeGroupConfigurations) > 0 && clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name] != nil {
		spec := machineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name].Spec
		name := clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
		workerNodeGroupMachineSpecs[name] = spec
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}
	return &cloudstackProvider{
		datacenterConfig:      datacenterConfig,
		machineConfigs:        machineConfigs,
		clusterConfig:         clusterConfig,
		providerKubectlClient: providerKubectlClient,
		writer:                writer,
		selfSigned:            false,
		templateBuilder: &CloudStackTemplateBuilder{
			datacenterConfigSpec:        &datacenterConfig.Spec,
			controlPlaneMachineSpec:     controlPlaneMachineSpec,
			WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
			etcdMachineSpec:             etcdMachineSpec,
			now:                         now,
		},
		skipIpCheck: skipIpCheck,
		validator:   NewValidator(providerCmkClient),
	}
}

func (p *cloudstackProvider) UpdateKubeConfig(_ *[]byte, _ string) error {
	// customize generated kube config
	return nil
}

func (p *cloudstackProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	return common.BootstrapClusterOpts(p.datacenterConfig.Spec.ManagementApiEndpoint, p.clusterConfig)
}

func (p *cloudstackProvider) Name() string {
	return constants.CloudStackProviderName
}

func (p *cloudstackProvider) DatacenterResourceType() string {
	return eksaCloudStackDatacenterResourceType
}

func (p *cloudstackProvider) MachineResourceType() string {
	return eksaCloudStackMachineResourceType
}

func (p *cloudstackProvider) setupSSHAuthKeysForCreate() error {
	var useKeyGeneratedForControlplane, useKeyGeneratedForWorker bool
	var err error
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
	p.controlPlaneSshAuthKey = controlPlaneUser.SshAuthorizedKeys[0]
	if len(p.controlPlaneSshAuthKey) > 0 {
		p.controlPlaneSshAuthKey, err = common.StripSshAuthorizedKeyComment(p.controlPlaneSshAuthKey)
		if err != nil {
			return err
		}
	} else {
		logger.Info("Provided control plane sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
		generatedKey, err := common.GenerateSSHAuthKey(p.writer)
		if err != nil {
			return err
		}
		p.controlPlaneSshAuthKey = generatedKey
		useKeyGeneratedForControlplane = true
	}
	workerUser := p.machineConfigs[p.clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name].Spec.Users[0]
	p.workerSshAuthKey = workerUser.SshAuthorizedKeys[0]
	if len(p.workerSshAuthKey) > 0 {
		p.workerSshAuthKey, err = common.StripSshAuthorizedKeyComment(p.workerSshAuthKey)
		if err != nil {
			return err
		}
	} else {
		if useKeyGeneratedForControlplane { // use the same key
			p.workerSshAuthKey = p.controlPlaneSshAuthKey
		} else {
			logger.Info("Provided worker sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
			generatedKey, err := common.GenerateSSHAuthKey(p.writer)
			if err != nil {
				return err
			}
			p.workerSshAuthKey = generatedKey
			useKeyGeneratedForWorker = true
		}
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.etcdSshAuthKey = etcdUser.SshAuthorizedKeys[0]
		if len(p.etcdSshAuthKey) > 0 {
			p.etcdSshAuthKey, err = common.StripSshAuthorizedKeyComment(p.etcdSshAuthKey)
			if err != nil {
				return err
			}
		} else {
			if useKeyGeneratedForControlplane { // use the same key as for controlplane
				p.etcdSshAuthKey = p.controlPlaneSshAuthKey
			} else if useKeyGeneratedForWorker {
				p.etcdSshAuthKey = p.workerSshAuthKey // if cp key was provided by user, check if worker key was generated by cli and use that
			} else {
				logger.Info("Provided etcd sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
				generatedKey, err := common.GenerateSSHAuthKey(p.writer)
				if err != nil {
					return err
				}
				p.etcdSshAuthKey = generatedKey
			}
		}
		etcdUser.SshAuthorizedKeys[0] = p.etcdSshAuthKey
	}
	controlPlaneUser.SshAuthorizedKeys[0] = p.controlPlaneSshAuthKey
	workerUser.SshAuthorizedKeys[0] = p.workerSshAuthKey
	return nil
}

func (p *cloudstackProvider) validateManagementApiEndpoint(rawurl string) error {
	_, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return fmt.Errorf("CloudStack managementApiEndpoint is invalid: #{err}")
	}
	return nil
}

func getHostnameFromUrl(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", fmt.Errorf("%s is not a valid url", rawurl)
	}
	return url.Hostname(), nil
}

func (p *cloudstackProvider) validateEnv(ctx context.Context) error {
	var cloudStackB64EncodedSecret string
	var ok bool

	if cloudStackB64EncodedSecret, ok = os.LookupEnv(decoder.EksacloudStackCloudConfigB64SecretKey); ok && len(cloudStackB64EncodedSecret) > 0 {
		if err := os.Setenv(decoder.CloudStackCloudConfigB64SecretKey, cloudStackB64EncodedSecret); err != nil {
			return fmt.Errorf("unable to set %s: %v", decoder.CloudStackCloudConfigB64SecretKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", decoder.EksacloudStackCloudConfigB64SecretKey)
	}
	execConfig, err := decoder.ParseCloudStackSecret()
	if err != nil {
		return fmt.Errorf("failed to parse environment variable exec config: %v", err)
	}
	if len(execConfig.ManagementUrl) <= 0 {
		return errors.New("cloudstack management api url is not set or is empty")
	}
	if err := p.validateManagementApiEndpoint(execConfig.ManagementUrl); err != nil {
		return errors.New("CloudStackDatacenterConfig managementApiEndpoint is invalid")
	}
	if _, ok := os.LookupEnv(eksaLicense); !ok {
		if err := os.Setenv(eksaLicense, ""); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksaLicense, err)
		}
	}
	return nil
}

func (p *cloudstackProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	cloudStackClusterSpec := NewSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	if err := p.validator.validateCloudStackAccess(ctx); err != nil {
		return err
	}
	if err := p.validator.ValidateCloudStackDatacenterConfig(ctx, p.datacenterConfig); err != nil {
		return err
	}
	if err := p.validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec); err != nil {
		return err
	}

	if err := p.setupSSHAuthKeysForCreate(); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	if clusterSpec.Cluster.IsManaged() {
		for _, mc := range p.MachineConfigs(clusterSpec) {
			em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, mc.GetName(), clusterSpec.ManagementCluster.KubeconfigFile, mc.GetNamespace())
			if err != nil {
				return err
			}
			if len(em) > 0 {
				return fmt.Errorf("CloudStackMachineConfig %s already exists", mc.GetName())
			}
		}
		existingDatacenter, err := p.providerKubectlClient.SearchCloudStackDatacenterConfig(ctx, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
		if err != nil {
			return err
		}
		if len(existingDatacenter) > 0 {
			return fmt.Errorf("CloudStackDatacenter %s already exists", p.datacenterConfig.Name)
		}
	}
	if p.skipIpCheck {
		logger.Info("Skipping check for whether control plane ip is in use")
		return nil
	}

	return nil
}

func (p *cloudstackProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	cloudStackClusterSpec := NewSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)
	if err := p.validator.validateCloudStackAccess(ctx); err != nil {
		return err
	}
	if err := p.validator.ValidateCloudStackDatacenterConfig(ctx, p.datacenterConfig); err != nil {
		return err
	}
	if err := p.validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec); err != nil {
		return err
	}

	if err := p.setupSSHAuthKeysForUpgrade(); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	err = p.validateMachineConfigsNameUniqueness(ctx, cluster, clusterSpec)
	if err != nil {
		return fmt.Errorf("failed validate machineconfig uniqueness: %v", err)
	}
	return nil
}

func (p *cloudstackProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	return nil
}

func NeedsNewControlPlaneTemplate(oldSpec, newSpec *cluster.Spec, oldCsdc, newCsdc *v1alpha1.CloudStackDatacenterConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	// Another option is to generate MachineTemplates based on the old and new eksa spec,
	// remove the name field and compare them with DeepEqual
	// We plan to approach this way since it's more flexible to add/remove fields and test out for validation
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host != newSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldCsdc, newCsdc, oldCsmc, newCsmc)
}

func NeedsNewWorkloadTemplate(oldSpec, newSpec *cluster.Spec, oldCsdc, newCsdc *v1alpha1.CloudStackDatacenterConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	if !v1alpha1.WorkerNodeGroupConfigurationSliceTaintsEqual(oldSpec.Cluster.Spec.WorkerNodeGroupConfigurations, newSpec.Cluster.Spec.WorkerNodeGroupConfigurations) ||
		!v1alpha1.WorkerNodeGroupConfigurationsLabelsMapEqual(oldSpec.Cluster.Spec.WorkerNodeGroupConfigurations, newSpec.Cluster.Spec.WorkerNodeGroupConfigurations) {
		return true
	}
	return AnyImmutableFieldChanged(oldCsdc, newCsdc, oldCsmc, newCsmc)
}

func NeedsNewKubeadmConfigTemplate(newWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration, oldWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration) bool {
	return !v1alpha1.TaintsSliceEqual(newWorkerNodeGroup.Taints, oldWorkerNodeGroup.Taints) || !v1alpha1.LabelsMapEqual(newWorkerNodeGroup.Labels, oldWorkerNodeGroup.Labels)
}

func NeedsNewEtcdTemplate(oldSpec, newSpec *cluster.Spec, oldVdc, newCsdc *v1alpha1.CloudStackDatacenterConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldVdc, newCsdc, oldCsmc, newCsmc)
}

func (p *cloudstackProvider) needsNewMachineTemplate(ctx context.Context, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, vdc *v1alpha1.CloudStackDatacenterConfig, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (bool, error) {
	if _, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		workerMachineConfig := p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		workerVmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, workerNodeGroupConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		if err != nil {
			return false, err
		}
		needsNewWorkloadTemplate := NeedsNewWorkloadTemplate(currentSpec, newClusterSpec, vdc, p.datacenterConfig, workerVmc, workerMachineConfig)
		return needsNewWorkloadTemplate, nil
	}
	return true, nil
}

func (p *cloudstackProvider) needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (bool, error) {
	if _, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		existingWorkerNodeGroupConfig := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]
		return NeedsNewKubeadmConfigTemplate(&workerNodeGroupConfiguration, &existingWorkerNodeGroupConfig), nil
	}
	return true, nil
}

func AnyImmutableFieldChanged(oldCsdc, newCsdc *v1alpha1.CloudStackDatacenterConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	for index, zone := range oldCsdc.Spec.Zones {
		if !zone.Equals(&newCsdc.Spec.Zones[index]) {
			return true
		}
	}
	if oldCsmc.Spec.Template != newCsmc.Spec.Template {
		return true
	}
	if oldCsmc.Spec.ComputeOffering != newCsmc.Spec.ComputeOffering {
		return true
	}
	if len(oldCsmc.Spec.UserCustomDetails) != len(newCsmc.Spec.UserCustomDetails) {
		return true
	}
	for key, value := range oldCsmc.Spec.UserCustomDetails {
		if value != newCsmc.Spec.UserCustomDetails[key] {
			return true
		}
	}
	return false
}

func NewCloudStackTemplateBuilder(CloudStackDatacenterConfigSpec *v1alpha1.CloudStackDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec, workerNodeGroupMachineSpecs map[string]v1alpha1.CloudStackMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &CloudStackTemplateBuilder{
		datacenterConfigSpec:        CloudStackDatacenterConfigSpec,
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		etcdMachineSpec:             etcdMachineSpec,
		now:                         now,
	}
}

type CloudStackTemplateBuilder struct {
	datacenterConfigSpec        *v1alpha1.CloudStackDatacenterConfigSpec
	controlPlaneMachineSpec     *v1alpha1.CloudStackMachineConfigSpec
	WorkerNodeGroupMachineSpecs map[string]v1alpha1.CloudStackMachineConfigSpec
	etcdMachineSpec             *v1alpha1.CloudStackMachineConfigSpec
	now                         types.NowFunc
}

func (cs *CloudStackTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	var etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *cs.etcdMachineSpec
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variable exec config: %v", err)
	}
	values := buildTemplateMapCP(clusterSpec, *cs.datacenterConfigSpec, *cs.controlPlaneMachineSpec, etcdMachineSpec)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (cs *CloudStackTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		values := buildTemplateMapMD(clusterSpec, *cs.datacenterConfigSpec, cs.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration)
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

func buildTemplateMapCP(clusterSpec *cluster.Spec, datacenterConfigSpec v1alpha1.CloudStackDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	host, port, _ := net.SplitHostPort(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
	etcdExtraArgs := clusterapi.SecureEtcdTlsCipherSuitesExtraArgs()
	sharedExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs()
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf)).
		Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration))
	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).
		Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).
		Append(clusterapi.PodIAMAuthExtraArgs(clusterSpec.Cluster.Spec.PodIAMConfig)).
		Append(sharedExtraArgs)

	values := map[string]interface{}{
		"clusterName":                                clusterSpec.Cluster.Name,
		"controlPlaneEndpointHost":                   host,
		"controlPlaneEndpointPort":                   port,
		"controlPlaneReplicas":                       clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"kubernetesRepository":                       bundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":                          bundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                             bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                               bundle.KubeDistro.Etcd.Tag,
		"corednsRepository":                          bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                             bundle.KubeDistro.CoreDNS.Tag,
		"nodeDriverRegistrarImage":                   bundle.KubeDistro.NodeDriverRegistrar.VersionedImage(),
		"livenessProbeImage":                         bundle.KubeDistro.LivenessProbe.VersionedImage(),
		"externalAttacherImage":                      bundle.KubeDistro.ExternalAttacher.VersionedImage(),
		"externalProvisionerImage":                   bundle.KubeDistro.ExternalProvisioner.VersionedImage(),
		"managerImage":                               bundle.CloudStack.ClusterAPIController.VersionedImage(),
		"cloudstackDomain":                           datacenterConfigSpec.Domain,
		"cloudstackZones":                            datacenterConfigSpec.Zones,
		"cloudstackAccount":                          datacenterConfigSpec.Account,
		"cloudstackControlPlaneComputeOfferingId":    controlPlaneMachineSpec.ComputeOffering.Id,
		"cloudstackControlPlaneComputeOfferingName":  controlPlaneMachineSpec.ComputeOffering.Name,
		"cloudstackControlPlaneTemplateOfferingId":   controlPlaneMachineSpec.Template.Id,
		"cloudstackControlPlaneTemplateOfferingName": controlPlaneMachineSpec.Template.Name,
		"cloudstackControlPlaneCustomDetails":        controlPlaneMachineSpec.UserCustomDetails,
		"affinityGroupIds":                           controlPlaneMachineSpec.AffinityGroupIds,
		"cloudstackEtcdComputeOfferingId":            etcdMachineSpec.ComputeOffering.Id,
		"cloudstackEtcdComputeOfferingName":          etcdMachineSpec.ComputeOffering.Name,
		"cloudstackEtcdTemplateOfferingId":           etcdMachineSpec.Template.Id,
		"cloudstackEtcdTemplateOfferingName":         etcdMachineSpec.Template.Name,
		"cloudstackEtcdCustomDetails":                etcdMachineSpec.UserCustomDetails,
		"cloudstackEtcdAffinityGroupIds":             etcdMachineSpec.AffinityGroupIds,
		"controlPlaneSshUsername":                    controlPlaneMachineSpec.Users[0].Name,
		"podCidrs":                                   clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                               clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"apiserverExtraArgs":                         apiServerExtraArgs.ToPartialYaml(),
		"kubeletExtraArgs":                           kubeletExtraArgs.ToPartialYaml(),
		"etcdExtraArgs":                              etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":                           crypto.SecureCipherSuitesString(),
		"controllermanagerExtraArgs":                 sharedExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":                         sharedExtraArgs.ToPartialYaml(),
		"format":                                     format,
		"externalEtcdVersion":                        bundle.KubeDistro.EtcdVersion,
		"etcdImage":                                  bundle.KubeDistro.EtcdImage.VersionedImage(),
		"eksaSystemNamespace":                        constants.EksaSystemNamespace,
		"auditPolicy":                                common.GetAuditPolicy(),
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint
		if len(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent
		}
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		capacity := len(clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks) +
			len(clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks) +
			len(clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy) + 4
		noProxyList := make([]string, 0, capacity)
		noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy...)

		// Add no-proxy defaults
		noProxyList = append(noProxyList, common.NoProxyDefaults...)
		cloudStackManagementApiEndpointHostname, err := getHostnameFromUrl(datacenterConfigSpec.ManagementApiEndpoint)
		if err == nil {
			noProxyList = append(noProxyList, cloudStackManagementApiEndpointHostname)
		}
		noProxyList = append(noProxyList,
			clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
	}

	if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 {
		values["controlPlaneTaints"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints
	}

	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}

	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, datacenterConfigSpec v1alpha1.CloudStackDatacenterConfigSpec, workerNodeGroupMachineSpec v1alpha1.CloudStackMachineConfigSpec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

	values := map[string]interface{}{
		"clusterName":                      clusterSpec.Cluster.Name,
		"kubernetesVersion":                bundle.KubeDistro.Kubernetes.Tag,
		"cloudstackTemplateId":             workerNodeGroupMachineSpec.Template.Id,
		"cloudstackTemplateName":           workerNodeGroupMachineSpec.Template.Name,
		"cloudstackOfferingId":             workerNodeGroupMachineSpec.ComputeOffering.Id,
		"cloudstackOfferingName":           workerNodeGroupMachineSpec.ComputeOffering.Name,
		"cloudstackCustomDetails":          workerNodeGroupMachineSpec.UserCustomDetails,
		"cloudstackAffinityGroupIds":       workerNodeGroupMachineSpec.AffinityGroupIds,
		"workerReplicas":                   clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count,
		"workerSshUsername":                workerNodeGroupMachineSpec.Users[0].Name,
		"cloudstackWorkerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys[0],
		"format":                           format,
		"kubeletExtraArgs":                 kubeletExtraArgs.ToPartialYaml(),
		"eksaSystemNamespace":              constants.EksaSystemNamespace,
		"workerNodeGroupName":              fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints":            workerNodeGroupConfiguration.Taints,
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint
		if len(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent
		}
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		capacity := len(clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks) +
			len(clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks) +
			len(clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy) + 4
		noProxyList := make([]string, 0, capacity)
		noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy...)

		// Add no-proxy defaults
		noProxyList = append(noProxyList, common.NoProxyDefaults...)
		cloudStackManagementApiEndpointHostname, err := getHostnameFromUrl(datacenterConfigSpec.ManagementApiEndpoint)
		if err == nil {
			noProxyList = append(noProxyList, cloudStackManagementApiEndpointHostname)
		}
		noProxyList = append(noProxyList,
			clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	return values
}

func (p *cloudstackProvider) generateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := clusterSpec.Cluster.Name

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = common.CPMachineTemplateName(clusterName, p.templateBuilder.now)
		values["cloudstackControlPlaneSshAuthorizedKey"] = p.controlPlaneSshAuthKey
		values["cloudstackEtcdSshAuthorizedKey"] = p.etcdSshAuthKey
		values["etcdTemplateName"] = common.EtcdMachineTemplateName(clusterName, p.templateBuilder.now)
	}
	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	if err != nil {
		return nil, nil, err
	}

	if len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations) > 1 {
		return nil, nil, fmt.Errorf("error generating cluster api Spec contents: multiple worker node group configurations are not supported for CloudStack provider")
	}

	workloadTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workloadTemplateNames[workerNodeGroupConfiguration.Name] = common.WorkerMachineTemplateName(clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
		kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = common.KubeadmConfigTemplateName(clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
		p.templateBuilder.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}

	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *cloudstackProvider) generateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := newClusterSpec.Cluster.Name
	var controlPlaneTemplateName, workloadTemplateName, kubeadmconfigTemplateName, etcdTemplateName string
	var needsNewEtcdTemplate bool

	c, err := p.providerKubectlClient.GetEksaCluster(ctx, workloadCluster, newClusterSpec.Cluster.Name)
	if err != nil {
		return nil, nil, err
	}
	vdc, err := p.providerKubectlClient.GetEksaCloudStackDatacenterConfig(ctx, p.datacenterConfig.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return nil, nil, err
	}
	controlPlaneMachineConfig := p.machineConfigs[newClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	controlPlaneVmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, c.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return nil, nil, err
	}
	needsNewControlPlaneTemplate := NeedsNewControlPlaneTemplate(currentSpec, newClusterSpec, vdc, p.datacenterConfig, controlPlaneVmc, controlPlaneMachineConfig)
	if !needsNewControlPlaneTemplate {
		cp, err := p.providerKubectlClient.GetKubeadmControlPlane(ctx, workloadCluster, c.Name, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
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
		needsNewWorkloadTemplate, err := p.needsNewMachineTemplate(ctx, workloadCluster, currentSpec, newClusterSpec, workerNodeGroupConfiguration, vdc, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, nil, err
		}

		needsNewKubeadmConfigTemplate, err := p.needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, nil, err
		}
		if !needsNewKubeadmConfigTemplate {
			mdName := machineDeploymentName(newClusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name)
			md, err := p.providerKubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
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
			md, err := p.providerKubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			workloadTemplateName = md.Spec.Template.Spec.InfrastructureRef.Name
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		} else {
			workloadTemplateName = common.WorkerMachineTemplateName(clusterName, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		}
		p.templateBuilder.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}

	if newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := p.machineConfigs[newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		etcdMachineVmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, c.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		if err != nil {
			return nil, nil, err
		}
		needsNewEtcdTemplate = NeedsNewEtcdTemplate(currentSpec, newClusterSpec, vdc, p.datacenterConfig, etcdMachineVmc, etcdMachineConfig)
		if !needsNewEtcdTemplate {
			etcdadmCluster, err := p.providerKubectlClient.GetEtcdadmCluster(ctx, workloadCluster, clusterName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			etcdTemplateName = etcdadmCluster.Spec.InfrastructureTemplate.Name
		} else {
			/* During a cluster upgrade, etcd machines need to be upgraded first, so that the etcd machines with new spec get created and can be used by controlplane machines
			   as etcd endpoints. KCP rollout should not start until then. As a temporary solution in the absence of static etcd endpoints, we annotate the etcd cluster as "upgrading",
			   so that KCP checks this annotation and does not proceed if etcd cluster is upgrading. The etcdadm controller removes this annotation once the etcd upgrade is complete.
			*/
			err = p.providerKubectlClient.UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", clusterName),
				map[string]string{etcdv1beta1.UpgradeInProgressAnnotation: "true"},
				executables.WithCluster(bootstrapCluster),
				executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			etcdTemplateName = common.EtcdMachineTemplateName(clusterName, p.templateBuilder.now)
		}
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["cloudstackControlPlaneSshAuthorizedKey"] = p.controlPlaneSshAuthKey
		values["cloudstackEtcdSshAuthorizedKey"] = p.etcdSshAuthKey
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

func (p *cloudstackProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForUpgrade(ctx, bootstrapCluster, workloadCluster, currentSpec, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating cluster api spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *cloudstackProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForCreate(ctx, cluster, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("generating cluster api Spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *cloudstackProvider) GenerateMHC() ([]byte, error) {
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

func (p *cloudstackProvider) CleanupProviderInfrastructure(_ context.Context) error {
	return nil
}

func (p *cloudstackProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// Nothing to do
	return nil
}

func (p *cloudstackProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.CloudStack.Version
}

func (p *cloudstackProvider) EnvMap(_ *cluster.Spec) (map[string]string, error) {
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

func (p *cloudstackProvider) GetDeployments() map[string][]string {
	return map[string][]string{"capc-system": {"capc-controller-manager"}}
}

func (p *cloudstackProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	bundle := clusterSpec.VersionsBundle
	folderName := fmt.Sprintf("infrastructure-cloudstack/%s/", bundle.CloudStack.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			bundle.CloudStack.Components,
			bundle.CloudStack.Metadata,
		},
	}
	return &infraBundle
}

func (p *cloudstackProvider) DatacenterConfig(_ *cluster.Spec) providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *cloudstackProvider) MachineConfigs(_ *cluster.Spec) []providers.MachineConfig {
	var configs []providers.MachineConfig
	controlPlaneMachineName := p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerMachineName := p.clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	p.machineConfigs[controlPlaneMachineName].Annotations = map[string]string{p.clusterConfig.ControlPlaneAnnotation(): "true"}
	if p.clusterConfig.IsManaged() {
		p.machineConfigs[controlPlaneMachineName].SetManagement(p.clusterConfig.ManagedBy())
	}

	configs = append(configs, p.machineConfigs[controlPlaneMachineName])
	if workerMachineName != controlPlaneMachineName {
		configs = append(configs, p.machineConfigs[workerMachineName])
		if p.clusterConfig.IsManaged() {
			p.machineConfigs[workerMachineName].SetManagement(p.clusterConfig.ManagedBy())
		}
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		p.machineConfigs[etcdMachineName].Annotations = map[string]string{p.clusterConfig.EtcdAnnotation(): "true"}
		if etcdMachineName != controlPlaneMachineName && etcdMachineName != workerMachineName {
			configs = append(configs, p.machineConfigs[etcdMachineName])
			p.machineConfigs[etcdMachineName].SetManagement(p.clusterConfig.ManagedBy())
		}
	}
	return configs
}

func (p *cloudstackProvider) UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec) (bool, error) {
	newV, oldV := newSpec.VersionsBundle.CloudStack, currentSpec.VersionsBundle.CloudStack

	return newV.ClusterAPIController.ImageDigest != oldV.ClusterAPIController.ImageDigest, nil
}

func (p *cloudstackProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range p.machineConfigs {
		if err := p.providerKubectlClient.DeleteEksaCloudStackMachineConfig(ctx, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaCloudStackDatacenterConfig(ctx, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, p.datacenterConfig.Namespace)
}

func (p *cloudstackProvider) GenerateStorageClass() []byte {
	return nil
}

func (p *cloudstackProvider) setupSSHAuthKeysForUpgrade() error {
	var err error
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
	p.controlPlaneSshAuthKey = controlPlaneUser.SshAuthorizedKeys[0]
	if len(p.controlPlaneSshAuthKey) > 0 {
		p.controlPlaneSshAuthKey, err = common.StripSshAuthorizedKeyComment(p.controlPlaneSshAuthKey)
		if err != nil {
			return err
		}
	}
	controlPlaneUser.SshAuthorizedKeys[0] = p.controlPlaneSshAuthKey
	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		workerUser := p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.workerSshAuthKey = workerUser.SshAuthorizedKeys[0]
		if len(p.workerSshAuthKey) > 0 {
			p.workerSshAuthKey, err = common.StripSshAuthorizedKeyComment(p.workerSshAuthKey)
			if err != nil {
				return err
			}
		}
		workerUser.SshAuthorizedKeys[0] = p.workerSshAuthKey
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.etcdSshAuthKey = etcdUser.SshAuthorizedKeys[0]
		if len(p.etcdSshAuthKey) > 0 {
			p.etcdSshAuthKey, err = common.StripSshAuthorizedKeyComment(p.etcdSshAuthKey)
			if err != nil {
				return err
			}
		}
		etcdUser.SshAuthorizedKeys[0] = p.etcdSshAuthKey
	}
	return nil
}

func (p *cloudstackProvider) validateMachineConfigsNameUniqueness(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName())
	if err != nil {
		return err
	}

	cpMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	if prevSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name != cpMachineConfigName {
		em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, cpMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace())
		if err != nil {
			return err
		}
		if len(em) > 0 {
			return fmt.Errorf("control plane VSphereMachineConfig %s already exists", cpMachineConfigName)
		}
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil && prevSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		if prevSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name != etcdMachineConfigName {
			em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, etcdMachineConfigName, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace())
			if err != nil {
				return err
			}
			if len(em) > 0 {
				return fmt.Errorf("external etcd machineconfig %s already exists", etcdMachineConfigName)
			}
		}
	}

	return nil
}

func machineDeploymentName(clusterName, nodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, nodeGroupName)
}
