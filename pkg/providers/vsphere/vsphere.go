package vsphere

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	CredentialsObjectName    = "vsphere-credentials"
	EksavSphereUsernameKey   = "EKSA_VSPHERE_USERNAME"
	EksavSpherePasswordKey   = "EKSA_VSPHERE_PASSWORD"
	eksaLicense              = "EKSA_LICENSE"
	vSphereUsernameKey       = "VSPHERE_USERNAME"
	vSpherePasswordKey       = "VSPHERE_PASSWORD"
	vSphereServerKey         = "VSPHERE_SERVER"
	govcInsecure             = "GOVC_INSECURE"
	expClusterResourceSetKey = "EXP_CLUSTER_RESOURCE_SET"
	privateKeyFileName       = "eks-a-id_rsa"
	publicKeyFileName        = "eks-a-id_rsa.pub"
	defaultTemplateLibrary   = "eks-a-templates"
	defaultTemplatesFolder   = "vm/Templates"
	bottlerocketDefaultUser  = "ec2-user"
	ubuntuDefaultUser        = "capv"
	maxRetries               = 30
	backOffPeriod            = 5 * time.Second
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

//go:embed config/secret.yaml
var defaultSecretObject string

//go:embed config/defaultStorageClass.yaml
var defaultStorageClass []byte

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

var (
	eksaVSphereDatacenterResourceType = fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereMachineResourceType    = fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group)
	noProxyDefaults                   = []string{
		"localhost",
		"127.0.0.1",
		".svc",
	}
)

var requiredEnvs = []string{vSphereUsernameKey, vSpherePasswordKey, expClusterResourceSetKey}

type vsphereProvider struct {
	datacenterConfig       *v1alpha1.VSphereDatacenterConfig
	machineConfigs         map[string]*v1alpha1.VSphereMachineConfig
	clusterConfig          *v1alpha1.Cluster
	providerGovcClient     ProviderGovcClient
	providerKubectlClient  ProviderKubectlClient
	writer                 filewriter.FileWriter
	controlPlaneSshAuthKey string
	workerSshAuthKey       string
	etcdSshAuthKey         string
	templateBuilder        *VsphereTemplateBuilder
	skipIpCheck            bool
	resourceSetManager     ClusterResourceSetManager
	Retrier                *retrier.Retrier
	validator              *Validator
	defaulter              *Defaulter
}

type ProviderGovcClient interface {
	SearchTemplate(ctx context.Context, datacenter string, machineConfig *v1alpha1.VSphereMachineConfig) (string, error)
	LibraryElementExists(ctx context.Context, library string) (bool, error)
	GetLibraryElementContentVersion(ctx context.Context, element string) (string, error)
	DeleteLibraryElement(ctx context.Context, element string) error
	TemplateHasSnapshot(ctx context.Context, template string) (bool, error)
	GetWorkloadAvailableSpace(ctx context.Context, datastore string) (float64, error)
	ValidateVCenterSetupMachineConfig(ctx context.Context, datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfig *v1alpha1.VSphereMachineConfig, selfSigned *bool) error
	ValidateVCenterConnection(ctx context.Context, server string) error
	ValidateVCenterAuthentication(ctx context.Context) error
	IsCertSelfSigned(ctx context.Context) bool
	GetCertThumbprint(ctx context.Context) (string, error)
	ConfigureCertThumbprint(ctx context.Context, server, thumbprint string) error
	DatacenterExists(ctx context.Context, datacenter string) (bool, error)
	NetworkExists(ctx context.Context, network string) (bool, error)
	CreateLibrary(ctx context.Context, datastore, library string) error
	DeployTemplateFromLibrary(ctx context.Context, templateDir, templateName, library, datacenter, datastore, resourcePool string, resizeDisk2 bool) error
	ImportTemplate(ctx context.Context, library, ovaURL, name string) error
	GetTags(ctx context.Context, path string) (tags []string, err error)
	ListTags(ctx context.Context) ([]string, error)
	CreateTag(ctx context.Context, tag, category string) error
	AddTag(ctx context.Context, path, tag string) error
	ListCategories(ctx context.Context) ([]string, error)
	CreateCategoryForVM(ctx context.Context, name string) error
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetNamespace(ctx context.Context, kubeconfig string, namespace string) error
	CreateNamespace(ctx context.Context, kubeconfig string, namespace string) error
	LoadSecret(ctx context.Context, secretObject string, secretObjType string, secretObjectName string, kubeConfFile string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
	GetEksaVSphereMachineConfig(ctx context.Context, vsphereMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereMachineConfig, error)
	GetMachineDeployment(ctx context.Context, machineDeploymentName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*controlplanev1.KubeadmControlPlane, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1.EtcdadmCluster, error)
	GetSecret(ctx context.Context, secretObjectName string, opts ...executables.KubectlOpt) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
	SearchVsphereMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.VSphereMachineConfig, error)
	SearchVsphereDatacenterConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.VSphereDatacenterConfig, error)
	SetDaemonSetImage(ctx context.Context, kubeconfigFile, name, namespace, container, image string) error
	DeleteEksaDatacenterConfig(ctx context.Context, vsphereDatacenterResourceType string, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaMachineConfig(ctx context.Context, vsphereMachineResourceType string, vsphereMachineConfigName string, kubeconfigFile string, namespace string) error
	ApplyTolerationsFromTaintsToDaemonSet(ctx context.Context, oldTaints []corev1.Taint, newTaints []corev1.Taint, dsName string, kubeconfigFile string) error
}

type ClusterResourceSetManager interface {
	ForceUpdate(ctx context.Context, name, namespace string, managementCluster, workloadCluster *types.Cluster) error
}

func NewProvider(datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfigs map[string]*v1alpha1.VSphereMachineConfig, clusterConfig *v1alpha1.Cluster, providerGovcClient ProviderGovcClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, now types.NowFunc, skipIpCheck bool, resourceSetManager ClusterResourceSetManager) *vsphereProvider {
	return NewProviderCustomNet(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		providerGovcClient,
		providerKubectlClient,
		writer,
		&networkutils.DefaultNetClient{},
		now,
		skipIpCheck,
		resourceSetManager,
	)
}

