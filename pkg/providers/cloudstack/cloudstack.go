package cloudstack

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	etcdv1beta1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
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
	eksaLicense         = "EKSA_LICENSE"
	etcdTemplateNameKey = "etcdTemplateName"
	cpTemplateNameKey   = "controlPlaneTemplateName"
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
	clusterConfig         *v1alpha1.Cluster
	providerKubectlClient ProviderKubectlClient
	writer                filewriter.FileWriter
	selfSigned            bool
	templateBuilder       *TemplateBuilder
	validator             ProviderValidator
	execConfig            *decoder.CloudStackExecConfig
	log                   logr.Logger
}

func (p *cloudstackProvider) PreBootstrapSetup(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	p.log.Info("Installing secrets on bootstrap cluster")
	return p.UpdateSecrets(ctx, cluster, nil)
}

func (p *cloudstackProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

// PostBootstrapDeleteForUpgrade runs any provider-specific operations after bootstrap cluster has been deleted.
func (p *cloudstackProvider) PostBootstrapDeleteForUpgrade(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *cloudstackProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster, _ *cluster.Spec) error {
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
			decoder.APIUrlKey:    profile.ManagementUrl,
			decoder.APIKeyKey:    profile.ApiKey,
			decoder.SecretKeyKey: profile.SecretKey,
			decoder.VerifySslKey: profile.VerifySsl,
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
	prevMachineConfig, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, newConfig.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
	if err != nil {
		return err
	}

	err = newConfig.ValidateUpdate(prevMachineConfig)
	if err != nil {
		return err
	}
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

	prevDatacenter.SetDefaults()

	if err = clusterSpec.CloudStackDatacenter.ValidateUpdate(prevDatacenter); err != nil {
		return err
	}

	prevMachineConfigRefs := machineRefSliceToMap(prevSpec.MachineConfigRefs())

	for _, machineConfigRef := range clusterSpec.Cluster.MachineConfigRefs() {
		machineConfig, ok := clusterSpec.CloudStackMachineConfigs[machineConfigRef.Name]
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

	return nil
}

// ChangeDiff returns the component change diff for the provider.
func (p *cloudstackProvider) ChangeDiff(currentComponents, newComponents *cluster.ManagementComponents) *types.ComponentChangeDiff {
	if currentComponents.CloudStack.Version == newComponents.CloudStack.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.CloudStackProviderName,
		NewVersion:    newComponents.CloudStack.Version,
		OldVersion:    currentComponents.CloudStack.Version,
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

// NewProvider initializes the CloudStack provider object.
func NewProvider(datacenterConfig *v1alpha1.CloudStackDatacenterConfig, clusterConfig *v1alpha1.Cluster, providerKubectlClient ProviderKubectlClient, validator ProviderValidator, writer filewriter.FileWriter, now types.NowFunc, log logr.Logger) *cloudstackProvider { //nolint:revive
	return &cloudstackProvider{
		datacenterConfig:      datacenterConfig,
		clusterConfig:         clusterConfig,
		providerKubectlClient: providerKubectlClient,
		writer:                writer,
		selfSigned:            false,
		templateBuilder:       NewTemplateBuilder(now),
		log:                   log,
		validator:             validator,
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

func (p *cloudstackProvider) generateSSHKeysIfNotSet(machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) error {
	var generatedKey string
	for _, machineConfig := range machineConfigs {
		user := machineConfig.Spec.Users[0]
		if user.SshAuthorizedKeys[0] == "" {
			if generatedKey != "" { // use same key already generated
				user.SshAuthorizedKeys[0] = generatedKey
			} else { // generate new key
				logger.Info("Provided sshAuthorizedKey is not set or is empty, auto-generating new key pair...", "cloudstackMachineConfig", machineConfig.Name)
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

func (p *cloudstackProvider) setMachineConfigDefaults(clusterSpec *cluster.Spec) {
	for _, mc := range clusterSpec.CloudStackMachineConfigs {
		mc.SetUserDefaults()
	}
}

func (p *cloudstackProvider) validateManagementApiEndpoint(rawurl string) error {
	_, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return fmt.Errorf("CloudStack managementApiEndpoint is invalid: #{err}")
	}
	return nil
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
	execConfig, err := decoder.ParseCloudStackCredsFromEnv()
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

func (p *cloudstackProvider) validateClusterSpec(ctx context.Context, clusterSpec *cluster.Spec) (err error) {
	if err := p.validator.ValidateCloudStackDatacenterConfig(ctx, clusterSpec.CloudStackDatacenter); err != nil {
		return err
	}
	if err := p.validator.ValidateClusterMachineConfigs(ctx, clusterSpec); err != nil {
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

	if err := p.validator.ValidateControlPlaneEndpointUniqueness(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host); err != nil {
		return fmt.Errorf("validating control plane endpoint uniqueness: %v", err)
	}

	if err := p.generateSSHKeysIfNotSet(clusterSpec.CloudStackMachineConfigs); err != nil {
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

	return nil
}

func (p *cloudstackProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, currentSpec *cluster.Spec) error {
	if err := p.SetupAndValidateUpgradeManagementComponents(ctx, clusterSpec); err != nil {
		return err
	}

	p.setMachineConfigDefaults(clusterSpec)

	if err := p.validateClusterSpec(ctx, clusterSpec); err != nil {
		return fmt.Errorf("validating cluster spec: %v", err)
	}

	if err := p.validateMachineConfigsNameUniqueness(ctx, cluster, clusterSpec); err != nil {
		return fmt.Errorf("failed validate machineconfig uniqueness: %v", err)
	}

	if err := p.validator.ValidateSecretsUnchanged(ctx, cluster, p.execConfig, p.providerKubectlClient); err != nil {
		return fmt.Errorf("validating secrets unchanged: %v", err)
	}

	return nil
}

func (p *cloudstackProvider) SetupAndValidateDeleteCluster(ctx context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("validating environment variables: %v", err)
	}
	return nil
}

// SetupAndValidateUpgradeManagementComponents performs necessary setup for upgrade management components operation.
func (p *cloudstackProvider) SetupAndValidateUpgradeManagementComponents(ctx context.Context, _ *cluster.Spec) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("validating environment variables: %v", err)
	}
	return nil
}

func needsNewControlPlaneTemplate(oldSpec, newSpec *cluster.Spec, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig, log logr.Logger) bool {
	// Another option is to generate MachineTemplates based on the old and new eksa spec,
	// remove the name field and compare them with DeepEqual
	// We plan to approach this way since it's more flexible to add/remove fields and test out for validation
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return NeedNewMachineTemplate(oldSpec.CloudStackDatacenter, newSpec.CloudStackDatacenter, oldCsmc, newCsmc, log)
}

// NeedsNewWorkloadTemplate determines if a new workload template is needed.
func NeedsNewWorkloadTemplate(oldSpec, newSpec *cluster.Spec, oldCsdc, newCsdc *v1alpha1.CloudStackDatacenterConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig, oldWorker, newWorker *v1alpha1.WorkerNodeGroupConfiguration, log logr.Logger) bool {
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	if !v1alpha1.TaintsSliceEqual(oldWorker.Taints, newWorker.Taints) ||
		!v1alpha1.MapEqual(oldWorker.Labels, newWorker.Labels) ||
		!v1alpha1.WorkerNodeGroupConfigurationKubeVersionUnchanged(oldWorker, newWorker, oldSpec.Cluster, newSpec.Cluster) {
		return true
	}
	return NeedNewMachineTemplate(oldCsdc, newCsdc, oldCsmc, newCsmc, log)
}

func NeedsNewKubeadmConfigTemplate(newWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration, oldWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration) bool {
	return !v1alpha1.TaintsSliceEqual(newWorkerNodeGroup.Taints, oldWorkerNodeGroup.Taints) || !v1alpha1.MapEqual(newWorkerNodeGroup.Labels, oldWorkerNodeGroup.Labels)
}

func needsNewEtcdTemplate(oldSpec, newSpec *cluster.Spec, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig, log logr.Logger) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return NeedNewMachineTemplate(oldSpec.CloudStackDatacenter, newSpec.CloudStackDatacenter, oldCsmc, newCsmc, log)
}

func (p *cloudstackProvider) needsNewMachineTemplate(ctx context.Context, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, csdc *v1alpha1.CloudStackDatacenterConfig, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (bool, error) {
	if oldWorkerNodeGroup, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		newWorkerMachineConfig := newClusterSpec.CloudStackMachineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		oldWorkerMachineConfig, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, oldWorkerNodeGroup.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		if err != nil {
			return false, err
		}
		needsNewWorkloadTemplate := NeedsNewWorkloadTemplate(currentSpec, newClusterSpec, csdc, newClusterSpec.CloudStackDatacenter, oldWorkerMachineConfig, newWorkerMachineConfig, &oldWorkerNodeGroup, &workerNodeGroupConfiguration, p.log)
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

// NeedNewMachineTemplate Used by EKS-A controller and CLI upgrade workflow to compare generated CSDC/CSMC's from
// CAPC resources in fetcher.go with those already on the cluster when deciding whether or not to generate and apply
// new CloudStackMachineTemplates.
func NeedNewMachineTemplate(
	oldDatacenterConfig, newDatacenterConfig *v1alpha1.CloudStackDatacenterConfig,
	oldMachineConfig, newMachineConfig *v1alpha1.CloudStackMachineConfig,
	log logr.Logger,
) bool {
	oldAzs := oldDatacenterConfig.Spec.AvailabilityZones
	newAzs := newDatacenterConfig.Spec.AvailabilityZones
	if !hasSameAvailabilityZones(oldAzs, newAzs) {
		log.V(4).Info(
			"CloudStackDatacenterConfigs do not match",
			"oldAvailabilityZones", oldDatacenterConfig.Spec.AvailabilityZones,
			"newAvailabilityZones", newDatacenterConfig.Spec.AvailabilityZones,
		)
		return true
	}

	if !oldMachineConfig.Spec.Template.Equal(&newMachineConfig.Spec.Template) {
		log.V(4).Info(
			"Old and new CloudStackMachineConfig Templates do not match",
			"machineConfig", oldMachineConfig.Name,
			"oldTemplate", oldMachineConfig.Spec.Template,
			"newTemplate", newMachineConfig.Spec.Template,
		)
		return true
	}

	if !oldMachineConfig.Spec.ComputeOffering.Equal(&newMachineConfig.Spec.ComputeOffering) {
		log.V(4).Info(
			"Old and new CloudStackMachineConfig Compute Offerings do not match",
			"machineConfig", oldMachineConfig.Name,
			"oldComputeOffering", oldMachineConfig.Spec.ComputeOffering,
			"newComputeOffering", newMachineConfig.Spec.ComputeOffering,
		)
		return true
	}

	if !oldMachineConfig.Spec.DiskOffering.Equal(newMachineConfig.Spec.DiskOffering) {
		log.V(4).Info(
			"Old and new CloudStackMachineConfig DiskOffering does not match",
			"machineConfig", oldMachineConfig.Name,
			"oldDiskOffering", oldMachineConfig.Spec.DiskOffering,
			"newDiskOffering", newMachineConfig.Spec.DiskOffering,
		)
		return true
	}

	if !isEqualMap(oldMachineConfig.Spec.UserCustomDetails, newMachineConfig.Spec.UserCustomDetails) {
		log.V(4).Info(
			"Old and new CloudStackMachineConfig UserCustomDetails does not match",
			"machineConfig", oldMachineConfig.Name,
			"oldUserCustomDetails", oldMachineConfig.Spec.UserCustomDetails,
			"newUserCustomDetails", newMachineConfig.Spec.UserCustomDetails,
		)
		return true
	}

	if !isEqualMap(oldMachineConfig.Spec.Symlinks, newMachineConfig.Spec.Symlinks) {
		log.V(4).Info(
			"Old and new CloudStackMachineConfig Symlinks does not match",
			"machineConfig", oldMachineConfig.Name,
			"oldSymlinks", oldMachineConfig.Spec.Symlinks,
			"newSymlinks", newMachineConfig.Spec.Symlinks,
		)
		return true
	}

	return false
}

func isEqualMap[K, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}

	// Ensure all keys are present in b, and a's values equal b's values.
	for k, av := range a {
		if bv, ok := b[k]; !ok || av != bv {
			return false
		}
	}

	return true
}

func hasSameAvailabilityZones(old, nw []v1alpha1.CloudStackAvailabilityZone) bool {
	if len(old) != len(nw) {
		return false
	}

	oldAzs := map[string]v1alpha1.CloudStackAvailabilityZone{}
	for _, az := range old {
		oldAzs[az.Name] = az
	}

	// Equality of availability zones doesn't take into consideration the availability zones
	// ManagementApiEndpoint. Its unclear why this is the case. The ManagementApiEndpoint seems
	// to only be used for proxy configuration.
	equal := func(old, nw v1alpha1.CloudStackAvailabilityZone) bool {
		return old.Zone.Equal(&nw.Zone) &&
			old.Name == nw.Name &&
			old.CredentialsRef == nw.CredentialsRef &&
			old.Account == nw.Account &&
			old.Domain == nw.Domain
	}

	for _, newAz := range nw {
		oldAz, found := oldAzs[newAz.Name]
		if !found || !equal(oldAz, newAz) {
			return false
		}
	}

	return true
}

func (p *cloudstackProvider) generateCAPISpecForCreate(ctx context.Context, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := clusterSpec.Cluster.Name
	cpOpt := func(values map[string]interface{}) {
		values[cpTemplateNameKey] = common.CPMachineTemplateName(clusterName, p.templateBuilder.now)
		values[etcdTemplateNameKey] = common.EtcdMachineTemplateName(clusterName, p.templateBuilder.now)
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
	}

	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *cloudstackProvider) getControlPlaneNameForCAPISpecUpgrade(ctx context.Context, oldCluster *v1alpha1.Cluster, currentSpec, newClusterSpec *cluster.Spec, bootstrapCluster, workloadCluster *types.Cluster, csdc *v1alpha1.CloudStackDatacenterConfig, clusterName string) (string, error) {
	controlPlaneMachineConfig := newClusterSpec.CloudStackMachineConfigs[newClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	controlPlaneVmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, oldCluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return "", err
	}
	if !needsNewControlPlaneTemplate(currentSpec, newClusterSpec, controlPlaneVmc, controlPlaneMachineConfig, p.log) {
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
	}
	return p.templateBuilder.GenerateCAPISpecWorkers(newClusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
}

func (p *cloudstackProvider) getEtcdTemplateNameForCAPISpecUpgrade(ctx context.Context, oldCluster *v1alpha1.Cluster, currentSpec, newClusterSpec *cluster.Spec, bootstrapCluster, workloadCluster *types.Cluster, csdc *v1alpha1.CloudStackDatacenterConfig, clusterName string) (string, error) {
	etcdMachineConfig := newClusterSpec.CloudStackMachineConfigs[newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
	etcdMachineVmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, oldCluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return "", err
	}
	needsNewEtcdTemplate := needsNewEtcdTemplate(currentSpec, newClusterSpec, etcdMachineVmc, etcdMachineConfig, p.log)
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

	csdc.SetDefaults()
	currentSpec.CloudStackDatacenter = csdc

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

	cpOpt := func(values map[string]interface{}) {
		values[cpTemplateNameKey] = controlPlaneTemplateName
		values[etcdTemplateNameKey] = etcdTemplateName
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
	for _, oldMcRef := range cc.MachineConfigRefs() {
		existingCsmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, oldMcRef.Name, cluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		if err != nil {
			return false, err
		}
		csmc, ok := newClusterSpec.CloudStackMachineConfigs[oldMcRef.Name]
		if !ok {
			p.log.V(3).Info(fmt.Sprintf("Old machine config spec %s not found in the existing spec", oldMcRef.Name))
			return true, nil
		}
		if !existingCsmc.Spec.Equal(&csmc.Spec) {
			p.log.V(3).Info(fmt.Sprintf("New machine config spec %s is different from the existing spec", oldMcRef.Name))
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

// Version returns the version of the provider.
func (p *cloudstackProvider) Version(componnets *cluster.ManagementComponents) string {
	return componnets.CloudStack.Version
}

// EnvMap returns a map of environment variables required for the cloudstack provider.
func (p *cloudstackProvider) EnvMap(_ *cluster.ManagementComponents, _ *cluster.Spec) (map[string]string, error) {
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

// GetInfrastructureBundle returns the infrastructure bundle for the provider.
func (p *cloudstackProvider) GetInfrastructureBundle(components *cluster.ManagementComponents) *types.InfrastructureBundle {
	folderName := fmt.Sprintf("infrastructure-cloudstack/%s/", components.CloudStack.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			components.CloudStack.Components,
			components.CloudStack.Metadata,
		},
	}
	return &infraBundle
}

func (p *cloudstackProvider) DatacenterConfig(clusterSpec *cluster.Spec) providers.DatacenterConfig {
	return clusterSpec.CloudStackDatacenter
}

func (p *cloudstackProvider) MachineConfigs(spec *cluster.Spec) []providers.MachineConfig {
	annotateMachineConfig(
		spec,
		spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name,
		spec.Cluster.ControlPlaneAnnotation(),
		"true",
	)
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		annotateMachineConfig(
			spec,
			spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name,
			spec.Cluster.EtcdAnnotation(),
			"true",
		)
	}
	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		setMachineConfigManagedBy(spec, workerNodeGroupConfiguration.MachineGroupRef.Name)
	}
	machineConfigs := make([]providers.MachineConfig, 0, len(spec.CloudStackMachineConfigs))
	for _, m := range spec.CloudStackMachineConfigs {
		machineConfigs = append(machineConfigs, m)
	}
	return machineConfigs
}

func annotateMachineConfig(spec *cluster.Spec, machineConfigName, annotationKey, annotationValue string) {
	machineConfig := spec.CloudStackMachineConfigs[machineConfigName]
	if machineConfig.Annotations == nil {
		machineConfig.Annotations = make(map[string]string, 1)
	}
	machineConfig.Annotations[annotationKey] = annotationValue
	setMachineConfigManagedBy(spec, machineConfigName)
}

func setMachineConfigManagedBy(spec *cluster.Spec, machineConfigName string) {
	machineConfig := spec.CloudStackMachineConfigs[machineConfigName]
	if spec.Cluster.IsManaged() {
		machineConfig.SetManagement(spec.Cluster.ManagedBy())
	}
}

func (p *cloudstackProvider) UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec, cluster *types.Cluster) (bool, error) {
	cc := currentSpec.Cluster
	existingCsdc, err := p.providerKubectlClient.GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, newSpec.Cluster.Namespace)
	if err != nil {
		return false, err
	}
	existingCsdc.SetDefaults()
	currentSpec.CloudStackDatacenter = existingCsdc
	if !existingCsdc.Spec.Equal(&newSpec.CloudStackDatacenter.Spec) {
		p.log.V(3).Info("New provider spec is different from the new spec")
		return true, nil
	}

	machineConfigsSpecChanged, err := p.machineConfigsSpecChanged(ctx, cc, cluster, newSpec)
	if err != nil {
		return false, err
	}
	return machineConfigsSpecChanged, nil
}

func (p *cloudstackProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range clusterSpec.CloudStackMachineConfigs {
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

func (p *cloudstackProvider) PostMoveManagementToBootstrap(_ context.Context, _ *types.Cluster) error {
	// NOOP
	return nil
}

func (p *cloudstackProvider) validateMachineConfigsNameUniqueness(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName())
	if err != nil {
		return err
	}

	cpMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	if prevSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name != cpMachineConfigName {
		err := p.validateMachineConfigNameUniqueness(ctx, cpMachineConfigName, cluster, clusterSpec)
		if err != nil {
			return err
		}
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil && prevSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		if prevSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name != etcdMachineConfigName {
			err := p.validateMachineConfigNameUniqueness(ctx, etcdMachineConfigName, cluster, clusterSpec)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *cloudstackProvider) validateMachineConfigNameUniqueness(ctx context.Context, machineConfigName string, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, machineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace())
	if err != nil {
		return err
	}
	if len(em) > 0 {
		return fmt.Errorf("machineconfig %s already exists", machineConfigName)
	}
	return nil
}

func (p *cloudstackProvider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	kubeVipDisabledString := strconv.FormatBool(features.IsActive(features.CloudStackKubeVipDisabled()))
	return p.providerKubectlClient.SetEksaControllerEnvVar(ctx, features.CloudStackKubeVipDisabledEnvVar, kubeVipDisabledString, kubeconfigFile)
}

func machineDeploymentName(clusterName, nodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, nodeGroupName)
}

// PreCoreComponentsUpgrade staisfies the Provider interface.
func (p *cloudstackProvider) PreCoreComponentsUpgrade(
	ctx context.Context,
	cluster *types.Cluster,
	managementComponents *cluster.ManagementComponents,
	clusterSpec *cluster.Spec,
) error {
	return nil
}
