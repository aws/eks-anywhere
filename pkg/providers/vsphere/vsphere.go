package vsphere

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	CredentialsObjectName    = "vsphere-credentials"
	eksaLicense              = "EKSA_LICENSE"
	vSphereUsernameKey       = "VSPHERE_USERNAME"
	vSpherePasswordKey       = "VSPHERE_PASSWORD"
	vSphereServerKey         = "VSPHERE_SERVER"
	govcDatacenterKey        = "GOVC_DATACENTER"
	govcInsecure             = "GOVC_INSECURE"
	expClusterResourceSetKey = "EXP_CLUSTER_RESOURCE_SET"
	defaultTemplateLibrary   = "eks-a-templates"
	defaultTemplatesFolder   = "vm/Templates"
	maxRetries               = 30
	backOffPeriod            = 5 * time.Second
	disk1                    = "Hard disk 1"
	disk2                    = "Hard disk 2"
	MemoryAvailable          = "Memory_Available"
)

const (
	// Documentation URLs.
	vSpherePermissionDoc = "https://anywhere.eks.amazonaws.com/docs/getting-started/vsphere/vsphere-preparation/"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

//go:embed config/secret.yaml
var defaultSecretObject string

//go:embed config/template-failuredomain.yaml
var defaultFailureDomainConfig string

//go:embed config/ipam-provider-crds.yaml
var defaultIPAMProviderCRDs string

//go:embed config/ipam-provider-deployment.yaml
var defaultIPAMProviderDeployment string

//go:embed config/template-ippool.yaml
var defaultIPPoolTemplate string

var (
	eksaVSphereDatacenterResourceType = fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereMachineResourceType    = fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group)
)

var requiredEnvs = []string{vSphereUsernameKey, vSpherePasswordKey, expClusterResourceSetKey}

type vsphereProvider struct {
	datacenterConfig      *v1alpha1.VSphereDatacenterConfig
	clusterConfig         *v1alpha1.Cluster
	providerGovcClient    ProviderGovcClient
	providerKubectlClient ProviderKubectlClient
	writer                filewriter.FileWriter
	templateBuilder       *VsphereTemplateBuilder
	skipIPCheck           bool
	Retrier               *retrier.Retrier
	validator             *Validator
	defaulter             *Defaulter
	ipValidator           IPValidator
	skippedValidations    map[string]bool
}