func NewProviderCustomNet(datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfigs map[string]*v1alpha1.VSphereMachineConfig, clusterConfig *v1alpha1.Cluster, providerGovcClient ProviderGovcClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, netClient networkutils.NetClient, now types.NowFunc, skipIpCheck bool, resourceSetManager ClusterResourceSetManager) *vsphereProvider {
	var controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.VSphereMachineConfigSpec
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.VSphereMachineConfigSpec, len(machineConfigs))

	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}
	retrier := retrier.NewWithMaxRetries(maxRetries, backOffPeriod)
	return &vsphereProvider{
		datacenterConfig:      datacenterConfig,
		machineConfigs:        machineConfigs,
		clusterConfig:         clusterConfig,
		providerGovcClient:    providerGovcClient,
		providerKubectlClient: providerKubectlClient,
		writer:                writer,
		templateBuilder: &VsphereTemplateBuilder{
			datacenterSpec:              &datacenterConfig.Spec,
			controlPlaneMachineSpec:     controlPlaneMachineSpec,
			WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
			etcdMachineSpec:             etcdMachineSpec,
			now:                         now,
		},
		skipIpCheck:        skipIpCheck,
		resourceSetManager: resourceSetManager,
		Retrier:            retrier,
		validator:          NewValidator(providerGovcClient, netClient),
		defaulter:          NewDefaulter(providerGovcClient),
	}
}

func (p *vsphereProvider) UpdateKubeConfig(_ *[]byte, _ string) error {
	// customize generated kube config
	return nil
}

func (p *vsphereProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, p.datacenterConfig.Spec.Server)
		for _, s := range p.clusterConfig.Spec.ProxyConfiguration.NoProxy {
			if s != "" {
				noProxy += "," + s
			}
		}
		env["HTTP_PROXY"] = p.clusterConfig.Spec.ProxyConfiguration.HttpProxy
		env["HTTPS_PROXY"] = p.clusterConfig.Spec.ProxyConfiguration.HttpsProxy
		env["NO_PROXY"] = noProxy
	}
	return []bootstrapper.BootstrapClusterOption{bootstrapper.WithEnv(env)}, nil
}

func (p *vsphereProvider) Name() string {
	return constants.VSphereProviderName
}

func (p *vsphereProvider) DatacenterResourceType() string {
	return eksaVSphereDatacenterResourceType
}

func (p *vsphereProvider) MachineResourceType() string {
	return eksaVSphereMachineResourceType
}

func (p *vsphereProvider) setupSSHAuthKeysForCreate() error {
	var useKeyGeneratedForControlplane, useKeyGeneratedForWorker bool
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
	p.controlPlaneSshAuthKey = controlPlaneUser.SshAuthorizedKeys[0]
	if err := p.parseSSHAuthKey(&p.controlPlaneSshAuthKey); err != nil {
		return err
	}
	if len(p.controlPlaneSshAuthKey) <= 0 {
		generatedKey, err := p.generateSSHAuthKey(controlPlaneUser.Name)
		if err != nil {
			return err
		}
		p.controlPlaneSshAuthKey = generatedKey
		useKeyGeneratedForControlplane = true
	}
	controlPlaneUser.SshAuthorizedKeys[0] = p.controlPlaneSshAuthKey
	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		workerUser := p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.workerSshAuthKey = workerUser.SshAuthorizedKeys[0]
		if err := p.parseSSHAuthKey(&p.workerSshAuthKey); err != nil {
			return err
		}
		if len(p.workerSshAuthKey) <= 0 {
			if useKeyGeneratedForControlplane { // use the same key
				p.workerSshAuthKey = p.controlPlaneSshAuthKey
			} else {
				generatedKey, err := p.generateSSHAuthKey(workerUser.Name)
				if err != nil {
					return err
				}
				p.workerSshAuthKey = generatedKey
				useKeyGeneratedForWorker = true
			}
		}
		workerUser.SshAuthorizedKeys[0] = p.workerSshAuthKey
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.etcdSshAuthKey = etcdUser.SshAuthorizedKeys[0]
		if err := p.parseSSHAuthKey(&p.etcdSshAuthKey); err != nil {
			return err
		}
		if len(p.etcdSshAuthKey) <= 0 {
			if useKeyGeneratedForControlplane { // use the same key as for controlplane
				p.etcdSshAuthKey = p.controlPlaneSshAuthKey
			} else if useKeyGeneratedForWorker {
				p.etcdSshAuthKey = p.workerSshAuthKey // if cp key was provided by user, check if worker key was generated by cli and use that
			} else {
				generatedKey, err := p.generateSSHAuthKey(etcdUser.Name)
				if err != nil {
					return err
				}
				p.etcdSshAuthKey = generatedKey
			}
		}
		etcdUser.SshAuthorizedKeys[0] = p.etcdSshAuthKey
	}
	return nil
}

func (p *vsphereProvider) setupSSHAuthKeysForUpgrade() error {
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
	p.controlPlaneSshAuthKey = controlPlaneUser.SshAuthorizedKeys[0]
	if err := p.parseSSHAuthKey(&p.controlPlaneSshAuthKey); err != nil {
		return err
	}
	controlPlaneUser.SshAuthorizedKeys[0] = p.controlPlaneSshAuthKey
	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		workerUser := p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.workerSshAuthKey = workerUser.SshAuthorizedKeys[0]
		if err := p.parseSSHAuthKey(&p.workerSshAuthKey); err != nil {
			return err
		}
		workerUser.SshAuthorizedKeys[0] = p.workerSshAuthKey
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.etcdSshAuthKey = etcdUser.SshAuthorizedKeys[0]
		if err := p.parseSSHAuthKey(&p.etcdSshAuthKey); err != nil {
			return err
		}
		etcdUser.SshAuthorizedKeys[0] = p.etcdSshAuthKey
	}
	return nil
}

