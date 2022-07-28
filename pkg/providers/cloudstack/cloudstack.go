package cloudstack

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"

	etcdv1beta1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
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

var requiredEnvs = []string{decoder.CloudStackCloudConfigB64SecretKey}

var (
	eksaCloudStackDatacenterResourceType = fmt.Sprintf("cloudstackdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaCloudStackMachineResourceType    = fmt.Sprintf("cloudstackmachineconfigs.%s", v1alpha1.GroupVersion.Group)
)

type cloudstackProvider struct {
	datacenterConfig      *v1alpha1.CloudStackDatacenterConfig
	machineConfigs        map[string]*v1alpha1.CloudStackMachineConfig
	clusterConfig         *v1alpha1.Cluster
	providerKubectlClient ProviderKubectlClient
	writer                filewriter.FileWriter
	selfSigned            bool
	templateBuilder       *CloudStackTemplateBuilder
	skipIpCheck           bool
	validator             *Validator
	execConfig            *decoder.CloudStackExecConfig
}

func (p *cloudstackProvider) PreBootstrapSetup(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	logger.Info("Installing secrets on bootstrap cluster")
	return p.UpdateSecrets(ctx, cluster)
}

func (p *cloudstackProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *cloudstackProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	contents, err := p.generateSecrets(ctx, cluster)
	if err != nil {
		return fmt.Errorf("creating secrets object: %v", err)
	}

	if len(contents) > 0 {
		if err := p.providerKubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, contents); err != nil {
			return fmt.Errorf("applying secrets object: %v", err)
		}
	}
	return nil
}

func (p *cloudstackProvider) generateSecrets(ctx context.Context, cluster *types.Cluster) ([]byte, error) {
	secrets := [][]byte{}
	for _, profile := range p.execConfig.Profiles {
		_, err := p.providerKubectlClient.GetSecretFromNamespace(ctx, cluster.KubeconfigFile, profile.Name, constants.EksaSystemNamespace)
		if err == nil {
			// When a secret already exists with the profile name we skip creating it
			continue
		}
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("getting secret for profile %s: %v", profile.Name, err)
		}

		bytes, err := yaml.Marshal(generateSecret(profile))
		if err != nil {
			return nil, fmt.Errorf("marshalling secret for profile %s: %v", profile.Name, err)
		}
		secrets = append(secrets, bytes)
	}
	return templater.AppendYamlResources(secrets...), nil
}

func generateSecret(profile decoder.CloudStackProfileConfig) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
			Name:      profile.Name,
		},
		StringData: map[string]string{
			"api-url":    profile.ManagementUrl,
			"api-key":    profile.ApiKey,
			"secret-key": profile.SecretKey,
			"verify-ssl": profile.VerifySsl,
		},
	}
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

	oSpec := prevDatacenter.Spec
	nSpec := clusterSpec.CloudStackDatacenter.Spec

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
			if !zone.Equal(&oSpec.Zones[i]) {
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

func (p *cloudstackProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// Nothing to do
	return nil
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	CreateNamespaceIfNotPresent(ctx context.Context, kubeconfig string, namespace string) error
	LoadSecret(ctx context.Context, secretObject string, secretObjType string, secretObjectName string, kubeConfFile string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackDatacenterConfig, error)
	GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*kubeadmv1beta1.KubeadmControlPlane, error)
	GetMachineDeployment(ctx context.Context, workerNodeGroupName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1beta1.EtcdadmCluster, error)
	GetSecretFromNamespace(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
	SearchCloudStackMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackMachineConfig, error)
	SearchCloudStackDatacenterConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackDatacenterConfig, error)
	DeleteEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) error
	SetEksaControllerEnvVar(ctx context.Context, envVar, envVarVal, kubeconfig string) error
}

func NewProvider(datacenterConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerKubectlClient ProviderKubectlClient, providerCmkClient ProviderCmkClient, writer filewriter.FileWriter, now types.NowFunc, skipIpCheck bool) *cloudstackProvider {
	var controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec, len(machineConfigs))
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
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