type ProviderGovcClient interface {
	SearchTemplate(ctx context.Context, datacenter, template string) (string, error)
	LibraryElementExists(ctx context.Context, library string) (bool, error)
	GetLibraryElementContentVersion(ctx context.Context, element string) (string, error)
	DeleteLibraryElement(ctx context.Context, element string) error
	TemplateHasSnapshot(ctx context.Context, template string) (bool, error)
	GetWorkloadAvailableSpace(ctx context.Context, datastore string) (float64, error)
	ValidateVCenterSetupMachineConfig(ctx context.Context, datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfig *v1alpha1.VSphereMachineConfig, selfSigned *bool) error
	ValidateFailureDomainConfig(ctx context.Context, datacenterConfig *v1alpha1.VSphereDatacenterConfig, failureDomain *v1alpha1.FailureDomain) error
	ValidateVCenterConnection(ctx context.Context, server string) error
	ValidateVCenterAuthentication(ctx context.Context) error
	IsCertSelfSigned(ctx context.Context) bool
	GetCertThumbprint(ctx context.Context) (string, error)
	ConfigureCertThumbprint(ctx context.Context, server, thumbprint string) error
	DatacenterExists(ctx context.Context, datacenter string) (bool, error)
	NetworkExists(ctx context.Context, network string) (bool, error)
	GetFolderPath(ctx context.Context, datacenter string, folder string, envMap map[string]string) (string, error)
	GetDatastorePath(ctx context.Context, datacenter string, datastorePath string, envMap map[string]string) (string, error)
	GetResourcePoolPath(ctx context.Context, datacenter string, resourcePool string, envMap map[string]string) (string, error)
	GetComputeClusterPath(ctx context.Context, datacenter string, computeCluster string, envMap map[string]string) (string, error)
	CreateLibrary(ctx context.Context, datastore, library string) error
	DeployTemplateFromLibrary(ctx context.Context, templateDir, templateName, library, datacenter, datastore, network, resourcePool string, resizeDisk2 bool) error
	ImportTemplate(ctx context.Context, library, ovaURL, name string) error
	GetVMDiskSizeInGB(ctx context.Context, vm, datacenter string) (int, error)
	GetTags(ctx context.Context, path string) (tags []string, err error)
	ListTags(ctx context.Context) ([]executables.Tag, error)
	CreateTag(ctx context.Context, tag, category string) error
	AddTag(ctx context.Context, path, tag string) error
	ListCategories(ctx context.Context) ([]string, error)
	CreateCategoryForVM(ctx context.Context, name string) error
	CreateUser(ctx context.Context, username, password string) error
	UserExists(ctx context.Context, username string) (bool, error)
	CreateGroup(ctx context.Context, name string) error
	GroupExists(ctx context.Context, name string) (bool, error)
	AddUserToGroup(ctx context.Context, name, username string) error
	RoleExists(ctx context.Context, name string) (bool, error)
	CreateRole(ctx context.Context, name string, privileges []string) error
	SetGroupRoleOnObject(ctx context.Context, principal, role, object, domain string) error
	GetHardDiskSize(ctx context.Context, vm, datacenter string) (map[string]float64, error)
	GetResourcePoolInfo(ctx context.Context, datacenter, resourcepool string, args ...string) (map[string]int, error)
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	CreateNamespaceIfNotPresent(ctx context.Context, kubeconfig, namespace string) error
	LoadSecret(ctx context.Context, secretObject, secretObjType, secretObjectName, kubeConfFile string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName, kubeconfigFile, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
	GetEksaVSphereMachineConfig(ctx context.Context, vsphereMachineConfigName, kubeconfigFile, namespace string) (*v1alpha1.VSphereMachineConfig, error)
	GetMachineDeployment(ctx context.Context, machineDeploymentName string, opts ...executables.KubectlOpt) (*clusterv1beta2.MachineDeployment, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*controlplanev1beta2.KubeadmControlPlane, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1.EtcdadmCluster, error)
	GetSecretFromNamespace(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
	RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error
	SearchVsphereMachineConfig(ctx context.Context, name, kubeconfigFile, namespace string) ([]*v1alpha1.VSphereMachineConfig, error)
	SearchVsphereDatacenterConfig(ctx context.Context, name, kubeconfigFile, namespace string) ([]*v1alpha1.VSphereDatacenterConfig, error)
	SetDaemonSetImage(ctx context.Context, kubeconfigFile, name, namespace, container, image string) error
	DeleteEksaDatacenterConfig(ctx context.Context, vsphereDatacenterResourceType, vsphereDatacenterConfigName, kubeconfigFile, namespace string) error
	DeleteEksaMachineConfig(ctx context.Context, vsphereMachineResourceType, vsphereMachineConfigName, kubeconfigFile, namespace string) error
	ApplyTolerationsFromTaintsToDaemonSet(ctx context.Context, oldTaints, newTaints []corev1.Taint, dsName, kubeconfigFile string) error
	HasCRD(ctx context.Context, crd, kubeconfig string) (bool, error)
}

// IPValidator is an interface that defines methods to validate the control plane IP.
type IPValidator interface {
	ValidateControlPlaneIPUniqueness(cluster *v1alpha1.Cluster) error
}

// NewProvider initializes and returns a new vsphereProvider.
func NewProvider(
	datacenterConfig *v1alpha1.VSphereDatacenterConfig,
	clusterConfig *v1alpha1.Cluster,
	providerGovcClient ProviderGovcClient,
	providerKubectlClient ProviderKubectlClient,
	writer filewriter.FileWriter,
	ipValidator IPValidator,
	now types.NowFunc,
	skipIPCheck bool,
	skippedValidations map[string]bool,
) *vsphereProvider { //nolint:revive
	// TODO(g-gaston): ignoring linter error for exported function returning unexported member
	// We should make it exported, but that would involve a bunch of changes, so will do it separately
	vcb := govmomi.NewVMOMIClientBuilder()
	v := NewValidator(
		providerGovcClient,
		vcb,
	)

	return NewProviderCustomNet(
		datacenterConfig,
		clusterConfig,
		providerGovcClient,
		providerKubectlClient,
		writer,
		ipValidator,
		now,
		skipIPCheck,
		v,
		skippedValidations,
	)
}

// NewProviderCustomNet initializes and returns a new vsphereProvider.
func NewProviderCustomNet(
	datacenterConfig *v1alpha1.VSphereDatacenterConfig,
	clusterConfig *v1alpha1.Cluster,
	providerGovcClient ProviderGovcClient,
	providerKubectlClient ProviderKubectlClient,
	writer filewriter.FileWriter,
	ipValidator IPValidator,
	now types.NowFunc,
	skipIPCheck bool,
	v *Validator,
	skippedValidations map[string]bool,
) *vsphereProvider { //nolint:revive
	// TODO(g-gaston): ignoring linter error for exported function returning unexported member
	// We should make it exported, but that would involve a bunch of changes, so will do it separately
	retrier := retrier.NewWithMaxRetries(maxRetries, backOffPeriod)
	return &vsphereProvider{
		datacenterConfig:      datacenterConfig,
		clusterConfig:         clusterConfig,
		providerGovcClient:    providerGovcClient,
		providerKubectlClient: providerKubectlClient,
		writer:                writer,
		templateBuilder: NewVsphereTemplateBuilder(
			now,
		),
		skipIPCheck:        skipIPCheck,
		Retrier:            retrier,
		validator:          v,
		defaulter:          NewDefaulter(providerGovcClient),
		ipValidator:        ipValidator,
		skippedValidations: skippedValidations,
	}
}

func (p *vsphereProvider) UpdateKubeConfig(_ *[]byte, _ string) error {
	// customize generated kube config
	return nil
}

func (p *vsphereProvider) BootstrapClusterOpts(spec *cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error) {
	return common.BootstrapClusterOpts(p.clusterConfig, spec.VSphereDatacenter.Spec.Server)
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

func (p *vsphereProvider) generateSSHKeysIfNotSet(machineConfigs map[string]*v1alpha1.VSphereMachineConfig) error {
	var generatedKey string
	for _, machineConfig := range machineConfigs {
		user := machineConfig.Spec.Users[0]
		if user.SshAuthorizedKeys[0] == "" {
			if generatedKey != "" { // use the same key
				user.SshAuthorizedKeys[0] = generatedKey
			} else {
				logger.Info("Provided sshAuthorizedKey is not set or is empty, auto-generating new key pair...", "vSphereMachineConfig", machineConfig.Name)
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

func (p *vsphereProvider) PostClusterDeleteValidate(_ context.Context, _ *types.Cluster) error {
	// No validations
	return nil
}

func (p *vsphereProvider) PostMoveManagementToBootstrap(_ context.Context, _ *types.Cluster) error {
	// NOOP
	return nil
}

func (p *vsphereProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := SetupEnvVars(clusterSpec.VSphereDatacenter); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	vSphereClusterSpec := NewSpec(clusterSpec)

	if err := p.defaulter.SetDefaultsForDatacenterConfig(ctx, vSphereClusterSpec.VSphereDatacenter); err != nil {
		return fmt.Errorf("failed setting default values for vsphere datacenter config: %v", err)
	}

	if err := vSphereClusterSpec.VSphereDatacenter.Validate(); err != nil {
		return err
	}

	// Validate IP pool size if configured
	if err := validateIPPoolSize(clusterSpec); err != nil {
		return err
	}

	if err := p.validator.ValidateVCenterConfig(ctx, vSphereClusterSpec.VSphereDatacenter); err != nil {
		return err
	}

	// Validate machine config networks right after basic vCenter validation
	if err := p.validator.validateNetworksFieldUsage(ctx, vSphereClusterSpec); err != nil {
		return err
	}

	if err := p.validator.ValidateFailureDomains(ctx, vSphereClusterSpec); err != nil {
		return err
	}

	if err := p.defaulter.setDefaultsForMachineConfig(ctx, vSphereClusterSpec); err != nil {
		return fmt.Errorf("failed setting default values for vsphere machine configs: %v", err)
	}

	if err := p.validator.ValidateClusterMachineConfigs(ctx, vSphereClusterSpec); err != nil {
		return err
	}
	if err := p.validateDatastoreUsageForCreate(ctx, vSphereClusterSpec); err != nil {
		return fmt.Errorf("validating vsphere machine configs datastore usage: %v", err)
	}
	if err := p.validateMemoryUsage(ctx, vSphereClusterSpec, nil); err != nil {
		return fmt.Errorf("validating vsphere machine configs resource pool memory usage: %v", err)
	}
	if err := p.generateSSHKeysIfNotSet(clusterSpec.VSphereMachineConfigs); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	// TODO: move this to validator
	if clusterSpec.Cluster.IsManaged() {
		for _, mc := range clusterSpec.VSphereMachineConfigs {
			em, err := p.providerKubectlClient.SearchVsphereMachineConfig(ctx, mc.GetName(), clusterSpec.ManagementCluster.KubeconfigFile, mc.GetNamespace())
			if err != nil {
				return err
			}
			if len(em) > 0 {
				return fmt.Errorf("VSphereMachineConfig %s already exists", mc.GetName())
			}
		}
		existingDatacenter, err := p.providerKubectlClient.SearchVsphereDatacenterConfig(ctx, clusterSpec.VSphereDatacenter.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
		if err != nil {
			return err
		}
		if len(existingDatacenter) > 0 {
			return fmt.Errorf("VSphereDatacenter %s already exists", clusterSpec.VSphereDatacenter.Name)
		}
		for _, identityProviderRef := range clusterSpec.Cluster.Spec.IdentityProviderRefs {
			if identityProviderRef.Kind == v1alpha1.OIDCConfigKind {
				clusterSpec.OIDCConfig.SetManagedBy(p.clusterConfig.ManagedBy())
			}
		}
	}

	if !p.skipIPCheck {
		if err := p.ipValidator.ValidateControlPlaneIPUniqueness(clusterSpec.Cluster); err != nil {
			return err
		}
	} else {
		logger.Info("Skipping check for whether control plane ip is in use")
	}

	if !p.skippedValidations[validations.VSphereUserPriv] {
		if err := p.validator.validateVsphereUserPrivs(ctx, vSphereClusterSpec); err != nil {
			return fmt.Errorf("validating vsphere user privileges: %w, please refer to %s for required permissions or use -v 3 for full missing permissions", err, vSpherePermissionDoc)
		}
	}

	return nil
}

func (p *vsphereProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec, _ *cluster.Spec) error {
	if err := SetupEnvVars(clusterSpec.VSphereDatacenter); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	vSphereClusterSpec := NewSpec(clusterSpec)

	if err := p.defaulter.SetDefaultsForDatacenterConfig(ctx, vSphereClusterSpec.VSphereDatacenter); err != nil {
		return fmt.Errorf("failed setting default values for vsphere datacenter config: %v", err)
	}

	if err := vSphereClusterSpec.VSphereDatacenter.Validate(); err != nil {
		return err
	}

	// Validate IP pool size if configured
	if err := validateIPPoolSize(clusterSpec); err != nil {
		return err
	}

	if err := p.validator.ValidateVCenterConfig(ctx, vSphereClusterSpec.VSphereDatacenter); err != nil {
		return err
	}

	// Validate machine config networks right after basic vCenter validation
	if err := p.validator.validateNetworksFieldUsage(ctx, vSphereClusterSpec); err != nil {
		return err
	}

	if err := p.validator.ValidateFailureDomains(ctx, vSphereClusterSpec); err != nil {
		return err
	}

	if err := p.defaulter.setDefaultsForMachineConfig(ctx, vSphereClusterSpec); err != nil {
		return fmt.Errorf("failed setting default values for vsphere machine configs: %v", err)
	}

	if err := p.validator.ValidateClusterMachineConfigs(ctx, vSphereClusterSpec); err != nil {
		return err
	}

	if err := p.validateDatastoreUsageForUpgrade(ctx, vSphereClusterSpec, cluster); err != nil {
		return fmt.Errorf("validating vsphere machine configs datastore usage: %v", err)
	}

	if err := p.validateMemoryUsage(ctx, vSphereClusterSpec, cluster); err != nil {
		return fmt.Errorf("validating vsphere machine configs resource pool memory usage: %v", err)
	}

	if !p.skippedValidations[validations.VSphereUserPriv] {
		if err := p.validator.validateVsphereUserPrivs(ctx, vSphereClusterSpec); err != nil {
			return fmt.Errorf("validating vsphere user privileges: %w, please refer to %s for required permissions or use -v 3 for full missing permissions", err, vSpherePermissionDoc)
		}
	}

	err := p.validateMachineConfigsNameUniqueness(ctx, cluster, clusterSpec)
	if err != nil {
		return fmt.Errorf("failed validate machineconfig uniqueness: %v", err)
	}
	return nil
}

// SetupAndValidateUpgradeManagementComponents performs necessary setup for upgrade management components operation.
func (p *vsphereProvider) SetupAndValidateUpgradeManagementComponents(ctx context.Context, clusterSpec *cluster.Spec) error {
	if err := SetupEnvVars(clusterSpec.VSphereDatacenter); err != nil {
		return fmt.Errorf("failed environment variable setup: %v", err)
	}

	return nil
}

func (p *vsphereProvider) validateMachineConfigsNameUniqueness(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName())
	if err != nil {
		return err
	}

	cpMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	if prevSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name != cpMachineConfigName {
		em, err := p.providerKubectlClient.SearchVsphereMachineConfig(ctx, cpMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace())
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
			em, err := p.providerKubectlClient.SearchVsphereMachineConfig(ctx, etcdMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace())
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

type datastoreUsage struct {
	availableSpace float64
	needGiBSpace   int
}

func (p *vsphereProvider) getPrevMachineConfigDatastoreUsage(ctx context.Context, machineConfig *v1alpha1.VSphereMachineConfig, cluster *types.Cluster, count int) (diskGiB float64, err error) {
	if count > 0 {
		em, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, machineConfig.Name, cluster.KubeconfigFile, machineConfig.GetNamespace())
		if err != nil {
			return 0, err
		}
		if em != nil {
			return float64(em.Spec.DiskGiB * count), nil
		}
	}
	return 0, nil
}

func (p *vsphereProvider) getMachineConfigDatastoreRequirements(ctx context.Context, machineConfig *v1alpha1.VSphereMachineConfig, count int) (available float64, need int, err error) {
	availableSpace, err := p.providerGovcClient.GetWorkloadAvailableSpace(ctx, machineConfig.Spec.Datastore) // TODO: remove dependency on machineConfig
	if err != nil {
		return 0, 0, fmt.Errorf("getting datastore details: %v", err)
	}
	needGiB := machineConfig.Spec.DiskGiB * count
	return availableSpace, needGiB, nil
}

func (p *vsphereProvider) calculateDatastoreUsage(ctx context.Context, machineConfig *v1alpha1.VSphereMachineConfig, cluster *types.Cluster, usage map[string]*datastoreUsage, prevCount, newCount int) error {
	availableSpace, needGiB, err := p.getMachineConfigDatastoreRequirements(ctx, machineConfig, newCount)
	if err != nil {
		return err
	}
	prevUsage, err := p.getPrevMachineConfigDatastoreUsage(ctx, machineConfig, cluster, prevCount)
	if err != nil {
		return err
	}
	availableSpace += prevUsage
	updateDatastoreUsageMap(machineConfig, needGiB, availableSpace, prevUsage, usage)
	return nil
}

func updateDatastoreUsageMap(machineConfig *v1alpha1.VSphereMachineConfig, needGiB int, availableSpace, prevUsage float64, usage map[string]*datastoreUsage) {
	if _, ok := usage[machineConfig.Spec.Datastore]; ok {
		usage[machineConfig.Spec.Datastore].needGiBSpace += needGiB
		usage[machineConfig.Spec.Datastore].availableSpace += prevUsage
	} else {
		usage[machineConfig.Spec.Datastore] = &datastoreUsage{
			availableSpace: availableSpace,
			needGiBSpace:   needGiB,
		}
	}
}

func (p *vsphereProvider) validateDatastoreUsageForUpgrade(ctx context.Context, currentClusterSpec *Spec, cluster *types.Cluster) error {
	usage := make(map[string]*datastoreUsage)
	prevEksaCluster, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, currentClusterSpec.Cluster.GetName())
	if err != nil {
		return err
	}

	cpMachineConfig := currentClusterSpec.controlPlaneMachineConfig()
	if err := p.calculateDatastoreUsage(ctx, cpMachineConfig, cluster, usage, prevEksaCluster.Spec.ControlPlaneConfiguration.Count, currentClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count); err != nil {
		return fmt.Errorf("calculating datastore usage: %v", err)
	}

	prevMachineConfigRefs := machineRefSliceToMap(prevEksaCluster.MachineConfigRefs())
	for _, workerNodeGroupConfiguration := range currentClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		prevCount := 0
		workerMachineConfig := currentClusterSpec.workerMachineConfig(workerNodeGroupConfiguration)
		if _, ok := prevMachineConfigRefs[workerNodeGroupConfiguration.MachineGroupRef.Name]; ok {
			prevCount = *workerNodeGroupConfiguration.Count
		}
		if err := p.calculateDatastoreUsage(ctx, workerMachineConfig, cluster, usage, prevCount, *workerNodeGroupConfiguration.Count); err != nil {
			return fmt.Errorf("calculating datastore usage: %v", err)
		}
	}

	etcdMachineConfig := currentClusterSpec.etcdMachineConfig()
	if etcdMachineConfig != nil {
		prevCount := 0
		if prevEksaCluster.Spec.ExternalEtcdConfiguration != nil {
			prevCount = prevEksaCluster.Spec.ExternalEtcdConfiguration.Count
		}
		if err := p.calculateDatastoreUsage(ctx, etcdMachineConfig, cluster, usage, prevCount, currentClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count); err != nil {
			return fmt.Errorf("calculating datastore usage: %v", err)
		}
	}

	for datastore, usage := range usage {
		if float64(usage.needGiBSpace) > usage.availableSpace {
			return fmt.Errorf("not enough space in datastore %v for given diskGiB and count for respective machine groups", datastore)
		}
	}
	return nil
}

func (p *vsphereProvider) validateDatastoreUsageForCreate(ctx context.Context, vsphereClusterSpec *Spec) error {
	usage := make(map[string]*datastoreUsage)
	cpMachineConfig := vsphereClusterSpec.controlPlaneMachineConfig()
	controlPlaneAvailableSpace, controlPlaneNeedGiB, err := p.getMachineConfigDatastoreRequirements(ctx, cpMachineConfig, vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count)
	if err != nil {
		return err
	}
	updateDatastoreUsageMap(cpMachineConfig, controlPlaneNeedGiB, controlPlaneAvailableSpace, 0, usage)

	for _, workerNodeGroupConfiguration := range vsphereClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerMachineConfig := vsphereClusterSpec.workerMachineConfig(workerNodeGroupConfiguration)
		workerAvailableSpace, workerNeedGiB, err := p.getMachineConfigDatastoreRequirements(ctx, workerMachineConfig, *workerNodeGroupConfiguration.Count)
		if err != nil {
			return err
		}
		updateDatastoreUsageMap(workerMachineConfig, workerNeedGiB, workerAvailableSpace, 0, usage)
	}

	etcdMachineConfig := vsphereClusterSpec.etcdMachineConfig()
	if etcdMachineConfig != nil {
		etcdAvailableSpace, etcdNeedGiB, err := p.getMachineConfigDatastoreRequirements(ctx, etcdMachineConfig, vsphereClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count)
		if err != nil {
			return err
		}
		updateDatastoreUsageMap(etcdMachineConfig, etcdNeedGiB, etcdAvailableSpace, 0, usage)
	}

	for datastore, usage := range usage {
		if float64(usage.needGiBSpace) > usage.availableSpace {
			return fmt.Errorf("not enough space in datastore %v for given diskGiB and count for respective machine groups", datastore)
		}
	}
	return nil
}

// getPrevMachineConfigMemoryUsage returns the memoryMiB freed up from the given machineConfig based on the count.
func (p *vsphereProvider) getPrevMachineConfigMemoryUsage(ctx context.Context, mc *v1alpha1.VSphereMachineConfig, cluster *types.Cluster, machineConfigCount int) (memoryMiB int, err error) {
	em, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, mc.Name, cluster.KubeconfigFile, mc.GetNamespace())
	if err != nil {
		return 0, err
	}
	if em != nil && em.Spec.ResourcePool == mc.Spec.ResourcePool {
		return em.Spec.MemoryMiB * machineConfigCount, nil
	}
	return 0, nil
}

// getMachineConfigMemoryAvailability accepts a machine config and returns available memory in the config's resource pool along with needed memory for the machine config.
func (p *vsphereProvider) getMachineConfigMemoryAvailability(ctx context.Context, datacenter string, mc *v1alpha1.VSphereMachineConfig, machineConfigCount int) (availableMemoryMiB, needMemoryMiB int, err error) {
	poolInfo, err := p.providerGovcClient.GetResourcePoolInfo(ctx, datacenter, mc.Spec.ResourcePool)
	if err != nil {
		return 0, 0, err
	}
	needMemoryMiB = mc.Spec.MemoryMiB * machineConfigCount
	return poolInfo[MemoryAvailable], needMemoryMiB, nil
}

// updateMemoryUsageMap updates the memory availability for the machine config's resource pool.
func updateMemoryUsageMap(mc *v1alpha1.VSphereMachineConfig, needMiB, availableMiB int, mu map[string]int) {
	if _, ok := mu[mc.Spec.ResourcePool]; !ok {
		mu[mc.Spec.ResourcePool] = availableMiB
	}
	// needMiB can be ignored when the resource pool memory limit is unset
	if availableMiB != -1 {
		mu[mc.Spec.ResourcePool] -= needMiB
	}
}

func addPrevMachineConfigMemoryUsage(mc *v1alpha1.VSphereMachineConfig, prevUsage int, memoryUsage map[string]int) {
	// when the memory limit for the respective resource pool is unset, skip accounting for previous usage and validating the needed memory
	if _, ok := memoryUsage[mc.Spec.ResourcePool]; ok && memoryUsage[mc.Spec.ResourcePool] != -1 {
		memoryUsage[mc.Spec.ResourcePool] += prevUsage
	}
}

func (p *vsphereProvider) validateMemoryUsage(ctx context.Context, clusterSpec *Spec, cluster *types.Cluster) error {
	memoryUsage := make(map[string]int)
	datacenter := clusterSpec.VSphereDatacenter.Spec.Datacenter
	for _, mc := range clusterSpec.machineConfigsWithCount() {
		availableMemoryMiB, needMemoryMiB, err := p.getMachineConfigMemoryAvailability(ctx, datacenter, mc.VSphereMachineConfig, mc.Count)
		if err != nil {
			return fmt.Errorf("calculating memory usage for machine config %v: %v", mc.VSphereMachineConfig.ObjectMeta.Name, err)
		}
		updateMemoryUsageMap(mc.VSphereMachineConfig, needMemoryMiB, availableMemoryMiB, memoryUsage)
	}
	// account for previous cluster resources that are freed up during upgrade.
	if cluster != nil {
		err := p.updatePrevClusterMemoryUsage(ctx, clusterSpec, cluster, memoryUsage)
		if err != nil {
			return err
		}
	}
	for resourcePool, remaniningMiB := range memoryUsage {
		if remaniningMiB != -1 && remaniningMiB < 0 {
			return fmt.Errorf("not enough memory available in resource pool %v for given memoryMiB and count for respective machine groups", resourcePool)
		}
	}
	logger.V(5).Info("Memory availability for machine configs in requested resource pool validated")
	return nil
}

// updatePrevClusterMemoryUsage calculates memory freed up from previous CP and worker nodes during upgrade and adds up the memory usage for the specific resource pool.
func (p *vsphereProvider) updatePrevClusterMemoryUsage(ctx context.Context, clusterSpec *Spec, cluster *types.Cluster, memoryUsage map[string]int) error {
	prevEksaCluster, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName())
	if err != nil {
		return err
	}
	prevMachineConfigRefs := machineRefSliceToMap(prevEksaCluster.MachineConfigRefs())
	if _, ok := prevMachineConfigRefs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]; ok {
		cpMachineConfig := clusterSpec.controlPlaneMachineConfig()
		// The last CP machine is deleted only after the desired number of new worker machines are rolled out, so don't add it's memory
		prevCPusage, err := p.getPrevMachineConfigMemoryUsage(ctx, cpMachineConfig, cluster, prevEksaCluster.Spec.ControlPlaneConfiguration.Count-1)
		if err != nil {
			return fmt.Errorf("calculating previous memory usage for control plane: %v", err)
		}
		addPrevMachineConfigMemoryUsage(cpMachineConfig, prevCPusage, memoryUsage)
	}
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerMachineConfig := clusterSpec.workerMachineConfig(workerNodeGroupConfiguration)
		if _, ok := prevMachineConfigRefs[workerNodeGroupConfiguration.MachineGroupRef.Name]; ok {
			prevCount := *workerNodeGroupConfiguration.Count
			// The last worker machine is deleted only after the desired number of new worker machines are rolled out, so don't add it's memory
			prevWorkerUsage, err := p.getPrevMachineConfigMemoryUsage(ctx, workerMachineConfig, cluster, prevCount-1)
			if err != nil {
				return fmt.Errorf("calculating previous memory usage for worker node group - %v: %v", workerMachineConfig.Name, err)
			}
			addPrevMachineConfigMemoryUsage(workerMachineConfig, prevWorkerUsage, memoryUsage)
		}
	}
	return nil
}

func (p *vsphereProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster, _ *cluster.Spec) error {
	var contents bytes.Buffer
	err := p.createSecret(ctx, cluster, &contents)
	if err != nil {
		return err
	}

	err = p.providerKubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, contents.Bytes())
	if err != nil {
		return fmt.Errorf("loading secrets object: %v", err)
	}
	return nil
}

