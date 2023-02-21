package tinkerbell

import (
	"context"
	"errors"
	"fmt"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	rufiov1alpha1 "github.com/tinkerbell/rufio/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
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
	unstructuredutil "github.com/aws/eks-anywhere/pkg/utils/unstructured"
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
	hardwareSpec    []byte
	catalogue       *hardware.Catalogue
	tinkerbellIP    string

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
	GetSecretFromNamespace(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.Secret, error)
	GetRufioMachine(ctx context.Context, name string, namespace string, kubeconfig string) (*rufiov1alpha1.Machine, error)
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

	return &Provider{
		clusterConfig:         clusterConfig,
		datacenterConfig:      datacenterConfig,
		machineConfigs:        machineConfigs,
		stackInstaller:        stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, clusterConfig.Spec.ClusterNetwork.Pods.CidrBlocks[0], registrymirror.FromCluster(clusterConfig)),
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

func (p *Provider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.Tinkerbell.Version
}

func (p *Provider) EnvMap(spec *cluster.Spec) (map[string]string, error) {
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

func (p *Provider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capt-system": {"capt-controller-manager"},
	}
}

func (p *Provider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	bundle := clusterSpec.VersionsBundle
	folderName := fmt.Sprintf("infrastructure-tinkerbell/%s/", bundle.Tinkerbell.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			bundle.Tinkerbell.Components,
			bundle.Tinkerbell.Metadata,
			bundle.Tinkerbell.ClusterTemplate,
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

func (p *Provider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	if currentSpec.VersionsBundle.Tinkerbell.Version == newSpec.VersionsBundle.Tinkerbell.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.TinkerbellProviderName,
		NewVersion:    newSpec.VersionsBundle.Tinkerbell.Version,
		OldVersion:    currentSpec.VersionsBundle.Tinkerbell.Version,
	}
}

func (p *Provider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	return nil
}

// AdditionalFiles returns additional files needed to be stored by providers.
// For Tinkerbell, this includes the hardware yaml containing the spec for the hardware in the management cluster.
func (p *Provider) AdditionalFiles() map[string][]byte {
	additionalFiles := make(map[string][]byte, 0)
	additionalFiles["hardware"] = p.hardwareSpec
	return additionalFiles
}

// generateHardwareSpec reads the hardware information from the available cluster, generates a yaml that can be submitted to a kubernetes cluster
// and returns it.
func (p *Provider) generateHardwareSpec(ctx context.Context, cluster *types.Cluster) ([]byte, error) {
	catalogue := hardware.NewCatalogue()
	catalogueWriter := hardware.NewMachineCatalogueWriter(catalogue)
	hwList, err := p.providerKubectlClient.AllTinkerbellHardware(ctx, cluster.KubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to build catalogue: %v", err)
	}

	for _, hw := range hwList {
		machine, err := p.buildHardwareMachineFromCluster(ctx, cluster, &hw)
		if err != nil {
			return nil, fmt.Errorf("reading hardware machine from cluster: %v", err)
		}

		err = catalogueWriter.Write(*machine)
		if err != nil {
			return nil, fmt.Errorf("writing machine to catalogue: %v", err)
		}
	}

	hardwareSpec, err := hardware.MarshalCatalogue(catalogue)
	if err != nil {
		return nil, fmt.Errorf("marshing hardware catalogue: %v", err)
	}
	hardwareSpec, err = unstructuredutil.StripNull(hardwareSpec)
	if err != nil {
		return nil, fmt.Errorf("stripping null values from hardware spec: %v", err)
	}

	return hardwareSpec, nil
}

// buildHardwareMachineFromCluster fetches all the hardware, bmc machines and related secret objects from the cluster,
// then it converts that data into a hardware.Machine and returns it.
func (p *Provider) buildHardwareMachineFromCluster(ctx context.Context, cluster *types.Cluster, hw *tinkv1alpha1.Hardware) (*hardware.Machine, error) {
	if hw.Spec.BMCRef == nil {
		machine := hardware.NewMachineFromHardware(*hw, nil, nil)
		return &machine, nil
	}

	rufioMachine, err := p.providerKubectlClient.GetRufioMachine(ctx, hw.Spec.BMCRef.Name, hw.Namespace, cluster.KubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("getting rufio machine: %v", err)
	}

	authSecret, err := p.providerKubectlClient.GetSecretFromNamespace(ctx, cluster.KubeconfigFile, rufioMachine.Spec.Connection.AuthSecretRef.Name, hw.Namespace)
	if err != nil {
		return nil, fmt.Errorf("getting rufio machine auth secret: %v", err)
	}

	machine := hardware.NewMachineFromHardware(*hw, rufioMachine, authSecret)
	return &machine, nil
}
