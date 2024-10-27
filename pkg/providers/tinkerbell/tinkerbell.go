package tinkerbell

import (
	"context"
	"errors"
	"fmt"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/rufiounreleased"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	maxRetries    = 30
	backOffPeriod = 5 * time.Second
)

var (
	eksaTinkerbellDatacenterResourceType = fmt.Sprintf("tinkerbelldatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaTinkerbellMachineResourceType    = fmt.Sprintf("tinkerbellmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	tinkerbellStackPorts                 = []int{42113, 50051, 50061}

	// errExternalEtcdUnsupported is returned from create or update when the user attempts to create
	// or upgrade a cluster with an external etcd configuration.
	errExternalEtcdUnsupported = errors.New("external etcd configuration is unsupported")

	referrencedMachineConfigsAvailabilityErrMsg = "some machine configs (%s) referenced in cluster config are not provided"
)

type Provider struct {
	clusterConfig         *v1alpha1.Cluster
	datacenterConfig      *v1alpha1.TinkerbellDatacenterConfig
	machineConfigs        map[string]*v1alpha1.TinkerbellMachineConfig
	stackInstaller        stack.StackInstaller
	providerKubectlClient ProviderKubectlClient
	templateBuilder       *TemplateBuilder
	writer                filewriter.FileWriter
	keyGenerator          SSHAuthKeyGenerator

	hardwareCSVFile string
	catalogue       *hardware.Catalogue
	tinkerbellIP    string
	// BMCOptions are Rufio BMC options that are used when creating Rufio machine CRDs.
	BMCOptions *hardware.BMCOptions

	// TODO(chrisdoheryt4) Temporarily depend on the netclient until the validator can be injected.
	// This is already a dependency, just uncached, because we require it during the initializing
	// constructor call for constructing the validator in-line.
	netClient networkutils.NetClient

	forceCleanup bool
	skipIpCheck  bool
	retrier      *retrier.Retrier
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error
	ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error
	DeleteEksaDatacenterConfig(ctx context.Context, eksaTinkerbellDatacenterResourceType string, tinkerbellDatacenterConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaMachineConfig(ctx context.Context, eksaTinkerbellMachineResourceType string, tinkerbellMachineConfigName string, kubeconfigFile string, namespace string) error
	GetMachineDeployment(ctx context.Context, machineDeploymentName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaTinkerbellDatacenterConfig(ctx context.Context, tinkerbellDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.TinkerbellDatacenterConfig, error)
	GetEksaTinkerbellMachineConfig(ctx context.Context, tinkerbellMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.TinkerbellMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*controlplanev1.KubeadmControlPlane, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1.EtcdadmCluster, error)
	GetSecret(ctx context.Context, secretObjectName string, opts ...executables.KubectlOpt) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
	WaitForDeployment(ctx context.Context, cluster *types.Cluster, timeout string, condition string, target string, namespace string) error
	GetUnprovisionedTinkerbellHardware(_ context.Context, kubeconfig, namespace string) ([]tinkv1alpha1.Hardware, error)
	GetProvisionedTinkerbellHardware(_ context.Context, kubeconfig, namespace string) ([]tinkv1alpha1.Hardware, error)
	WaitForRufioMachines(ctx context.Context, cluster *types.Cluster, timeout string, condition string, namespace string) error
	SearchTinkerbellMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.TinkerbellMachineConfig, error)
	SearchTinkerbellDatacenterConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.TinkerbellDatacenterConfig, error)
	AllTinkerbellHardware(ctx context.Context, kuebconfig string) ([]tinkv1alpha1.Hardware, error)
	AllBaseboardManagements(ctx context.Context, kubeconfig string) ([]rufiounreleased.BaseboardManagement, error)
	HasCRD(ctx context.Context, kubeconfig, crd string) (bool, error)
	DeleteCRD(ctx context.Context, kubeconfig, crd string) error
}

// KeyGenerator generates ssh keys and writes them to a FileWriter.
type SSHAuthKeyGenerator interface {
	GenerateSSHAuthKey(filewriter.FileWriter) (string, error)
}

func NewProvider(
	datacenterConfig *v1alpha1.TinkerbellDatacenterConfig,
	machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig,
	clusterConfig *v1alpha1.Cluster,
	hardwareCSVPath string,
	writer filewriter.FileWriter,
	docker stack.Docker,
	helm stack.Helm,
	providerKubectlClient ProviderKubectlClient,
	tinkerbellIP string,
	now types.NowFunc,
	forceCleanup bool,
	skipIpCheck bool,
) (*Provider, error) {
	var controlPlaneMachineSpec, workerNodeGroupMachineSpec, etcdMachineSpec *v1alpha1.TinkerbellMachineConfigSpec

	if err := validateRefrencedMachineConfigsAvailability(machineConfigs, clusterConfig); err != nil {
		return nil, err
	}

	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.TinkerbellMachineConfigSpec, len(machineConfigs))
	for _, wnConfig := range clusterConfig.Spec.WorkerNodeGroupConfigurations {
		if wnConfig.MachineGroupRef != nil && machineConfigs[wnConfig.MachineGroupRef.Name] != nil {
			workerNodeGroupMachineSpec = &machineConfigs[wnConfig.MachineGroupRef.Name].Spec
			workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = *workerNodeGroupMachineSpec
		}
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}

	var proxyConfig *v1alpha1.ProxyConfiguration
	if clusterConfig.Spec.ProxyConfiguration != nil {
		proxyConfig = &v1alpha1.ProxyConfiguration{
			HttpProxy:  clusterConfig.Spec.ProxyConfiguration.HttpProxy,
			HttpsProxy: clusterConfig.Spec.ProxyConfiguration.HttpsProxy,
			NoProxy:    generateNoProxyList(clusterConfig, datacenterConfig.Spec),
		}
		// We need local tinkerbell IP only in case of management
		// cluster's create and upgrade that too for the kind cluster.
		// GenerateNoProxyList is getting used by all the cluster operations.
		// Thus moving adding tinkerbell Local IP to here.
		if !slices.Contains(proxyConfig.NoProxy, tinkerbellIP) {
			proxyConfig.NoProxy = append(proxyConfig.NoProxy, tinkerbellIP)
		}
	} else {
		proxyConfig = nil
	}

	return &Provider{
		clusterConfig:         clusterConfig,
		datacenterConfig:      datacenterConfig,
		machineConfigs:        machineConfigs,
		stackInstaller:        stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, clusterConfig.Spec.ClusterNetwork.Pods.CidrBlocks[0], registrymirror.FromCluster(clusterConfig), proxyConfig),
		providerKubectlClient: providerKubectlClient,
		templateBuilder: &TemplateBuilder{
			datacenterSpec:              &datacenterConfig.Spec,
			controlPlaneMachineSpec:     controlPlaneMachineSpec,
			WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
			etcdMachineSpec:             etcdMachineSpec,
			tinkerbellIP:                tinkerbellIP,
			now:                         now,
		},
		writer:          writer,
		hardwareCSVFile: hardwareCSVPath,
		// TODO(chrisdoherty4) Inject the catalogue dependency so we can dynamically construcft the
		// indexing capabilities.
		catalogue: hardware.NewCatalogue(
			hardware.WithHardwareIDIndex(),
			hardware.WithHardwareBMCRefIndex(),
			hardware.WithBMCNameIndex(),
			hardware.WithSecretNameIndex(),
		),
		tinkerbellIP: tinkerbellIP,
		netClient:    &networkutils.DefaultNetClient{},
		retrier:      retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
		// (chrisdoherty4) We're hard coding the dependency and monkey patching in testing because the provider
		// isn't very testable right now and we already have tests in the `tinkerbell` package so can monkey patch
		// directly. This is very much a hack for testability.
		keyGenerator: common.SshAuthKeyGenerator{},
		// Behavioral flags.
		forceCleanup: forceCleanup,
		skipIpCheck:  skipIpCheck,
	}, nil
}