func (p *vsphereProvider) parseSSHAuthKey(key *string) error {
	if len(*key) > 0 {
		// When public key is entered by user in provider config, it may contain email address (or any other comment) at the end. ssh-keygen allows users to add comments as suffixes to public key in
		// public key file. When CLI generates the key pair, no comments will be present. So we get rid of the comment from the public key to ensure unit tests that do string compare on the sshAuthorizedKey
		// will pass
		parts := strings.Fields(strings.TrimSpace(*key))
		if len(parts) >= 3 {
			*key = parts[0] + " " + parts[1]
		}
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*key))
		if err != nil {
			return fmt.Errorf("provided VSphereMachineConfig sshAuthorizedKey is invalid: %v", err)
		}
	}
	return nil
}

func (p *vsphereProvider) generateSSHAuthKey(username string) (string, error) {
	logger.Info("Provided VSphereMachineConfig sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
	keygenerator, _ := crypto.NewKeyGenerator(p.writer)
	sshAuthorizedKeyBytes, err := keygenerator.GenerateSSHKeyPair("", "", privateKeyFileName, publicKeyFileName, username)
	if err != nil || sshAuthorizedKeyBytes == nil {
		return "", fmt.Errorf("VSphereMachineConfig error generating sshAuthorizedKey: %v", err)
	}
	key := string(sshAuthorizedKeyBytes)
	key = strings.TrimRight(key, "\n")
	return key, nil
}

func (p *vsphereProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range p.machineConfigs {
		if err := p.providerKubectlClient.DeleteEksaMachineConfig(ctx, eksaVSphereMachineResourceType, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaDatacenterConfig(ctx, eksaVSphereDatacenterResourceType, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, p.datacenterConfig.Namespace)
}

func (p *vsphereProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := SetupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	vSphereClusterSpec := NewSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	if err := p.defaulter.SetDefaultsForDatacenterConfig(ctx, vSphereClusterSpec.datacenterConfig); err != nil {
		return fmt.Errorf("failed setting default values for vsphere datacenter config: %v", err)
	}

	if err := vSphereClusterSpec.datacenterConfig.ValidateFields(); err != nil {
		return err
	}

	if err := p.validator.ValidateVCenterConfig(ctx, vSphereClusterSpec.datacenterConfig); err != nil {
		return err
	}

	if err := p.defaulter.setDefaultsForMachineConfig(ctx, vSphereClusterSpec); err != nil {
		return fmt.Errorf("failed setting default values for vsphere machine configs: %v", err)
	}

	if err := p.validator.ValidateClusterMachineConfigs(ctx, vSphereClusterSpec); err != nil {
		return err
	}

	if err := p.setupSSHAuthKeysForCreate(); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	// TODO: move this to validator
	if clusterSpec.IsManaged() {
		for _, mc := range p.MachineConfigs() {
			em, err := p.providerKubectlClient.SearchVsphereMachineConfig(ctx, mc.GetName(), clusterSpec.ManagementCluster.KubeconfigFile, mc.GetNamespace())
			if err != nil {
				return err
			}
			if len(em) > 0 {
				return fmt.Errorf("VSphereMachineConfig %s already exists", mc.GetName())
			}
		}
		existingDatacenter, err := p.providerKubectlClient.SearchVsphereDatacenterConfig(ctx, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Namespace)
		if err != nil {
			return err
		}
		if len(existingDatacenter) > 0 {
			return fmt.Errorf("VSphereDatacenter %s already exists", p.datacenterConfig.Name)
		}
	}

	if !p.skipIpCheck {
		if err := p.validator.validateControlPlaneIpUniqueness(vSphereClusterSpec); err != nil {
			return err
		}
	} else {
		logger.Info("Skipping check for whether control plane ip is in use")
	}

	return nil
}

func (p *vsphereProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if err := SetupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	vSphereClusterSpec := NewSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	if err := p.defaulter.SetDefaultsForDatacenterConfig(ctx, vSphereClusterSpec.datacenterConfig); err != nil {
		return fmt.Errorf("failed setting default values for vsphere datacenter config: %v", err)
	}

	if err := vSphereClusterSpec.datacenterConfig.ValidateFields(); err != nil {
		return err
	}

	if err := p.validator.ValidateVCenterConfig(ctx, vSphereClusterSpec.datacenterConfig); err != nil {
		return err
	}

	if err := p.defaulter.setDefaultsForMachineConfig(ctx, vSphereClusterSpec); err != nil {
		return fmt.Errorf("failed setting default values for vsphere machine configs: %v", err)
	}

	if err := p.validator.ValidateClusterMachineConfigs(ctx, vSphereClusterSpec); err != nil {
		return err
	}

	err := p.setupSSHAuthKeysForUpgrade()
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	err = p.validateMachineConfigsNameUniqueness(ctx, cluster, clusterSpec)
	if err != nil {
		return fmt.Errorf("failed validate machineconfig uniqueness: %v", err)
	}
	return nil
}

func (p *vsphereProvider) validateMachineConfigsNameUniqueness(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.GetName())
	if err != nil {
		return err
	}

	cpMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	if prevSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name != cpMachineConfigName {
		em, err := p.providerKubectlClient.SearchVsphereMachineConfig(ctx, cpMachineConfigName, cluster.KubeconfigFile, clusterSpec.GetNamespace())
		if err != nil {
			return err
		}
		if len(em) > 0 {
			return fmt.Errorf("control plane VSphereMachineConfig %s already exists", cpMachineConfigName)
		}
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil && prevSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		if prevSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name != etcdMachineConfigName {
			em, err := p.providerKubectlClient.SearchVsphereMachineConfig(ctx, etcdMachineConfigName, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.GetNamespace())
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

func (p *vsphereProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	var contents bytes.Buffer
	err := p.createSecret(ctx, cluster, &contents)
	if err != nil {
		return err
	}

	err = p.providerKubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, contents.Bytes())
	if err != nil {
		return fmt.Errorf("error loading secrets object: %v", err)
	}
	return nil
}

func (p *vsphereProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	if err := SetupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	return nil
}

func NeedsNewControlPlaneTemplate(oldSpec, newSpec *cluster.Spec, oldVdc, newVdc *v1alpha1.VSphereDatacenterConfig, oldVmc, newVmc *v1alpha1.VSphereMachineConfig) bool {
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
	return AnyImmutableFieldChanged(oldVdc, newVdc, oldVmc, newVmc)
}

func NeedsNewWorkloadTemplate(oldSpec, newSpec *cluster.Spec, oldVdc, newVdc *v1alpha1.VSphereDatacenterConfig, oldVmc, newVmc *v1alpha1.VSphereMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	if !v1alpha1.WorkerNodeGroupConfigurationSliceTaintsEqual(oldSpec.Spec.WorkerNodeGroupConfigurations, newSpec.Spec.WorkerNodeGroupConfigurations) ||
		!v1alpha1.WorkerNodeGroupConfigurationsLabelsMapEqual(oldSpec.Spec.WorkerNodeGroupConfigurations, newSpec.Spec.WorkerNodeGroupConfigurations) {
		return true
	}
	return AnyImmutableFieldChanged(oldVdc, newVdc, oldVmc, newVmc)
}

func NeedsNewKubeadmConfigTemplate(newWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration, oldWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration) bool {
	return !v1alpha1.TaintsSliceEqual(newWorkerNodeGroup.Taints, oldWorkerNodeGroup.Taints) || !v1alpha1.LabelsMapEqual(newWorkerNodeGroup.Labels, oldWorkerNodeGroup.Labels)
}

func NeedsNewEtcdTemplate(oldSpec, newSpec *cluster.Spec, oldVdc, newVdc *v1alpha1.VSphereDatacenterConfig, oldVmc, newVmc *v1alpha1.VSphereMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldVdc, newVdc, oldVmc, newVmc)
}

func AnyImmutableFieldChanged(oldVdc, newVdc *v1alpha1.VSphereDatacenterConfig, oldVmc, newVmc *v1alpha1.VSphereMachineConfig) bool {
	if oldVmc.Spec.NumCPUs != newVmc.Spec.NumCPUs {
		return true
	}
	if oldVmc.Spec.MemoryMiB != newVmc.Spec.MemoryMiB {
		return true
	}
	if oldVmc.Spec.DiskGiB != newVmc.Spec.DiskGiB {
		return true
	}
	if oldVmc.Spec.Datastore != newVmc.Spec.Datastore {
		return true
	}
	if oldVmc.Spec.Folder != newVmc.Spec.Folder {
		return true
	}
	if oldVdc.Spec.Network != newVdc.Spec.Network {
		return true
	}
	if oldVmc.Spec.ResourcePool != newVmc.Spec.ResourcePool {
		return true
	}
	if oldVdc.Spec.Thumbprint != newVdc.Spec.Thumbprint {
		return true
	}
	if oldVmc.Spec.Template != newVmc.Spec.Template {
		return true
	}
	return false
}

func NewVsphereTemplateBuilder(datacenterSpec *v1alpha1.VSphereDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.VSphereMachineConfigSpec, workerNodeGroupMachineSpecs map[string]v1alpha1.VSphereMachineConfigSpec, now types.NowFunc, fromController bool) *VsphereTemplateBuilder {
	return &VsphereTemplateBuilder{
		datacenterSpec:              datacenterSpec,
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		etcdMachineSpec:             etcdMachineSpec,
		now:                         now,
		fromController:              fromController,
	}
}

type VsphereTemplateBuilder struct {
	datacenterSpec              *v1alpha1.VSphereDatacenterConfigSpec
	controlPlaneMachineSpec     *v1alpha1.VSphereMachineConfigSpec
	WorkerNodeGroupMachineSpecs map[string]v1alpha1.VSphereMachineConfigSpec
	etcdMachineSpec             *v1alpha1.VSphereMachineConfigSpec
	now                         types.NowFunc
	fromController              bool
}

func (vs *VsphereTemplateBuilder) WorkerMachineTemplateName(clusterName, workerNodeGroupName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%s-%d", clusterName, workerNodeGroupName, t)
}

func (vs *VsphereTemplateBuilder) CPMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-control-plane-template-%d", clusterName, t)
}

func (vs *VsphereTemplateBuilder) EtcdMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-etcd-template-%d", clusterName, t)
}

func (vs *VsphereTemplateBuilder) KubeadmConfigTemplateName(clusterName, workerNodeGroupName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%s-template-%d", clusterName, workerNodeGroupName, t)
}

func (vs *VsphereTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	var etcdMachineSpec v1alpha1.VSphereMachineConfigSpec
	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *vs.etcdMachineSpec
	}
	values := buildTemplateMapCP(clusterSpec, *vs.datacenterSpec, *vs.controlPlaneMachineSpec, etcdMachineSpec)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (vs *VsphereTemplateBuilder) isCgroupDriverSystemd(clusterSpec *cluster.Spec) (bool, error) {
	bundle := clusterSpec.VersionsBundle
	k8sVersion, err := semver.New(bundle.KubeDistro.Kubernetes.Tag)
	if err != nil {
		return false, fmt.Errorf("error parsing kubernetes version %v: %v", bundle.KubeDistro.Kubernetes.Tag, err)
	}
	if vs.fromController && k8sVersion.Major == 1 && k8sVersion.Minor >= 21 {
		return true, nil
	}
	return false, nil
}

func (vs *VsphereTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	// pin cgroupDriver to systemd for k8s >= 1.21 when generating template in controller
	// remove this check once the controller supports order upgrade.
	// i.e. control plane, etcd upgrade before worker nodes.
	cgroupDriverSystemd, err := vs.isCgroupDriverSystemd(clusterSpec)
	if err != nil {
		return nil, err
	}

	workerSpecs := make([][]byte, 0, len(clusterSpec.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Spec.WorkerNodeGroupConfigurations {
		values := buildTemplateMapMD(clusterSpec, *vs.datacenterSpec, vs.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration)
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]

		values["cgroupDriverSystemd"] = cgroupDriverSystemd

		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

func buildTemplateMapCP(clusterSpec *cluster.Spec, datacenterSpec v1alpha1.VSphereDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec v1alpha1.VSphereMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	etcdExtraArgs := clusterapi.SecureEtcdTlsCipherSuitesExtraArgs()
	sharedExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs()
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Spec.ClusterNetwork.DNS.ResolvConf)).
		Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(clusterSpec.Spec.ControlPlaneConfiguration))
	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).
		Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).
		Append(clusterapi.PodIAMAuthExtraArgs(clusterSpec.Spec.PodIAMConfig)).
		Append(sharedExtraArgs)

	values := map[string]interface{}{
		"clusterName":                          clusterSpec.ObjectMeta.Name,
		"controlPlaneEndpointIp":               clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":                 clusterSpec.Spec.ControlPlaneConfiguration.Count,
		"kubernetesRepository":                 bundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":                    bundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                       bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                         bundle.KubeDistro.Etcd.Tag,
		"corednsRepository":                    bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                       bundle.KubeDistro.CoreDNS.Tag,
		"nodeDriverRegistrarImage":             bundle.KubeDistro.NodeDriverRegistrar.VersionedImage(),
		"livenessProbeImage":                   bundle.KubeDistro.LivenessProbe.VersionedImage(),
		"externalAttacherImage":                bundle.KubeDistro.ExternalAttacher.VersionedImage(),
		"externalProvisionerImage":             bundle.KubeDistro.ExternalProvisioner.VersionedImage(),
		"thumbprint":                           datacenterSpec.Thumbprint,
		"vsphereDatacenter":                    datacenterSpec.Datacenter,
		"controlPlaneVsphereDatastore":         controlPlaneMachineSpec.Datastore,
		"controlPlaneVsphereFolder":            controlPlaneMachineSpec.Folder,
		"managerImage":                         bundle.VSphere.Manager.VersionedImage(),
		"kubeVipImage":                         bundle.VSphere.KubeVip.VersionedImage(),
		"driverImage":                          bundle.VSphere.Driver.VersionedImage(),
		"syncerImage":                          bundle.VSphere.Syncer.VersionedImage(),
		"insecure":                             datacenterSpec.Insecure,
		"vsphereNetwork":                       datacenterSpec.Network,
		"controlPlaneVsphereResourcePool":      controlPlaneMachineSpec.ResourcePool,
		"vsphereServer":                        datacenterSpec.Server,
		"controlPlaneVsphereStoragePolicyName": controlPlaneMachineSpec.StoragePolicyName,
		"vsphereTemplate":                      controlPlaneMachineSpec.Template,
		"controlPlaneVMsMemoryMiB":             controlPlaneMachineSpec.MemoryMiB,
		"controlPlaneVMsNumCPUs":               controlPlaneMachineSpec.NumCPUs,
		"controlPlaneDiskGiB":                  controlPlaneMachineSpec.DiskGiB,
		"controlPlaneSshUsername":              controlPlaneMachineSpec.Users[0].Name,
		"podCidrs":                             clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                         clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks,
		"etcdExtraArgs":                        etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":                     crypto.SecureCipherSuitesString(),
		"apiserverExtraArgs":                   apiServerExtraArgs.ToPartialYaml(),
		"controllermanagerExtraArgs":           sharedExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":                   sharedExtraArgs.ToPartialYaml(),
		"kubeletExtraArgs":                     kubeletExtraArgs.ToPartialYaml(),
		"format":                               format,
		"externalEtcdVersion":                  bundle.KubeDistro.EtcdVersion,
		"etcdImage":                            bundle.KubeDistro.EtcdImage.VersionedImage(),
		"eksaSystemNamespace":                  constants.EksaSystemNamespace,
		"auditPolicy":                          common.GetAuditPolicy(),
		"resourceSetName":                      resourceSetName(clusterSpec),
		"eksaVsphereUsername":                  os.Getenv(EksavSphereUsernameKey),
		"eksaVspherePassword":                  os.Getenv(EksavSpherePasswordKey),
	}

	if clusterSpec.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = net.JoinHostPort(clusterSpec.Spec.RegistryMirrorConfiguration.Endpoint, clusterSpec.Spec.RegistryMirrorConfiguration.Port)
		if len(clusterSpec.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Spec.RegistryMirrorConfiguration.CACertContent
		}
	}

	if clusterSpec.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		capacity := len(clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks) +
			len(clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks) +
			len(clusterSpec.Spec.ProxyConfiguration.NoProxy) + 4
		noProxyList := make([]string, 0, capacity)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ProxyConfiguration.NoProxy...)

		// Add no-proxy defaults
		noProxyList = append(noProxyList, noProxyDefaults...)
		noProxyList = append(noProxyList,
			datacenterSpec.Server,
			clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Spec.ExternalEtcdConfiguration.Count
		values["etcdVsphereDatastore"] = etcdMachineSpec.Datastore
		values["etcdVsphereFolder"] = etcdMachineSpec.Folder
		values["etcdDiskGiB"] = etcdMachineSpec.DiskGiB
		values["etcdVMsMemoryMiB"] = etcdMachineSpec.MemoryMiB
		values["etcdVMsNumCPUs"] = etcdMachineSpec.NumCPUs
		values["etcdVsphereResourcePool"] = etcdMachineSpec.ResourcePool
		values["etcdVsphereStoragePolicyName"] = etcdMachineSpec.StoragePolicyName
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
	}

	if controlPlaneMachineSpec.OSFamily == v1alpha1.Bottlerocket {
		values["format"] = string(v1alpha1.Bottlerocket)
		values["pauseRepository"] = bundle.KubeDistro.Pause.Image()
		values["pauseVersion"] = bundle.KubeDistro.Pause.Tag()
		values["bottlerocketBootstrapRepository"] = bundle.BottleRocketBootstrap.Bootstrap.Image()
		values["bottlerocketBootstrapVersion"] = bundle.BottleRocketBootstrap.Bootstrap.Tag()
	}

	if len(clusterSpec.Spec.ControlPlaneConfiguration.Taints) > 0 {
		values["controlPlaneTaints"] = clusterSpec.Spec.ControlPlaneConfiguration.Taints
	}

	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}

	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, datacenterSpec v1alpha1.VSphereDatacenterConfigSpec, workerNodeGroupMachineSpec v1alpha1.VSphereMachineConfigSpec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Spec.ClusterNetwork.DNS.ResolvConf))

	values := map[string]interface{}{
		"clusterName":                    clusterSpec.ObjectMeta.Name,
		"kubernetesVersion":              bundle.KubeDistro.Kubernetes.Tag,
		"thumbprint":                     datacenterSpec.Thumbprint,
		"vsphereDatacenter":              datacenterSpec.Datacenter,
		"workerVsphereDatastore":         workerNodeGroupMachineSpec.Datastore,
		"workerVsphereFolder":            workerNodeGroupMachineSpec.Folder,
		"vsphereNetwork":                 datacenterSpec.Network,
		"workerVsphereResourcePool":      workerNodeGroupMachineSpec.ResourcePool,
		"vsphereServer":                  datacenterSpec.Server,
		"workerVsphereStoragePolicyName": workerNodeGroupMachineSpec.StoragePolicyName,
		"vsphereTemplate":                workerNodeGroupMachineSpec.Template,
		"workloadVMsMemoryMiB":           workerNodeGroupMachineSpec.MemoryMiB,
		"workloadVMsNumCPUs":             workerNodeGroupMachineSpec.NumCPUs,
		"workloadDiskGiB":                workerNodeGroupMachineSpec.DiskGiB,
		"workerSshUsername":              workerNodeGroupMachineSpec.Users[0].Name,
		"format":                         format,
		"eksaSystemNamespace":            constants.EksaSystemNamespace,
		"kubeletExtraArgs":               kubeletExtraArgs.ToPartialYaml(),
		"vsphereWorkerSshAuthorizedKey":  workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys[0],
		"workerReplicas":                 workerNodeGroupConfiguration.Count,
		"workerNodeGroupName":            fmt.Sprintf("%s-%s", clusterSpec.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints":          workerNodeGroupConfiguration.Taints,
	}

	if clusterSpec.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = net.JoinHostPort(clusterSpec.Spec.RegistryMirrorConfiguration.Endpoint, clusterSpec.Spec.RegistryMirrorConfiguration.Port)
		if len(clusterSpec.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Spec.RegistryMirrorConfiguration.CACertContent
		}
	}

	if clusterSpec.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		capacity := len(clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks) +
			len(clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks) +
			len(clusterSpec.Spec.ProxyConfiguration.NoProxy) + 4
		noProxyList := make([]string, 0, capacity)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ProxyConfiguration.NoProxy...)

		// Add no-proxy defaults
		noProxyList = append(noProxyList, noProxyDefaults...)
		noProxyList = append(noProxyList,
			datacenterSpec.Server,
			clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	if workerNodeGroupMachineSpec.OSFamily == v1alpha1.Bottlerocket {
		values["format"] = string(v1alpha1.Bottlerocket)
		values["pauseRepository"] = bundle.KubeDistro.Pause.Image()
		values["pauseVersion"] = bundle.KubeDistro.Pause.Tag()
		values["bottlerocketBootstrapRepository"] = bundle.BottleRocketBootstrap.Bootstrap.Image()
		values["bottlerocketBootstrapVersion"] = bundle.BottleRocketBootstrap.Bootstrap.Tag()
	}

	return values
}