func (p *cloudstackProvider) BootstrapClusterOpts(clusterSpec *cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error) {
	endpoints := []string{}
	for _, az := range clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones {
		endpoints = append(endpoints, az.ManagementApiEndpoint)
	}

	return common.BootstrapClusterOpts(p.clusterConfig, endpoints...)
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
	if len(controlPlaneUser.SshAuthorizedKeys[0]) > 0 {
		controlPlaneUser.SshAuthorizedKeys[0], err = common.StripSshAuthorizedKeyComment(controlPlaneUser.SshAuthorizedKeys[0])
		if err != nil {
			return err
		}
	} else {
		logger.Info("Provided control plane sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
		generatedKey, err := common.GenerateSSHAuthKey(p.writer)
		if err != nil {
			return err
		}
		controlPlaneUser.SshAuthorizedKeys[0] = generatedKey
		useKeyGeneratedForControlplane = true
	}
	var workerUser v1alpha1.UserConfiguration
	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		workerUser = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec.Users[0]
		if len(workerUser.SshAuthorizedKeys[0]) > 0 {
			workerUser.SshAuthorizedKeys[0], err = common.StripSshAuthorizedKeyComment(workerUser.SshAuthorizedKeys[0])
			if err != nil {
				return err
			}
		} else {
			if useKeyGeneratedForControlplane { // use the same key
				workerUser.SshAuthorizedKeys[0] = controlPlaneUser.SshAuthorizedKeys[0]
			} else {
				logger.Info("Provided worker sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
				generatedKey, err := common.GenerateSSHAuthKey(p.writer)
				if err != nil {
					return err
				}
				workerUser.SshAuthorizedKeys[0] = generatedKey
				useKeyGeneratedForWorker = true
			}
		}
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		if len(etcdUser.SshAuthorizedKeys[0]) > 0 {
			etcdUser.SshAuthorizedKeys[0], err = common.StripSshAuthorizedKeyComment(etcdUser.SshAuthorizedKeys[0])
			if err != nil {
				return err
			}
		} else {
			if useKeyGeneratedForControlplane { // use the same key as for controlplane
				etcdUser.SshAuthorizedKeys[0] = controlPlaneUser.SshAuthorizedKeys[0]
			} else if useKeyGeneratedForWorker {
				etcdUser.SshAuthorizedKeys[0] = workerUser.SshAuthorizedKeys[0] // if cp key was provided by user, check if worker key was generated by cli and use that
			} else {
				logger.Info("Provided etcd sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
				generatedKey, err := common.GenerateSSHAuthKey(p.writer)
				if err != nil {
					return err
				}
				etcdUser.SshAuthorizedKeys[0] = generatedKey
			}
		}
	}
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
	if len(execConfig.Profiles) <= 0 {
		return errors.New("cloudstack instances are not defined")
	}

	for _, instance := range execConfig.Profiles {
		if err := p.validateManagementApiEndpoint(instance.ManagementUrl); err != nil {
			return fmt.Errorf("CloudStack instance %s's managementApiEndpoint %s is invalid: %v",
				instance.Name, instance.ManagementUrl, err)
		}
	}
	p.execConfig = execConfig

	if _, ok := os.LookupEnv(eksaLicense); !ok {
		if err := os.Setenv(eksaLicense, ""); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksaLicense, err)
		}
	}
	return nil
}

// TODO: Consider to move this functionality to validator.go
func (p *cloudstackProvider) validateSecretsUnchanged(ctx context.Context, cluster *types.Cluster) error {
	for _, profile := range p.execConfig.Profiles {
		secret, err := p.providerKubectlClient.GetSecretFromNamespace(ctx, cluster.KubeconfigFile, profile.Name, constants.EksaSystemNamespace)
		if apierrors.IsNotFound(err) {
			// When the secret is not found we allow for new secrets
			continue
		}
		if err != nil {
			return fmt.Errorf("getting secret for profile %s: %v", profile.Name, err)
		}
		if secretDifferentFromProfile(secret, profile) {
			return fmt.Errorf("profile '%s' is different from the secret", profile.Name)
		}
	}
	return nil
}

func secretDifferentFromProfile(secret *corev1.Secret, profile decoder.CloudStackProfileConfig) bool {
	return string(secret.Data["api-url"]) != profile.ManagementUrl ||
		string(secret.Data["api-key"]) != profile.ApiKey ||
		string(secret.Data["secret-key"]) != profile.SecretKey ||
		string(secret.Data["verify-ssl"]) != profile.VerifySsl
}

func (p *cloudstackProvider) validateClusterSpec(ctx context.Context, clusterSpec *cluster.Spec) (err error) {
	if err := p.validator.validateCloudStackAccess(ctx, clusterSpec.CloudStackDatacenter); err != nil {
		return err
	}
	if err := p.validator.ValidateCloudStackDatacenterConfig(ctx, clusterSpec.CloudStackDatacenter); err != nil {
		return err
	}
	if err := p.validator.ValidateClusterMachineConfigs(ctx, NewSpec(clusterSpec, p.machineConfigs, clusterSpec.CloudStackDatacenter)); err != nil {
		return err
	}
	return nil
}