func (p *vsphereProvider) SetupAndValidateDeleteCluster(ctx context.Context, _ *types.Cluster, spec *cluster.Spec) error {
	if err := SetupEnvVars(spec.VSphereDatacenter); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	return nil
}

func (p *vsphereProvider) createSecret(ctx context.Context, cluster *types.Cluster, contents *bytes.Buffer) error {
	t, err := template.New("tmpl").Funcs(sprig.TxtFuncMap()).Parse(defaultSecretObject)
	if err != nil {
		return fmt.Errorf("creating secret object template: %v", err)
	}
	vuc := config.NewVsphereUserConfig()

	values := map[string]string{
		"vspherePassword":           os.Getenv(vSpherePasswordKey),
		"vsphereUsername":           os.Getenv(vSphereUsernameKey),
		"eksaCloudProviderUsername": vuc.EksaVsphereCPUsername,
		"eksaCloudProviderPassword": vuc.EksaVsphereCPPassword,
		"eksaLicense":               os.Getenv(eksaLicense),
		"eksaSystemNamespace":       constants.EksaSystemNamespace,
		"vsphereCredentialsName":    constants.VSphereCredentialsName,
		"eksaLicenseName":           constants.EksaLicenseName,
	}
	err = t.Execute(contents, values)
	if err != nil {
		return fmt.Errorf("substituting values for secret object template: %v", err)
	}
	return nil
}