func (p *vsphereProvider) generateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := newClusterSpec.ObjectMeta.Name
	var controlPlaneTemplateName, workloadTemplateName, kubeadmconfigTemplateName, etcdTemplateName string
	var needsNewEtcdTemplate bool

	c, err := p.providerKubectlClient.GetEksaCluster(ctx, workloadCluster, newClusterSpec.Name)
	if err != nil {
		return nil, nil, err
	}
	vdc, err := p.providerKubectlClient.GetEksaVSphereDatacenterConfig(ctx, p.datacenterConfig.Name, workloadCluster.KubeconfigFile, newClusterSpec.Namespace)
	if err != nil {
		return nil, nil, err
	}
	controlPlaneMachineConfig := p.machineConfigs[newClusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	controlPlaneVmc, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, c.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Namespace)
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
		controlPlaneTemplateName = p.templateBuilder.CPMachineTemplateName(clusterName)
	}

	previousWorkerNodeGroupConfigs := cluster.BuildMapForWorkerNodeGroupsByName(currentSpec.Spec.WorkerNodeGroupConfigurations)

	workloadTemplateNames := make(map[string]string, len(newClusterSpec.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(newClusterSpec.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range newClusterSpec.Spec.WorkerNodeGroupConfigurations {
		needsNewWorkloadTemplate, err := p.needsNewMachineTemplate(ctx, workloadCluster, currentSpec, newClusterSpec, workerNodeGroupConfiguration, vdc, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, nil, err
		}

		needsNewKubeadmConfigTemplate, err := p.needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, nil, err
		}
		if !needsNewKubeadmConfigTemplate {
			mdName := machineDeploymentName(newClusterSpec.Name, workerNodeGroupConfiguration.Name)
			md, err := p.providerKubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			kubeadmconfigTemplateName = md.Spec.Template.Spec.Bootstrap.ConfigRef.Name
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = kubeadmconfigTemplateName
		} else {
			kubeadmconfigTemplateName = p.templateBuilder.KubeadmConfigTemplateName(clusterName, workerNodeGroupConfiguration.Name)
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = kubeadmconfigTemplateName
		}

		if !needsNewWorkloadTemplate {
			mdName := machineDeploymentName(newClusterSpec.Name, workerNodeGroupConfiguration.Name)
			md, err := p.providerKubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			workloadTemplateName = md.Spec.Template.Spec.InfrastructureRef.Name
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		} else {
			workloadTemplateName = p.templateBuilder.WorkerMachineTemplateName(clusterName, workerNodeGroupConfiguration.Name)
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		}
		p.templateBuilder.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}

	if newClusterSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := p.machineConfigs[newClusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		etcdMachineVmc, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, c.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Namespace)
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
				map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"},
				executables.WithCluster(bootstrapCluster),
				executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			etcdTemplateName = p.templateBuilder.EtcdMachineTemplateName(clusterName)
		}
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["vsphereControlPlaneSshAuthorizedKey"] = p.controlPlaneSshAuthKey
		values["vsphereEtcdSshAuthorizedKey"] = p.etcdSshAuthKey
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

func (p *vsphereProvider) generateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := clusterSpec.ObjectMeta.Name

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = p.templateBuilder.CPMachineTemplateName(clusterName)
		values["vsphereControlPlaneSshAuthorizedKey"] = p.controlPlaneSshAuthKey
		values["vsphereEtcdSshAuthorizedKey"] = p.etcdSshAuthKey
		values["etcdTemplateName"] = p.templateBuilder.EtcdMachineTemplateName(clusterName)
	}
	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	if err != nil {
		return nil, nil, err
	}

	workloadTemplateNames := make(map[string]string, len(clusterSpec.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(clusterSpec.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Spec.WorkerNodeGroupConfigurations {
		workloadTemplateNames[workerNodeGroupConfiguration.Name] = p.templateBuilder.WorkerMachineTemplateName(clusterSpec.Name, workerNodeGroupConfiguration.Name)
		kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = p.templateBuilder.KubeadmConfigTemplateName(clusterSpec.Name, workerNodeGroupConfiguration.Name)
		p.templateBuilder.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}
	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *vsphereProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForUpgrade(ctx, bootstrapCluster, workloadCluster, currentSpec, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating cluster api spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *vsphereProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForCreate(ctx, cluster, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating cluster api spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *vsphereProvider) GenerateStorageClass() []byte {
	return defaultStorageClass
}

func (p *vsphereProvider) GenerateMHC() ([]byte, error) {
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

func (p *vsphereProvider) createSecret(ctx context.Context, cluster *types.Cluster, contents *bytes.Buffer) error {
	if err := p.providerKubectlClient.GetNamespace(ctx, cluster.KubeconfigFile, constants.EksaSystemNamespace); err != nil {
		if err := p.providerKubectlClient.CreateNamespace(ctx, cluster.KubeconfigFile, constants.EksaSystemNamespace); err != nil {
			return err
		}
	}
	t, err := template.New("tmpl").Parse(defaultSecretObject)
	if err != nil {
		return fmt.Errorf("error creating secret object template: %v", err)
	}

	values := map[string]string{
		"vspherePassword":        os.Getenv(vSpherePasswordKey),
		"vsphereUsername":        os.Getenv(vSphereUsernameKey),
		"eksaLicense":            os.Getenv(eksaLicense),
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
		"vsphereCredentialsName": constants.VSphereCredentialsName,
		"eksaLicenseName":        constants.EksaLicenseName,
	}
	err = t.Execute(contents, values)
	if err != nil {
		return fmt.Errorf("error substituting values for secret object template: %v", err)
	}
	return nil
}

func (p *vsphereProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

func (p *vsphereProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.VSphere.Version
}

func (p *vsphereProvider) EnvMap() (map[string]string, error) {
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

func (p *vsphereProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capv-system": {"capv-controller-manager"},
	}
}

func (p *vsphereProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	bundle := clusterSpec.VersionsBundle
	folderName := fmt.Sprintf("infrastructure-vsphere/%s/", bundle.VSphere.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			bundle.VSphere.Components,
			bundle.VSphere.Metadata,
			bundle.VSphere.ClusterTemplate,
		},
	}
	return &infraBundle
}

func (p *vsphereProvider) DatacenterConfig() providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *vsphereProvider) MachineConfigs() []providers.MachineConfig {
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

func (p *vsphereProvider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Name)
	if err != nil {
		return err
	}

	prevDatacenter, err := p.providerKubectlClient.GetEksaVSphereDatacenterConfig(ctx, prevSpec.Spec.DatacenterRef.Name, cluster.KubeconfigFile, prevSpec.Namespace)
	if err != nil {
		return err
	}

	datacenter := p.datacenterConfig

	oSpec := prevDatacenter.Spec
	nSpec := datacenter.Spec

	prevMachineConfigRefs := machineRefSliceToMap(prevSpec.MachineConfigRefs())

	for _, machineConfigRef := range clusterSpec.MachineConfigRefs() {
		machineConfig, ok := p.machineConfigs[machineConfigRef.Name]
		if !ok {
			return fmt.Errorf("cannot find machine config %s in vsphere provider machine configs", machineConfigRef.Name)
		}

		if _, ok = prevMachineConfigRefs[machineConfig.Name]; ok {
			err = p.validateMachineConfigImmutability(ctx, cluster, machineConfig, clusterSpec)
			if err != nil {
				return err
			}
		}
	}

	if nSpec.Server != oSpec.Server {
		return fmt.Errorf("spec.server is immutable. Previous value %s, new value %s", oSpec.Server, nSpec.Server)
	}
	if nSpec.Datacenter != oSpec.Datacenter {
		return fmt.Errorf("spec.datacenter is immutable. Previous value %s, new value %s", oSpec.Datacenter, nSpec.Datacenter)
	}

	if nSpec.Network != oSpec.Network {
		return fmt.Errorf("spec.network is immutable. Previous value %s, new value %s", oSpec.Network, nSpec.Network)
	}

	if nSpec.Insecure != oSpec.Insecure {
		return fmt.Errorf("spec.insecure is immutable. Previous value %t, new value %t", oSpec.Insecure, nSpec.Insecure)
	}

	if nSpec.Thumbprint != oSpec.Thumbprint {
		return fmt.Errorf("spec.thumbprint is immutable. Previous value %s, new value %s", oSpec.Thumbprint, nSpec.Thumbprint)
	}

	secretChanged, err := p.secretContentsChanged(ctx, cluster)
	if err != nil {
		return err
	}

	if secretChanged {
		return fmt.Errorf("the VSphere credentials derived from %s and %s are immutable; please use the same credentials for the upgraded cluster", vSpherePasswordKey, vSphereUsernameKey)
	}
	return nil
}

func (p *vsphereProvider) needsNewMachineTemplate(ctx context.Context, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, vdc *v1alpha1.VSphereDatacenterConfig, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (bool, error) {
	if _, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		workerMachineConfig := p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		workerVmc, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, workerNodeGroupConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Namespace)
		if err != nil {
			return false, err
		}
		needsNewWorkloadTemplate := NeedsNewWorkloadTemplate(currentSpec, newClusterSpec, vdc, p.datacenterConfig, workerVmc, workerMachineConfig)
		return needsNewWorkloadTemplate, nil
	}
	return true, nil
}

func (p *vsphereProvider) needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (bool, error) {
	if _, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		existingWorkerNodeGroupConfig := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]
		return NeedsNewKubeadmConfigTemplate(&workerNodeGroupConfiguration, &existingWorkerNodeGroupConfig), nil
	}
	return true, nil
}

func (p *vsphereProvider) validateMachineConfigImmutability(ctx context.Context, cluster *types.Cluster, newConfig *v1alpha1.VSphereMachineConfig, clusterSpec *cluster.Spec) error {
	prevMachineConfig, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, newConfig.Name, cluster.KubeconfigFile, clusterSpec.Namespace)
	if err != nil {
		return err
	}

	if newConfig.Spec.StoragePolicyName != prevMachineConfig.Spec.StoragePolicyName {
		return fmt.Errorf("spec.storagePolicyName is immutable. Previous value %s, new value %s", prevMachineConfig.Spec.StoragePolicyName, newConfig.Spec.StoragePolicyName)
	}

	if !reflect.DeepEqual(newConfig.Spec.Users, prevMachineConfig.Spec.Users) {
		return fmt.Errorf("vsphereMachineConfig %s users are immutable; new user: %v; old user: %v", newConfig.Name, newConfig.Spec.Users, prevMachineConfig.Spec.Users)
	}

	return nil
}