func (p *cloudstackProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := p.validateEnv(ctx); err != nil {
		return fmt.Errorf("validating environment variables: %v", err)
	}

	if err := p.validateClusterSpec(ctx, clusterSpec); err != nil {
		return fmt.Errorf("validating cluster spec: %v", err)
	}

	if err := p.setupSSHAuthKeysForCreate(); err != nil {
		return fmt.Errorf("setting up SSH keys: %v", err)
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
		existingDatacenter, err := p.providerKubectlClient.SearchCloudStackDatacenterConfig(ctx, clusterSpec.CloudStackDatacenter.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
		if err != nil {
			return err
		}
		if len(existingDatacenter) > 0 {
			return fmt.Errorf("CloudStackDatacenter %s already exists", clusterSpec.CloudStackDatacenter.Name)
		}
	}
	if p.skipIpCheck {
		logger.Info("Skipping check for whether control plane ip is in use")
		return nil
	}

	return nil
}

func (p *cloudstackProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, currentSpec *cluster.Spec) error {
	if err := p.validateEnv(ctx); err != nil {
		return fmt.Errorf("validating environment variables: %v", err)
	}

	if err := p.validateClusterSpec(ctx, clusterSpec); err != nil {
		return fmt.Errorf("validating cluster spec: %v", err)
	}

	if err := p.setupSSHAuthKeysForUpgrade(); err != nil {
		return fmt.Errorf("setting up SSH keys: %v", err)
	}

	if err := p.validateMachineConfigsNameUniqueness(ctx, cluster, clusterSpec); err != nil {
		return fmt.Errorf("failed validate machineconfig uniqueness: %v", err)
	}

	if err := p.validateSecretsUnchanged(ctx, cluster); err != nil {
		return fmt.Errorf("validating secrets unchanged: %v", err)
	}

	return nil
}

func (p *cloudstackProvider) SetupAndValidateDeleteCluster(ctx context.Context, _ *types.Cluster) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("validating environment variables: %v", err)
	}
	return nil
}

func needsNewControlPlaneTemplate(oldSpec, newSpec *cluster.Spec, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
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
	return AnyImmutableFieldChanged(oldSpec.CloudStackDatacenter, newSpec.CloudStackDatacenter, oldCsmc, newCsmc)
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

func needsNewEtcdTemplate(oldSpec, newSpec *cluster.Spec, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldSpec.CloudStackDatacenter, newSpec.CloudStackDatacenter, oldCsmc, newCsmc)
}