func (p *Provider) Name() string {
	return constants.TinkerbellProviderName
}

func (p *Provider) DatacenterResourceType() string {
	return eksaTinkerbellDatacenterResourceType
}

func (p *Provider) MachineResourceType() string {
	return eksaTinkerbellMachineResourceType
}

func (p *Provider) UpdateSecrets(ctx context.Context, cluster *types.Cluster, _ *cluster.Spec) error {
	// TODO: implement
	return nil
}

func (p *Provider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	// TODO: Figure out if something is needed here
	return nil
}

// Version returns the version of the provider.
func (p *Provider) Version(components *cluster.ManagementComponents) string {
	return components.Tinkerbell.Version
}

// EnvMap returns a map of environment variables for the tinkerbell provider.
func (p *Provider) EnvMap(_ *cluster.ManagementComponents, _ *cluster.Spec) (map[string]string, error) {
	return map[string]string{
		// The TINKERBELL_IP is input for the CAPT deployment and used as part of default template
		// generation. However, we use custom templates and leverage the template override
		// functionality of CAPT hence this never gets used.
		//
		// Deployment manifest requiring the env var for replacement.
		// https://github.com/tinkerbell/cluster-api-provider-tinkerbell/blob/main/config/manager/manager.yaml#L23
		//
		// Template override
		// https://github.com/tinkerbell/cluster-api-provider-tinkerbell/blob/main/controllers/machine.go#L182
		//
		// Env read having set TINKERBELL_IP in the deployment manifest.
		// https://github.com/tinkerbell/cluster-api-provider-tinkerbell/blob/main/controllers/machine.go#L192
		"TINKERBELL_IP":               "IGNORED",
		"KUBEADM_BOOTSTRAP_TOKEN_TTL": "120m",
	}, nil
}