func (p *vsphereProvider) secretContentsChanged(ctx context.Context, workloadCluster *types.Cluster) (bool, error) {
	nPassword := os.Getenv(vSpherePasswordKey)
	oSecret, err := p.providerKubectlClient.GetSecret(ctx, CredentialsObjectName, executables.WithCluster(workloadCluster), executables.WithNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return false, fmt.Errorf("error when obtaining VSphere secret %s from workload cluster: %v", CredentialsObjectName, err)
	}

	if string(oSecret.Data["password"]) != nPassword {
		return true, nil
	}

	nUser := os.Getenv(vSphereUsernameKey)
	if string(oSecret.Data["username"]) != nUser {
		return true, nil
	}
	return false, nil
}

func (p *vsphereProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	if currentSpec.VersionsBundle.VSphere.Version == newSpec.VersionsBundle.VSphere.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.VSphereProviderName,
		NewVersion:    newSpec.VersionsBundle.VSphere.Version,
		OldVersion:    currentSpec.VersionsBundle.VSphere.Version,
	}
}

func (p *vsphereProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// Use retrier so that cluster upgrade does not fail due to any intermittent failure while connecting to kube-api server

	// This is unfortunate, but ClusterResourceSet's don't support any type of reapply of the resources they manage
	// Even if we create a new ClusterResourceSet, if such resources already exist in the cluster, they won't be reapplied
	// The long term solution is to add this capability to the cluster-api controller,
	// with a new mode like "ReApplyOnChanges" or "ReApplyOnCreate" vs the current "ReApplyOnce"
	err := p.Retrier.Retry(
		func() error {
			return p.resourceSetManager.ForceUpdate(ctx, resourceSetName(clusterSpec), constants.EksaSystemNamespace, managementCluster, workloadCluster)
		},
	)
	if err != nil {
		return fmt.Errorf("failed updating the vsphere provider resource set post upgrade: %v", err)
	}
	return nil
}