func (p *cloudstackProvider) needsNewMachineTemplate(ctx context.Context, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, csdc *v1alpha1.CloudStackDatacenterConfig, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (bool, error) {
	if _, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		workerMachineConfig := p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		workerVmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, workerNodeGroupConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		if err != nil {
			return false, err
		}
		needsNewWorkloadTemplate := NeedsNewWorkloadTemplate(currentSpec, newClusterSpec, csdc, newClusterSpec.CloudStackDatacenter, workerVmc, workerMachineConfig)
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
		if !zone.Equal(&newCsdc.Spec.Zones[index]) {
			return true
		}
	}
	if oldCsmc.Spec.Template != newCsmc.Spec.Template {
		return true
	}
	if oldCsmc.Spec.ComputeOffering != newCsmc.Spec.ComputeOffering {
		return true
	}
	if !oldCsmc.Spec.DiskOffering.Equal(&newCsmc.Spec.DiskOffering) {
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
	if len(oldCsmc.Spec.Symlinks) != len(newCsmc.Spec.Symlinks) {
		return true
	}
	for key, value := range oldCsmc.Spec.Symlinks {
		if value != newCsmc.Spec.Symlinks[key] {
			return true
		}
	}
	return false
}

func NewCloudStackTemplateBuilder(CloudStackDatacenterConfigSpec *v1alpha1.CloudStackDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec, workerNodeGroupMachineSpecs map[string]v1alpha1.CloudStackMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &CloudStackTemplateBuilder{
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		etcdMachineSpec:             etcdMachineSpec,
		now:                         now,
	}
}

type CloudStackTemplateBuilder struct {
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
	values := buildTemplateMapCP(clusterSpec, *cs.controlPlaneMachineSpec, etcdMachineSpec)

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
		values := buildTemplateMapMD(clusterSpec, cs.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration)
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

func buildTemplateMapCP(clusterSpec *cluster.Spec, controlPlaneMachineSpec, etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec) map[string]interface{} {
	datacenterConfigSpec := clusterSpec.CloudStackDatacenter.Spec
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
	controllerManagerExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.NodeCIDRMaskExtraArgs(&clusterSpec.Cluster.Spec.ClusterNetwork))

	values := map[string]interface{}{
		"clusterName":                                  clusterSpec.Cluster.Name,
		"controlPlaneEndpointHost":                     host,
		"controlPlaneEndpointPort":                     port,
		"controlPlaneReplicas":                         clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"kubernetesRepository":                         bundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":                            bundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                               bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                                 bundle.KubeDistro.Etcd.Tag,
		"corednsRepository":                            bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                               bundle.KubeDistro.CoreDNS.Tag,
		"nodeDriverRegistrarImage":                     bundle.KubeDistro.NodeDriverRegistrar.VersionedImage(),
		"livenessProbeImage":                           bundle.KubeDistro.LivenessProbe.VersionedImage(),
		"externalAttacherImage":                        bundle.KubeDistro.ExternalAttacher.VersionedImage(),
		"externalProvisionerImage":                     bundle.KubeDistro.ExternalProvisioner.VersionedImage(),
		"managerImage":                                 bundle.CloudStack.ClusterAPIController.VersionedImage(),
		"kubeVipImage":                                 bundle.CloudStack.KubeVip.VersionedImage(),
		"cloudstackKubeVip":                            !features.IsActive(features.CloudStackKubeVipDisabled()),
		"cloudstackAvailabilityZones":                  datacenterConfigSpec.AvailabilityZones,
		"cloudstackAnnotationSuffix":                   constants.CloudstackAnnotationSuffix,
		"cloudstackControlPlaneDiskOfferingProvided":   len(controlPlaneMachineSpec.DiskOffering.Id) > 0 || len(controlPlaneMachineSpec.DiskOffering.Name) > 0,
		"cloudstackControlPlaneDiskOfferingId":         controlPlaneMachineSpec.DiskOffering.Id,
		"cloudstackControlPlaneDiskOfferingName":       controlPlaneMachineSpec.DiskOffering.Name,
		"cloudstackControlPlaneDiskOfferingCustomSize": controlPlaneMachineSpec.DiskOffering.CustomSize,
		"cloudstackControlPlaneDiskOfferingPath":       controlPlaneMachineSpec.DiskOffering.MountPath,
		"cloudstackControlPlaneDiskOfferingDevice":     controlPlaneMachineSpec.DiskOffering.Device,
		"cloudstackControlPlaneDiskOfferingFilesystem": controlPlaneMachineSpec.DiskOffering.Filesystem,
		"cloudstackControlPlaneDiskOfferingLabel":      controlPlaneMachineSpec.DiskOffering.Label,
		"cloudstackControlPlaneComputeOfferingId":      controlPlaneMachineSpec.ComputeOffering.Id,
		"cloudstackControlPlaneComputeOfferingName":    controlPlaneMachineSpec.ComputeOffering.Name,
		"cloudstackControlPlaneTemplateOfferingId":     controlPlaneMachineSpec.Template.Id,
		"cloudstackControlPlaneTemplateOfferingName":   controlPlaneMachineSpec.Template.Name,
		"cloudstackControlPlaneCustomDetails":          controlPlaneMachineSpec.UserCustomDetails,
		"cloudstackControlPlaneSymlinks":               controlPlaneMachineSpec.Symlinks,
		"cloudstackControlPlaneAffinity":               controlPlaneMachineSpec.Affinity,
		"cloudstackControlPlaneAffinityGroupIds":       controlPlaneMachineSpec.AffinityGroupIds,
		"cloudstackEtcdDiskOfferingProvided":           len(etcdMachineSpec.DiskOffering.Id) > 0 || len(etcdMachineSpec.DiskOffering.Name) > 0,
		"cloudstackEtcdDiskOfferingId":                 etcdMachineSpec.DiskOffering.Id,
		"cloudstackEtcdDiskOfferingName":               etcdMachineSpec.DiskOffering.Name,
		"cloudstackEtcdDiskOfferingCustomSize":         etcdMachineSpec.DiskOffering.CustomSize,
		"cloudstackEtcdDiskOfferingPath":               etcdMachineSpec.DiskOffering.MountPath,
		"cloudstackEtcdDiskOfferingDevice":             etcdMachineSpec.DiskOffering.Device,
		"cloudstackEtcdDiskOfferingFilesystem":         etcdMachineSpec.DiskOffering.Filesystem,
		"cloudstackEtcdDiskOfferingLabel":              etcdMachineSpec.DiskOffering.Label,
		"cloudstackEtcdComputeOfferingId":              etcdMachineSpec.ComputeOffering.Id,
		"cloudstackEtcdComputeOfferingName":            etcdMachineSpec.ComputeOffering.Name,
		"cloudstackEtcdTemplateOfferingId":             etcdMachineSpec.Template.Id,
		"cloudstackEtcdTemplateOfferingName":           etcdMachineSpec.Template.Name,
		"cloudstackEtcdCustomDetails":                  etcdMachineSpec.UserCustomDetails,
		"cloudstackEtcdSymlinks":                       etcdMachineSpec.Symlinks,
		"cloudstackEtcdAffinity":                       etcdMachineSpec.Affinity,
		"cloudstackEtcdAffinityGroupIds":               etcdMachineSpec.AffinityGroupIds,
		"controlPlaneSshUsername":                      controlPlaneMachineSpec.Users[0].Name,
		"podCidrs":                                     clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                                 clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"apiserverExtraArgs":                           apiServerExtraArgs.ToPartialYaml(),
		"kubeletExtraArgs":                             kubeletExtraArgs.ToPartialYaml(),
		"etcdExtraArgs":                                etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":                             crypto.SecureCipherSuitesString(),
		"controllermanagerExtraArgs":                   controllerManagerExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":                           sharedExtraArgs.ToPartialYaml(),
		"format":                                       format,
		"externalEtcdVersion":                          bundle.KubeDistro.EtcdVersion,
		"etcdImage":                                    bundle.KubeDistro.EtcdImage.VersionedImage(),
		"eksaSystemNamespace":                          constants.EksaSystemNamespace,
		"auditPolicy":                                  common.GetAuditPolicy(),
	}
	values["cloudstackControlPlaneAnnotations"] = values["cloudstackControlPlaneDiskOfferingProvided"].(bool) || len(controlPlaneMachineSpec.Symlinks) > 0
	values["cloudstackEtcdAnnotations"] = values["cloudstackEtcdDiskOfferingProvided"].(bool) || len(etcdMachineSpec.Symlinks) > 0

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint
		if len(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent
		}
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		fillProxyConfigurations(values, clusterSpec)
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

func fillProxyConfigurations(values map[string]interface{}, clusterSpec *cluster.Spec) {
	datacenterConfigSpec := clusterSpec.CloudStackDatacenter.Spec
	values["proxyConfig"] = true
	capacity := len(clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks) +
		len(clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks) +
		len(clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy) + 4
	noProxyList := make([]string, 0, capacity)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks...)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks...)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy...)

	noProxyList = append(noProxyList, clusterapi.NoProxyDefaults()...)
	for _, az := range datacenterConfigSpec.AvailabilityZones {
		if cloudStackManagementApiEndpointHostname, err := getHostnameFromUrl(az.ManagementApiEndpoint); err == nil {
			noProxyList = append(noProxyList, cloudStackManagementApiEndpointHostname)
		}
	}
	noProxyList = append(noProxyList,
		clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
	)

	values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
	values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
	values["noProxy"] = noProxyList
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.CloudStackMachineConfigSpec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

	values := map[string]interface{}{
		"clusterName":                      clusterSpec.Cluster.Name,
		"kubernetesVersion":                bundle.KubeDistro.Kubernetes.Tag,
		"cloudstackAnnotationSuffix":       constants.CloudstackAnnotationSuffix,
		"cloudstackTemplateId":             workerNodeGroupMachineSpec.Template.Id,
		"cloudstackTemplateName":           workerNodeGroupMachineSpec.Template.Name,
		"cloudstackOfferingId":             workerNodeGroupMachineSpec.ComputeOffering.Id,
		"cloudstackOfferingName":           workerNodeGroupMachineSpec.ComputeOffering.Name,
		"cloudstackDiskOfferingProvided":   len(workerNodeGroupMachineSpec.DiskOffering.Id) > 0 || len(workerNodeGroupMachineSpec.DiskOffering.Name) > 0,
		"cloudstackDiskOfferingId":         workerNodeGroupMachineSpec.DiskOffering.Id,
		"cloudstackDiskOfferingName":       workerNodeGroupMachineSpec.DiskOffering.Name,
		"cloudstackDiskOfferingCustomSize": workerNodeGroupMachineSpec.DiskOffering.CustomSize,
		"cloudstackDiskOfferingPath":       workerNodeGroupMachineSpec.DiskOffering.MountPath,
		"cloudstackDiskOfferingDevice":     workerNodeGroupMachineSpec.DiskOffering.Device,
		"cloudstackDiskOfferingFilesystem": workerNodeGroupMachineSpec.DiskOffering.Filesystem,
		"cloudstackDiskOfferingLabel":      workerNodeGroupMachineSpec.DiskOffering.Label,
		"cloudstackCustomDetails":          workerNodeGroupMachineSpec.UserCustomDetails,
		"cloudstackSymlinks":               workerNodeGroupMachineSpec.Symlinks,
		"cloudstackAffinity":               workerNodeGroupMachineSpec.Affinity,
		"cloudstackAffinityGroupIds":       workerNodeGroupMachineSpec.AffinityGroupIds,
		"workerReplicas":                   workerNodeGroupConfiguration.Count,
		"workerSshUsername":                workerNodeGroupMachineSpec.Users[0].Name,
		"cloudstackWorkerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys[0],
		"format":                           format,
		"kubeletExtraArgs":                 kubeletExtraArgs.ToPartialYaml(),
		"eksaSystemNamespace":              constants.EksaSystemNamespace,
		"workerNodeGroupName":              fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints":            workerNodeGroupConfiguration.Taints,
	}
	values["cloudstackAnnotations"] = values["cloudstackDiskOfferingProvided"].(bool) || len(workerNodeGroupMachineSpec.Symlinks) > 0

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint
		if len(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent
		}
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		fillProxyConfigurations(values, clusterSpec)
	}

	return values
}