func (p *vsphereProvider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if err := p.UpdateSecrets(ctx, cluster, nil); err != nil {
		return err
	}

	// Install IPAM provider and create InClusterIPPool BEFORE CAPI installs VSphereMachineTemplates.
	// This is critical because VSphereMachineTemplates reference the InClusterIPPool via addressesFromPools,
	// and CAPV will fail to create IPAddressClaims if the pool doesn't exist.
	if clusterSpec != nil && clusterSpec.Config != nil && clusterSpec.VSphereDatacenter != nil && clusterSpec.VSphereDatacenter.Spec.IPPool != nil {
		logger.V(3).Info("Installing IPAM provider for static IP allocation",
			"poolName", clusterSpec.VSphereDatacenter.Spec.IPPool.Name,
		)
		if err := p.installIPAMProviderAndCreatePool(ctx, cluster, clusterSpec.VSphereDatacenter); err != nil {
			return fmt.Errorf("setting up static IP allocation in PreCAPIInstallOnBootstrap: %v", err)
		}
	}

	return nil
}

func (p *vsphereProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// IPAM setup has been moved to PreCAPIInstallOnBootstrap to ensure the InClusterIPPool
	// exists BEFORE VSphereMachineTemplates are applied. This is required because CAPV
	// creates IPAddressClaims when processing VSphereMachineTemplates with addressesFromPools.
	return nil
}