// SetStackInstaller configures p to use installer for Tinkerbell stack install and upgrade.
func (p *Provider) SetStackInstaller(installer stack.StackInstaller) {
	p.stackInstaller = installer
}

func (p *Provider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capt-system": {"capt-controller-manager"},
	}
}

// GetInfrastructureBundle returns the infrastructure bundle for the provider.
func (p *Provider) GetInfrastructureBundle(components *cluster.ManagementComponents) *types.InfrastructureBundle {
	folderName := fmt.Sprintf("infrastructure-tinkerbell/%s/", components.Tinkerbell.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			components.Tinkerbell.Components,
			components.Tinkerbell.Metadata,
			components.Tinkerbell.ClusterTemplate,
		},
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

	return providers.ConfigsMapToSlice(configs)
}

// ChangeDiff returns the component change diff for the provider.
func (p *Provider) ChangeDiff(currentComponents, newComponents *cluster.ManagementComponents) *types.ComponentChangeDiff {
	if currentComponents.Tinkerbell.Version == newComponents.Tinkerbell.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.TinkerbellProviderName,
		NewVersion:    newComponents.Tinkerbell.Version,
		OldVersion:    currentComponents.Tinkerbell.Version,
	}
}

func (p *Provider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	return nil
}

func validateRefrencedMachineConfigsAvailability(machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster) error {
	unavailableMachineConfigNames := ""

	controlPlaneMachineName := clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	if _, ok := machineConfigs[controlPlaneMachineName]; !ok {
		unavailableMachineConfigNames = fmt.Sprintf("%s, %s", unavailableMachineConfigNames, controlPlaneMachineName)
	}

	for _, workerNodeGroupConfiguration := range clusterConfig.Spec.WorkerNodeGroupConfigurations {
		if workerNodeGroupConfiguration.MachineGroupRef == nil {
			continue
		}
		workerMachineName := workerNodeGroupConfiguration.MachineGroupRef.Name
		if _, ok := machineConfigs[workerMachineName]; !ok {
			unavailableMachineConfigNames = fmt.Sprintf("%s, %s", unavailableMachineConfigNames, workerMachineName)
		}
	}

	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		if _, ok := machineConfigs[etcdMachineName]; !ok {
			unavailableMachineConfigNames = fmt.Sprintf("%s, %s", unavailableMachineConfigNames, etcdMachineName)
		}
	}

	if len(unavailableMachineConfigNames) > 2 {
		unavailableMachineConfigNames = unavailableMachineConfigNames[2:]
		return fmt.Errorf(referrencedMachineConfigsAvailabilityErrMsg, unavailableMachineConfigNames)
	}

	return nil
}