func (p *cloudstackProvider) generateCAPISpecForCreate(ctx context.Context, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := clusterSpec.Cluster.Name
	var etcdSshAuthorizedKey string
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdSshAuthorizedKey = p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0].SshAuthorizedKeys[0]
	}
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = common.CPMachineTemplateName(clusterName, p.templateBuilder.now)
		values["cloudstackControlPlaneSshAuthorizedKey"] = controlPlaneUser.SshAuthorizedKeys[0]
		values["cloudstackEtcdSshAuthorizedKey"] = etcdSshAuthorizedKey
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
		p.templateBuilder.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}

	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *cloudstackProvider) getControlPlaneNameForCAPISpecUpgrade(ctx context.Context, oldCluster *v1alpha1.Cluster, currentSpec, newClusterSpec *cluster.Spec, bootstrapCluster, workloadCluster *types.Cluster, csdc *v1alpha1.CloudStackDatacenterConfig, clusterName string) (string, error) {
	controlPlaneMachineConfig := p.machineConfigs[newClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	controlPlaneVmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, oldCluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return "", err
	}
	needsNewControlPlaneTemplate := needsNewControlPlaneTemplate(currentSpec, newClusterSpec, controlPlaneVmc, controlPlaneMachineConfig)
	if !needsNewControlPlaneTemplate {
		cp, err := p.providerKubectlClient.GetKubeadmControlPlane(ctx, workloadCluster, oldCluster.Name, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return "", err
		}
		return cp.Spec.MachineTemplate.InfrastructureRef.Name, nil
	} else {
		return common.CPMachineTemplateName(clusterName, p.templateBuilder.now), nil
	}
}

