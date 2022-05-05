package tinkerbell

import (
	"context"
	"fmt"
	"os"
	"time"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	tinkhardware "github.com/tinkerbell/tink/protos/hardware"
	tinkworkflow "github.com/tinkerbell/tink/protos/workflow"
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
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	tinkerbellCertURLKey           = "TINKERBELL_CERT_URL"
	tinkerbellGRPCAuthKey          = "TINKERBELL_GRPC_AUTHORITY"
	tinkerbellIPKey                = "TINKERBELL_IP"
	tinkerbellPBnJGRPCAuthorityKey = "PBNJ_GRPC_AUTHORITY"
	tinkerbellHegelURLKey          = "TINKERBELL_HEGEL_URL"
	bmcStatePowerActionHardoff     = "POWER_ACTION_HARDOFF"
	tinkerbellOwnerNameLabel       = "v1alpha1.tinkerbell.org/ownerName"
	Provisioning                   = "provisioning"
	maxRetries                     = 30
	backOffPeriod                  = 5 * time.Second
)

const defaultUsername = "ec2-user"

var (
	eksaTinkerbellDatacenterResourceType = fmt.Sprintf("tinkerbelldatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaTinkerbellMachineResourceType    = fmt.Sprintf("tinkerbellmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	requiredEnvs                         = []string{tinkerbellCertURLKey, tinkerbellGRPCAuthKey, tinkerbellIPKey, tinkerbellPBnJGRPCAuthorityKey, tinkerbellHegelURLKey}
)

type tinkerbellProvider struct {
	clusterConfig         *v1alpha1.Cluster
	datacenterConfig      *v1alpha1.TinkerbellDatacenterConfig
	machineConfigs        map[string]*v1alpha1.TinkerbellMachineConfig
	hardwares             []tinkv1alpha1.Hardware
	providerKubectlClient ProviderKubectlClient
	providerTinkClient    ProviderTinkClient
	pbnj                  ProviderPbnjClient
	templateBuilder       *TinkerbellTemplateBuilder
	validator             *Validator
	writer                filewriter.FileWriter
	keyGenerator          SSHAuthKeyGenerator

	hardwareManifestPath string
	// catalogue is a cache initialized during SetupAndValidateCreateCluster() from hardwareManifestPath.
	catalogue hardware.Catalogue

	skipIpCheck      bool
	skipPowerActions bool
	setupTinkerbell  bool
	force            bool
	Retrier          *retrier.Retrier
}

type TinkerbellClients struct {
	ProviderTinkClient ProviderTinkClient
	ProviderPbnjClient ProviderPbnjClient
}

// TODO: Add necessary kubectl functions here
type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error
	DeleteEksaDatacenterConfig(ctx context.Context, eksaTinkerbellDatacenterResourceType string, tinkerbellDatacenterConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaMachineConfig(ctx context.Context, eksaTinkerbellMachineResourceType string, tinkerbellMachineConfigName string, kubeconfigFile string, namespace string) error
	GetMachineDeployment(ctx context.Context, machineDeploymentName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetHardwareWithLabel(ctx context.Context, label, kubeconfigFile, namespace string) ([]tinkv1alpha1.Hardware, error)
	GetBmcsPowerState(ctx context.Context, bmcNames []string, kubeconfigFile, namespace string) ([]string, error)
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaTinkerbellDatacenterConfig(ctx context.Context, tinkerbellDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.TinkerbellDatacenterConfig, error)
	GetEksaTinkerbellMachineConfig(ctx context.Context, tinkerbellMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.TinkerbellMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*controlplanev1.KubeadmControlPlane, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1.EtcdadmCluster, error)
	GetSecret(ctx context.Context, secretObjectName string, opts ...executables.KubectlOpt) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
}

type ProviderTinkClient interface {
	GetHardware(ctx context.Context) ([]*tinkhardware.Hardware, error)
	GetHardwareByUuid(ctx context.Context, uuid string) (*hardware.Hardware, error)
	PushHardware(ctx context.Context, hardware []byte) error
	GetWorkflow(ctx context.Context) ([]*tinkworkflow.Workflow, error)
	DeleteWorkflow(ctx context.Context, workflowIDs ...string) error
}

type ProviderPbnjClient interface {
	GetPowerState(ctx context.Context, bmc pbnj.BmcSecretConfig) (pbnj.PowerState, error)
	PowerOff(context.Context, pbnj.BmcSecretConfig) error
	PowerOn(context.Context, pbnj.BmcSecretConfig) error
	SetBootDevice(ctx context.Context, info pbnj.BmcSecretConfig, mode pbnj.BootDevice) error
}

// KeyGenerator generates ssh keys and writes them to a FileWriter.
type SSHAuthKeyGenerator interface {
	GenerateSSHAuthKey(filewriter.FileWriter) (string, error)
}

func NewProvider(
	datacenterConfig *v1alpha1.TinkerbellDatacenterConfig,
	machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig,
	clusterConfig *v1alpha1.Cluster,
	writer filewriter.FileWriter,
	providerKubectlClient ProviderKubectlClient,
	providerTinkbellClient TinkerbellClients,
	now types.NowFunc,
	skipIpCheck bool,
	hardwareManifestPath string,
	skipPowerActions bool,
	setupTinkerbell bool,
	force bool,
) *tinkerbellProvider {
	return NewProviderCustomDep(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		writer,
		providerKubectlClient,
		providerTinkbellClient.ProviderTinkClient,
		providerTinkbellClient.ProviderPbnjClient,
		&networkutils.DefaultNetClient{},
		now,
		skipIpCheck,
		hardwareManifestPath,
		skipPowerActions,
		setupTinkerbell,
		force,
	)
}

func NewProviderCustomDep(
	datacenterConfig *v1alpha1.TinkerbellDatacenterConfig,
	machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig,
	clusterConfig *v1alpha1.Cluster,
	writer filewriter.FileWriter,
	providerKubectlClient ProviderKubectlClient,
	providerTinkClient ProviderTinkClient,
	pbnjClient ProviderPbnjClient,
	netClient networkutils.NetClient,
	now types.NowFunc,
	skipIpCheck bool,
	hardwareManifestPath string,
	skipPowerActions bool,
	setupTinkerbell bool,
	force bool,
) *tinkerbellProvider {
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
	retrier := retrier.NewWithMaxRetries(maxRetries, backOffPeriod)
	return &tinkerbellProvider{
		clusterConfig:         clusterConfig,
		datacenterConfig:      datacenterConfig,
		machineConfigs:        machineConfigs,
		providerKubectlClient: providerKubectlClient,
		providerTinkClient:    providerTinkClient,
		pbnj:                  pbnjClient,
		templateBuilder: &TinkerbellTemplateBuilder{
			datacenterSpec:              &datacenterConfig.Spec,
			controlPlaneMachineSpec:     controlPlaneMachineSpec,
			WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
			etcdMachineSpec:             etcdMachineSpec,
			now:                         now,
		},
		hardwareManifestPath: hardwareManifestPath,
		validator:            NewValidator(providerTinkClient, netClient, pbnjClient),
		writer:               writer,

		// (chrisdoherty4) We're hard coding the dependency and monkey patching in testing because the provider
		// isn't very testable right now and we already have tests in the `tinkerbell` package so can monkey patch
		// directly. This is very much a hack for testability.
		keyGenerator: common.SshAuthKeyGenerator{},

		// Behavioral flags.
		skipIpCheck:      skipIpCheck,
		skipPowerActions: skipPowerActions,
		setupTinkerbell:  setupTinkerbell,
		force:            force,
		Retrier:          retrier,
	}
}

func (p *tinkerbellProvider) Name() string {
	return constants.TinkerbellProviderName
}

func (p *tinkerbellProvider) DatacenterResourceType() string {
	return eksaTinkerbellDatacenterResourceType
}

func (p *tinkerbellProvider) MachineResourceType() string {
	return eksaTinkerbellMachineResourceType
}

func (p *tinkerbellProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	// TODO: implement
	return nil
}

func (p *tinkerbellProvider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *tinkerbellProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.Tinkerbell.Version
}

func (p *tinkerbellProvider) EnvMap(_ *cluster.Spec) (map[string]string, error) {
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

func (p *tinkerbellProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capt-system": {"capt-controller-manager"},
	}
}

func (p *tinkerbellProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
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

func (p *tinkerbellProvider) DatacenterConfig(_ *cluster.Spec) providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *tinkerbellProvider) MachineConfigs(_ *cluster.Spec) []providers.MachineConfig {
	// TODO: Figure out if something is needed here
	var configs []providers.MachineConfig
	controlPlaneMachineName := p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerMachineName := p.clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	p.machineConfigs[controlPlaneMachineName].Annotations = map[string]string{p.clusterConfig.ControlPlaneAnnotation(): "true"}
	if p.clusterConfig.IsManaged() {
		p.machineConfigs[controlPlaneMachineName].SetManagedBy(p.clusterConfig.ManagedBy())
	}

	configs = append(configs, p.machineConfigs[controlPlaneMachineName])
	if workerMachineName != controlPlaneMachineName {
		configs = append(configs, p.machineConfigs[workerMachineName])
		if p.clusterConfig.IsManaged() {
			p.machineConfigs[workerMachineName].SetManagedBy(p.clusterConfig.ManagedBy())
		}
	}

	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		p.machineConfigs[etcdMachineName].Annotations = map[string]string{p.clusterConfig.EtcdAnnotation(): "true"}
		if etcdMachineName != controlPlaneMachineName {
			configs = append(configs, p.machineConfigs[etcdMachineName])
			if p.clusterConfig.IsManaged() {
				p.machineConfigs[etcdMachineName].SetManagedBy(p.clusterConfig.ManagedBy())
			}
		}
	}

	return configs
}

func (p *tinkerbellProvider) ValidateNewSpec(_ context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *tinkerbellProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	// TODO: implement
	return nil
}

func (p *tinkerbellProvider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	nodeGroupsToDelete := cluster.NodeGroupsToDelete(currentSpec, newSpec)
	machineDeployments := make([]string, 0, len(nodeGroupsToDelete))
	for _, nodeGroup := range nodeGroupsToDelete {
		mdName := machineDeploymentName(workloadCluster.Name, nodeGroup.Name)
		machineDeployments = append(machineDeployments, mdName)
	}
	return machineDeployments
}

func (p *tinkerbellProvider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	return nil
}