// installIPAMProviderAndCreatePool installs the CAPI IPAM provider and creates the InClusterIPPool.
func (p *vsphereProvider) installIPAMProviderAndCreatePool(ctx context.Context, cluster *types.Cluster, datacenterConfig *v1alpha1.VSphereDatacenterConfig) error {
	// Step 1: Install CAPI IPAM provider (cluster-api-ipam-provider-in-cluster)
	if err := p.installCAPIIPAMProvider(ctx, cluster); err != nil {
		return fmt.Errorf("installing CAPI IPAM provider: %v", err)
	}

	// Step 2: Create InClusterIPPool resource
	if err := p.createInClusterIPPoolFromConfig(ctx, cluster, datacenterConfig); err != nil {
		return fmt.Errorf("creating InClusterIPPool: %v", err)
	}

	return nil
}

// installCAPIIPAMProvider installs the cluster-api-ipam-provider-in-cluster components.
// Note: IPAddressClaim and IPAddress CRDs are part of core CAPI (v1.5+) and are already installed.
// This function only installs the InClusterIPPool/GlobalInClusterIPPool CRDs and the IPAM controller.
// The installation is split into two phases to avoid controller startup failures:
// 1. Apply CRDs and namespace first, wait for CRDs to be established
// 2. Apply RBAC and deployment after CRDs are ready.
func (p *vsphereProvider) installCAPIIPAMProvider(ctx context.Context, cluster *types.Cluster) error {
	// Step 1: Apply CRDs and namespace first
	err := p.providerKubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, []byte(defaultIPAMProviderCRDs))
	if err != nil {
		return fmt.Errorf("applying IPAM provider CRDs: %v", err)
	}

	// Step 2: Wait for CRDs to be established before deploying the controller
	// This prevents the controller from failing on startup due to missing API resources
	if err := p.waitForIPAMCRDs(ctx, cluster); err != nil {
		return fmt.Errorf("waiting for IPAM CRDs to be established: %v", err)
	}

	// Step 3: Apply RBAC and deployment
	err = p.providerKubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, []byte(defaultIPAMProviderDeployment))
	if err != nil {
		return fmt.Errorf("applying IPAM provider deployment: %v", err)
	}

	logger.V(3).Info("CAPI IPAM provider installed successfully")
	return nil
}