func (p *cloudstackProvider) getWorkloadTemplateSpecForCAPISpecUpgrade(ctx context.Context, currentSpec, newClusterSpec *cluster.Spec, bootstrapCluster, workloadCluster *types.Cluster, csdc *v1alpha1.CloudStackDatacenterConfig, clusterName string) ([]byte, error) {
	var kubeadmconfigTemplateName, workloadTemplateName string
	previousWorkerNodeGroupConfigs := cluster.BuildMapForWorkerNodeGroupsByName(currentSpec.Cluster.Spec.WorkerNodeGroupConfigurations)
	workloadTemplateNames := make(map[string]string, len(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		needsNewWorkloadTemplate, err := p.needsNewMachineTemplate(ctx, workloadCluster, currentSpec, newClusterSpec, workerNodeGroupConfiguration, csdc, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, err
		}

		needsNewKubeadmConfigTemplate, err := p.needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, err
		}
		if !needsNewKubeadmConfigTemplate {
			mdName := machineDeploymentName(newClusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name)
			md, err := p.providerKubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, err
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
				return nil, err
			}
			workloadTemplateName = md.Spec.Template.Spec.InfrastructureRef.Name
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		} else {
			workloadTemplateName = common.WorkerMachineTemplateName(clusterName, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		}
		p.templateBuilder.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}
	return p.templateBuilder.GenerateCAPISpecWorkers(newClusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
}

func (p *cloudstackProvider) getEtcdTemplateNameForCAPISpecUpgrade(ctx context.Context, oldCluster *v1alpha1.Cluster, currentSpec, newClusterSpec *cluster.Spec, bootstrapCluster, workloadCluster *types.Cluster, csdc *v1alpha1.CloudStackDatacenterConfig, clusterName string) (string, error) {
	etcdMachineConfig := p.machineConfigs[newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
	etcdMachineVmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, oldCluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return "", err
	}
	needsNewEtcdTemplate := needsNewEtcdTemplate(currentSpec, newClusterSpec, etcdMachineVmc, etcdMachineConfig)
	if !needsNewEtcdTemplate {
		etcdadmCluster, err := p.providerKubectlClient.GetEtcdadmCluster(ctx, workloadCluster, clusterName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return "", err
		}
		return etcdadmCluster.Spec.InfrastructureTemplate.Name, nil
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
			return "", err
		}
		return common.EtcdMachineTemplateName(clusterName, p.templateBuilder.now), nil
	}
}