func resourceSetName(clusterSpec *cluster.Spec) string {
	return fmt.Sprintf("%s-crs-0", clusterSpec.Name)
}

func (p *vsphereProvider) UpgradeNeeded(_ context.Context, newSpec, currentSpec *cluster.Spec) (bool, error) {
	newV, oldV := newSpec.VersionsBundle.VSphere, currentSpec.VersionsBundle.VSphere

	return newV.Driver.ImageDigest != oldV.Driver.ImageDigest ||
		newV.Syncer.ImageDigest != oldV.Syncer.ImageDigest ||
		newV.Manager.ImageDigest != oldV.Manager.ImageDigest ||
		newV.KubeVip.ImageDigest != oldV.KubeVip.ImageDigest, nil
}

func (p *vsphereProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	return nil
}

func configsMapToSlice(c map[string]providers.MachineConfig) []providers.MachineConfig {
	configs := make([]providers.MachineConfig, 0, len(c))
	for _, config := range c {
		configs = append(configs, config)
	}

	return configs
}

func machineRefSliceToMap(machineRefs []v1alpha1.Ref) map[string]v1alpha1.Ref {
	refMap := make(map[string]v1alpha1.Ref, len(machineRefs))
	for _, ref := range machineRefs {
		refMap[ref.Name] = ref
	}
	return refMap
}

func machineDeploymentName(clusterName, nodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, nodeGroupName)
}

func (p *vsphereProvider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	nodeGroupsToDelete := cluster.NodeGroupsToDelete(currentSpec, newSpec)
	machineDeployments := make([]string, 0, len(nodeGroupsToDelete))
	for _, group := range nodeGroupsToDelete {
		mdName := machineDeploymentName(workloadCluster.Name, group.Name)
		machineDeployments = append(machineDeployments, mdName)
	}
	return machineDeployments
}