// waitForIPAMCRDs waits for the IPAM CRDs to be established in the cluster.
func (p *vsphereProvider) waitForIPAMCRDs(ctx context.Context, cluster *types.Cluster) error {
	crdsToWait := []string{
		"inclusterippools.ipam.cluster.x-k8s.io",
		"globalinclusterippools.ipam.cluster.x-k8s.io",
	}

	for _, crdName := range crdsToWait {
		err := p.Retrier.Retry(func() error {
			hasCRD, err := p.providerKubectlClient.HasCRD(ctx, crdName, cluster.KubeconfigFile)
			if err != nil {
				return err
			}
			if !hasCRD {
				return fmt.Errorf("CRD %s not yet established", crdName)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("waiting for CRD %s: %v", crdName, err)
		}
	}

	return nil
}

// validateIPPoolSize validates that the IP pool has sufficient IPs for the cluster nodes.
// It calculates the total number of nodes (control plane + workers + etcd) and ensures
// the pool has enough addresses, including extra IPs for rolling upgrades.
// For worker nodes with autoscaling, it uses MaxCount to ensure the pool can accommodate scale-up.
func validateIPPoolSize(clusterSpec *cluster.Spec) error {
	if clusterSpec.VSphereDatacenter == nil || clusterSpec.VSphereDatacenter.Spec.IPPool == nil {
		return nil
	}

	ipPool := clusterSpec.VSphereDatacenter.Spec.IPPool
	clusterConfig := clusterSpec.Cluster

	// Calculate total nodes needed
	cpCount := clusterConfig.Spec.ControlPlaneConfiguration.Count
	workerCount := 0
	etcdCount := 0

	for _, wng := range clusterConfig.Spec.WorkerNodeGroupConfigurations {
		// For autoscaling, use MaxCount to ensure pool has enough IPs for scale-up
		if wng.AutoScalingConfiguration != nil && wng.AutoScalingConfiguration.MaxCount > 0 {
			workerCount += wng.AutoScalingConfiguration.MaxCount
		} else if wng.Count != nil {
			workerCount += *wng.Count
		}
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdCount = clusterConfig.Spec.ExternalEtcdConfiguration.Count
	}

	totalNodes := cpCount + workerCount + etcdCount

	// Calculate pool size
	poolSize, err := v1alpha1.CalculateIPPoolSize(ipPool.Addresses)
	if err != nil {
		return fmt.Errorf("failed to calculate IP pool size: %v", err)
	}

	// Need extra IPs for rolling upgrades (maxSurge)
	requiredIPs := totalNodes + 1 // +1 for rolling upgrade buffer
	if poolSize < requiredIPs {
		return fmt.Errorf("ipPool '%s' has %d addresses but cluster requires at least %d (control plane: %d, workers: %d, etcd: %d, rolling upgrade buffer: 1)",
			ipPool.Name, poolSize, requiredIPs, cpCount, workerCount, etcdCount)
	}

	logger.Info("IP pool size validation passed",
		"poolName", ipPool.Name,
		"poolSize", poolSize,
		"totalNodes", totalNodes,
		"requiredIPs", requiredIPs,
	)

	return nil
}

// createInClusterIPPoolFromConfig creates an InClusterIPPool resource from the datacenter config.
func (p *vsphereProvider) createInClusterIPPoolFromConfig(ctx context.Context, cluster *types.Cluster, datacenterConfig *v1alpha1.VSphereDatacenterConfig) error {
	ipPool := datacenterConfig.Spec.IPPool
	if ipPool == nil {
		return nil
	}

	// Build the InClusterIPPool YAML using template
	values := map[string]interface{}{
		"ipPoolName":      ipPool.Name,
		"ipPoolNamespace": constants.EksaSystemNamespace,
		"ipPoolAddresses": ipPool.Addresses,
		"ipPoolPrefix":    ipPool.Prefix,
		"ipPoolGateway":   ipPool.Gateway,
	}

	t, err := template.New("ippool").Parse(defaultIPPoolTemplate)
	if err != nil {
		return fmt.Errorf("parsing InClusterIPPool template: %v", err)
	}

	var poolYAML bytes.Buffer
	if err := t.Execute(&poolYAML, values); err != nil {
		return fmt.Errorf("executing InClusterIPPool template: %v", err)
	}

	err = p.providerKubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, poolYAML.Bytes())
	if err != nil {
		return fmt.Errorf("applying InClusterIPPool: %v", err)
	}

	logger.V(3).Info("InClusterIPPool created successfully",
		"name", ipPool.Name,
		"namespace", constants.EksaSystemNamespace,
	)
	return nil
}

func (p *vsphereProvider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

// EnvMap returns a map of environment variables required for the vsphere provider.
func (p *vsphereProvider) EnvMap(_ *cluster.ManagementComponents, _ *cluster.Spec) (map[string]string, error) {
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

func (p *vsphereProvider) DatacenterConfig(spec *cluster.Spec) providers.DatacenterConfig {
	return spec.VSphereDatacenter
}

func (p *vsphereProvider) MachineConfigs(spec *cluster.Spec) []providers.MachineConfig {
	annotateMachineConfig(
		spec,
		spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name,
		spec.Cluster.ControlPlaneAnnotation(),
		"true",
	)
	if spec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		annotateMachineConfig(
			spec,
			spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name,
			spec.Cluster.EtcdAnnotation(),
			"true",
		)
	}

	for _, workerNodeGroupConfiguration := range p.clusterConfig.Spec.WorkerNodeGroupConfigurations {
		setMachineConfigManagedBy(
			spec,
			workerNodeGroupConfiguration.MachineGroupRef.Name,
		)
	}

	machineConfigs := make([]providers.MachineConfig, 0, len(spec.VSphereMachineConfigs))
	for _, m := range spec.VSphereMachineConfigs {
		machineConfigs = append(machineConfigs, m)
	}

	return machineConfigs
}

func annotateMachineConfig(spec *cluster.Spec, machineConfigName, annotationKey, annotationValue string) {
	machineConfig := spec.VSphereMachineConfigs[machineConfigName]
	if machineConfig.Annotations == nil {
		machineConfig.Annotations = make(map[string]string, 1)
	}
	machineConfig.Annotations[annotationKey] = annotationValue
	setMachineConfigManagedBy(spec, machineConfigName)
}

func setMachineConfigManagedBy(spec *cluster.Spec, machineConfigName string) {
	machineConfig := spec.VSphereMachineConfigs[machineConfigName]
	if machineConfig.Annotations == nil {
		machineConfig.Annotations = make(map[string]string, 1)
	}
	if spec.Cluster.IsManaged() {
		machineConfig.SetManagedBy(spec.Cluster.ManagedBy())
	}
}

func (p *vsphereProvider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name)
	if err != nil {
		return err
	}

	prevDatacenter, err := p.providerKubectlClient.GetEksaVSphereDatacenterConfig(ctx, prevSpec.Spec.DatacenterRef.Name, cluster.KubeconfigFile, prevSpec.Namespace)
	if err != nil {
		return err
	}

	datacenter := clusterSpec.VSphereDatacenter

	oSpec := prevDatacenter.Spec
	nSpec := datacenter.Spec

	prevMachineConfigRefs := machineRefSliceToMap(prevSpec.MachineConfigRefs())

	for _, machineConfigRef := range clusterSpec.Cluster.MachineConfigRefs() {
		machineConfig, ok := clusterSpec.VSphereMachineConfigs[machineConfigRef.Name]
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

	secretChanged, err := p.secretContentsChanged(ctx, cluster)
	if err != nil {
		return err
	}

	if secretChanged {
		return fmt.Errorf("the VSphere credentials derived from %s and %s are immutable; please use the same credentials for the upgraded cluster", vSpherePasswordKey, vSphereUsernameKey)
	}
	return nil
}

func (p *vsphereProvider) validateMachineConfigImmutability(ctx context.Context, cluster *types.Cluster, newConfig *v1alpha1.VSphereMachineConfig, clusterSpec *cluster.Spec) error {
	prevMachineConfig, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, newConfig.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
	if err != nil {
		return err
	}

	if newConfig.Spec.StoragePolicyName != prevMachineConfig.Spec.StoragePolicyName {
		return fmt.Errorf("spec.storagePolicyName is immutable. Previous value %s, new value %s", prevMachineConfig.Spec.StoragePolicyName, newConfig.Spec.StoragePolicyName)
	}

	if newConfig.Spec.OSFamily != prevMachineConfig.Spec.OSFamily {
		return fmt.Errorf("spec.osFamily is immutable. Previous value %v, new value %v", prevMachineConfig.Spec.OSFamily, newConfig.Spec.OSFamily)
	}

	return nil
}

func (p *vsphereProvider) secretContentsChanged(ctx context.Context, workloadCluster *types.Cluster) (bool, error) {
	nPassword := os.Getenv(vSpherePasswordKey)
	oSecret, err := p.providerKubectlClient.GetSecretFromNamespace(ctx, workloadCluster.KubeconfigFile, CredentialsObjectName, constants.EksaSystemNamespace)
	if err != nil {
		return false, fmt.Errorf("obtaining VSphere secret %s from workload cluster: %v", CredentialsObjectName, err)
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

// ChangeDiff returns the component change diff for the provider.
func (p *vsphereProvider) ChangeDiff(currentComponents, newComponents *cluster.ManagementComponents) *types.ComponentChangeDiff {
	if currentComponents.VSphere.Version == newComponents.VSphere.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.VSphereProviderName,
		NewVersion:    newComponents.VSphere.Version,
		OldVersion:    currentComponents.VSphere.Version,
	}
}

// GetInfrastructureBundle returns the infrastructure bundle for the provider.
func (p *vsphereProvider) GetInfrastructureBundle(components *cluster.ManagementComponents) *types.InfrastructureBundle {
	folderName := fmt.Sprintf("infrastructure-vsphere/%s/", components.VSphere.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			components.VSphere.Components,
			components.VSphere.Metadata,
			components.VSphere.ClusterTemplate,
		},
	}

	return &infraBundle
}

// Version returns the version of the provider.
func (p *vsphereProvider) Version(components *cluster.ManagementComponents) string {
	return components.VSphere.Version
}

func (p *vsphereProvider) RunPostControlPlaneUpgrade(_ context.Context, _, _ *cluster.Spec, _, _ *types.Cluster) error {
	return nil
}

func cpiResourceSetName(clusterSpec *cluster.Spec) string {
	return fmt.Sprintf("%s-cpi", clusterSpec.Cluster.Name)
}

func machineRefSliceToMap(machineRefs []v1alpha1.Ref) map[string]v1alpha1.Ref {
	refMap := make(map[string]v1alpha1.Ref, len(machineRefs))
	for _, ref := range machineRefs {
		refMap[ref.Name] = ref
	}
	return refMap
}

func (p *vsphereProvider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	// Check if IPPool is configured in VSphereDatacenterConfig - if so, install IPAM provider on the management cluster
	if p.datacenterConfig != nil && p.datacenterConfig.Spec.IPPool != nil {
		logger.V(3).Info("Installing IPAM provider on management cluster",
			"poolName", p.datacenterConfig.Spec.IPPool.Name,
		)

		// Create a temporary cluster object with the kubeconfig for kubectl operations
		targetCluster := &types.Cluster{
			KubeconfigFile: kubeconfigFile,
		}

		// Step 1: Install CAPI IPAM provider (CRDs, RBAC, Deployment)
		if err := p.installCAPIIPAMProvider(ctx, targetCluster); err != nil {
			return fmt.Errorf("installing CAPI IPAM provider on management cluster: %v", err)
		}

		// Step 2: Create InClusterIPPool resource
		if err := p.createInClusterIPPoolFromConfig(ctx, targetCluster, p.datacenterConfig); err != nil {
			return fmt.Errorf("creating InClusterIPPool on management cluster: %v", err)
		}
	}

	return nil
}

// PreCoreComponentsUpgrade satisfies the Provider interface.
func (p *vsphereProvider) PreCoreComponentsUpgrade(
	ctx context.Context,
	cluster *types.Cluster,
	managementComponents *cluster.ManagementComponents,
	clusterSpec *cluster.Spec,
) error {
	return nil
}