func (p *cloudstackProvider) generateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := newClusterSpec.Cluster.Name
	var controlPlaneTemplateName, etcdTemplateName string

	c, err := p.providerKubectlClient.GetEksaCluster(ctx, workloadCluster, clusterName)
	if err != nil {
		return nil, nil, err
	}
	csdc, err := p.providerKubectlClient.GetEksaCloudStackDatacenterConfig(ctx, newClusterSpec.CloudStackDatacenter.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return nil, nil, err
	}

	controlPlaneTemplateName, err = p.getControlPlaneNameForCAPISpecUpgrade(ctx, c, currentSpec, newClusterSpec, bootstrapCluster, workloadCluster, csdc, clusterName)
	if err != nil {
		return nil, nil, err
	}

	workersSpec, err = p.getWorkloadTemplateSpecForCAPISpecUpgrade(ctx, currentSpec, newClusterSpec, bootstrapCluster, workloadCluster, csdc, clusterName)
	if err != nil {
		return nil, nil, err
	}

	if newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdTemplateName, err = p.getEtcdTemplateNameForCAPISpecUpgrade(ctx, c, currentSpec, newClusterSpec, bootstrapCluster, workloadCluster, csdc, clusterName)
		if err != nil {
			return nil, nil, err
		}
	}
	var etcdSshAuthorizedKey string
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		if len(etcdUser.SshAuthorizedKeys) > 0 {
			etcdSshAuthorizedKey = etcdUser.SshAuthorizedKeys[0]
		}
	}
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["cloudstackControlPlaneSshAuthorizedKey"] = controlPlaneUser.SshAuthorizedKeys[0]
		values["cloudstackEtcdSshAuthorizedKey"] = etcdSshAuthorizedKey
		values["etcdTemplateName"] = etcdTemplateName
	}
	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(newClusterSpec, cpOpt)
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

func (p *cloudstackProvider) GenerateCAPISpecForCreate(ctx context.Context, _ *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForCreate(ctx, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("generating cluster api Spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *cloudstackProvider) machineConfigsSpecChanged(ctx context.Context, cc *v1alpha1.Cluster, cluster *types.Cluster, newClusterSpec *cluster.Spec) (bool, error) {
	machineConfigMap := make(map[string]*v1alpha1.CloudStackMachineConfig)
	for _, config := range p.MachineConfigs(nil) {
		mc := config.(*v1alpha1.CloudStackMachineConfig)
		machineConfigMap[mc.Name] = mc
	}

	for _, oldMcRef := range cc.MachineConfigRefs() {
		existingCsmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, oldMcRef.Name, cluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		if err != nil {
			return false, err
		}
		csmc, ok := machineConfigMap[oldMcRef.Name]
		if !ok {
			logger.V(3).Info(fmt.Sprintf("Old machine config spec %s not found in the existing spec", oldMcRef.Name))
			return true, nil
		}
		if !existingCsmc.Spec.Equal(&csmc.Spec) {
			logger.V(3).Info(fmt.Sprintf("New machine config spec %s is different from the existing spec", oldMcRef.Name))
			return true, nil
		}
	}

	return false, nil
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

func (p *cloudstackProvider) DatacenterConfig(clusterSpec *cluster.Spec) providers.DatacenterConfig {
	return clusterSpec.CloudStackDatacenter
}

func (p *cloudstackProvider) MachineConfigs(_ *cluster.Spec) []providers.MachineConfig {
	configs := make(map[string]providers.MachineConfig, len(p.machineConfigs))
	controlPlaneMachineName := p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	p.machineConfigs[controlPlaneMachineName].Annotations = map[string]string{p.clusterConfig.ControlPlaneAnnotation(): "true"}
	if p.clusterConfig.IsManaged() {
		p.machineConfigs[controlPlaneMachineName].SetManagement(p.clusterConfig.ManagedBy())
	}
	configs[controlPlaneMachineName] = p.machineConfigs[controlPlaneMachineName]

	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		p.machineConfigs[etcdMachineName].Annotations = map[string]string{p.clusterConfig.EtcdAnnotation(): "true"}
		if etcdMachineName != controlPlaneMachineName {
			configs[etcdMachineName] = p.machineConfigs[etcdMachineName]
			if p.clusterConfig.IsManaged() {
				p.machineConfigs[etcdMachineName].SetManagement(p.clusterConfig.ManagedBy())
			}
		}
	}

	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		workerMachineName := workerNodeGroupConfiguration.MachineGroupRef.Name
		if _, ok := configs[workerMachineName]; !ok {
			configs[workerMachineName] = p.machineConfigs[workerMachineName]
			if p.clusterConfig.IsManaged() {
				p.machineConfigs[workerMachineName].SetManagement(p.clusterConfig.ManagedBy())
			}
		}
	}
	return providers.ConfigsMapToSlice(configs)
}

func (p *cloudstackProvider) UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec, cluster *types.Cluster) (bool, error) {
	cc := currentSpec.Cluster
	existingCsdc, err := p.providerKubectlClient.GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, newSpec.Cluster.Namespace)
	if err != nil {
		return false, err
	}
	if !existingCsdc.Spec.Equal(&newSpec.CloudStackDatacenter.Spec) {
		logger.V(3).Info("New provider spec is different from the new spec")
		return true, nil
	}

	machineConfigsSpecChanged, err := p.machineConfigsSpecChanged(ctx, cc, cluster, newSpec)
	if err != nil {
		return false, err
	}
	return machineConfigsSpecChanged, nil
}

func (p *cloudstackProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range p.machineConfigs {
		if err := p.providerKubectlClient.DeleteEksaCloudStackMachineConfig(ctx, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaCloudStackDatacenterConfig(ctx, clusterSpec.CloudStackDatacenter.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.CloudStackDatacenter.Namespace)
}

func (p *cloudstackProvider) PostClusterDeleteValidate(_ context.Context, _ *types.Cluster) error {
	// No validations
	return nil
}

func (p *cloudstackProvider) GenerateStorageClass() []byte {
	return nil
}

func (p *cloudstackProvider) setupSSHAuthKeysForUpgrade() error {
	var err error
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
	if len(controlPlaneUser.SshAuthorizedKeys[0]) > 0 {
		controlPlaneUser.SshAuthorizedKeys[0], err = common.StripSshAuthorizedKeyComment(controlPlaneUser.SshAuthorizedKeys[0])
		if err != nil {
			return err
		}
	}
	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		workerUser := p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec.Users[0]
		if len(workerUser.SshAuthorizedKeys[0]) > 0 {
			workerUser.SshAuthorizedKeys[0], err = common.StripSshAuthorizedKeyComment(workerUser.SshAuthorizedKeys[0])
			if err != nil {
				return err
			}
		}
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		if len(etcdUser.SshAuthorizedKeys[0]) > 0 {
			etcdUser.SshAuthorizedKeys[0], err = common.StripSshAuthorizedKeyComment(etcdUser.SshAuthorizedKeys[0])
			if err != nil {
				return err
			}
		}
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
		err := p.validateMachineConfigNameUniqueness(ctx, cpMachineConfigName, clusterSpec)
		if err != nil {
			return err
		}
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil && prevSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		if prevSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name != etcdMachineConfigName {
			err := p.validateMachineConfigNameUniqueness(ctx, etcdMachineConfigName, clusterSpec)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *cloudstackProvider) validateMachineConfigNameUniqueness(ctx context.Context, machineConfigName string, clusterSpec *cluster.Spec) error {
	em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, machineConfigName, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace())
	if err != nil {
		return err
	}
	if len(em) > 0 {
		return fmt.Errorf("machineconfig %s already exists", machineConfigName)
	}
	return nil
}

func (p *cloudstackProvider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	if err := p.providerKubectlClient.SetEksaControllerEnvVar(ctx, features.CloudStackProviderEnvVar, "true", kubeconfigFile); err != nil {
		return err
	}
	kubeVipDisabledString := strconv.FormatBool(features.IsActive(features.CloudStackKubeVipDisabled()))
	return p.providerKubectlClient.SetEksaControllerEnvVar(ctx, features.CloudStackKubeVipDisabledEnvVar, kubeVipDisabledString, kubeconfigFile)
}

func machineDeploymentName(clusterName, nodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, nodeGroupName)
}

func (p *cloudstackProvider) PostClusterDeleteForUpgrade(ctx context.Context, managementCluster *types.Cluster) error {
	return nil
}
