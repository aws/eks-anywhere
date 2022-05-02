package tinkerbell

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	tinkhardware "github.com/tinkerbell/tink/protos/hardware"
	tinkworkflow "github.com/tinkerbell/tink/protos/workflow"
	corev1 "k8s.io/api/core/v1"
	errorutil "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
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
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

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
	hardwareConfigFile    string
	validator             *Validator
	writer                filewriter.FileWriter
	keyGenerator          SSHAuthKeyGenerator

	skipIpCheck      bool
	skipPowerActions bool
	setupTinkerbell  bool
	force            bool
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
	hardwareConfigFile string,
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
		hardwareConfigFile,
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
	hardwareConfigFile string,
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
		hardwareConfigFile: hardwareConfigFile,
		validator:          NewValidator(providerTinkClient, netClient, hardware.HardwareConfig{}, pbnjClient),
		writer:             writer,

		// (chrisdoherty4) We're hard coding the dependency and monkey patching in testing because the provider
		// isn't very testable right now and we already have tests in the `tinkerbell` package so can monkey patch
		// directly. This is very much a hack for testability.
		keyGenerator: common.SshAuthKeyGenerator{},

		// Behavioral flags.
		skipIpCheck:      skipIpCheck,
		skipPowerActions: skipPowerActions,
		setupTinkerbell:  setupTinkerbell,
		force:            force,
	}
}

func (p *tinkerbellProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	// Adding proxy environment vars to the bootstrap cluster
	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, p.datacenterConfig.Spec.TinkerbellIP)
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

func (p *tinkerbellProvider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *tinkerbellProvider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO: figure out if we need something else here
	hardwareSpec, err := p.validator.hardwareConfig.HardwareSpecMarshallable()
	if err != nil {
		return fmt.Errorf("failed marshalling resources for hardware spec: %v", err)
	}
	err = p.providerKubectlClient.ApplyKubeSpecFromBytesForce(ctx, cluster, hardwareSpec)
	if err != nil {
		return fmt.Errorf("applying hardware yaml: %v", err)
	}
	return nil
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

func (p *tinkerbellProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range p.machineConfigs {
		if err := p.providerKubectlClient.DeleteEksaMachineConfig(ctx, eksaTinkerbellDatacenterResourceType, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaDatacenterConfig(ctx, eksaTinkerbellMachineResourceType, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, p.datacenterConfig.Namespace)
}

func (p *tinkerbellProvider) PostClusterDeleteValidate(ctx context.Context, managementCluster *types.Cluster) error {
	// We want to validate cluster nodes are powered off.
	// We wait on BMC status.powerState to check for power off.
	bmcRefs := make([]string, 0, len(p.hardwares))
	for _, hw := range p.hardwares {
		bmcRefs = append(bmcRefs, hw.Spec.BmcRef)
	}

	// TODO (pokearu): The retry logic can be substituted by changing GetBmcsPowerState to use kubectl wait --for
	// In the current version of kubectl in EKSA --for does not support jsonpath.
	err := retrier.Retry(10, 10*time.Second, func() error {
		powerStates, err := p.providerKubectlClient.GetBmcsPowerState(ctx, bmcRefs, managementCluster.KubeconfigFile, constants.EksaSystemNamespace)
		if err != nil {
			return fmt.Errorf("retrieving bmc power state: %w", err)
		}

		for _, state := range powerStates {
			if !strings.Contains(state, bmcStatePowerActionHardoff) {
				return fmt.Errorf("bmc current power state '%s'; expected power state '%s'", state, bmcStatePowerActionHardoff)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func ensureMachineConfigsHaveAtLeast1User(machines map[string]*v1alpha1.TinkerbellMachineConfig) {
	for _, machine := range machines {
		if len(machine.Spec.Users) == 0 {
			machine.Spec.Users = []v1alpha1.UserConfiguration{{Name: defaultUsername}}
		}
	}
}

func extractUserConfigurationsWithoutSshKeys(machines map[string]*v1alpha1.TinkerbellMachineConfig) []*v1alpha1.UserConfiguration {
	var users []*v1alpha1.UserConfiguration

	for _, machine := range machines {
		if len(machine.Spec.Users[0].SshAuthorizedKeys) == 0 || len(machine.Spec.Users[0].SshAuthorizedKeys[0]) == 0 {
			users = append(users, &machine.Spec.Users[0])
		}
	}

	return users
}

func applySshKeyToUsers(users []*v1alpha1.UserConfiguration, key string) {
	for _, user := range users {
		if len(user.SshAuthorizedKeys) == 0 {
			user.SshAuthorizedKeys = make([]string, 1)
		}

		user.SshAuthorizedKeys[0] = key
	}
}

func stripCommentsFromSshKeys(machines map[string]*v1alpha1.TinkerbellMachineConfig) error {
	for _, machine := range machines {
		key, err := common.StripSshAuthorizedKeyComment(machine.Spec.Users[0].SshAuthorizedKeys[0])
		if err != nil {
			return fmt.Errorf("TinkerbellMachineConfig name=%v: %v", machine.Name, err)
		}
		machine.Spec.Users[0].SshAuthorizedKeys[0] = key
	}

	return nil
}

func (p *tinkerbellProvider) configureSshKeys() error {
	ensureMachineConfigsHaveAtLeast1User(p.machineConfigs)

	users := extractUserConfigurationsWithoutSshKeys(p.machineConfigs)
	if len(users) > 0 {
		publicAuthorizedKey, err := p.keyGenerator.GenerateSSHAuthKey(p.writer)
		if err != nil {
			return err
		}

		applySshKeyToUsers(users, publicAuthorizedKey)
	}

	if err := stripCommentsFromSshKeys(p.machineConfigs); err != nil {
		return fmt.Errorf("stripping ssh key comment: %v", err)
	}

	return nil
}

func (p *tinkerbellProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The tinkerbell infrastructure provider is still in development and should not be used in production")

	hardware, err := p.providerTinkClient.GetHardware(ctx)
	if err != nil {
		return fmt.Errorf("retrieving tinkerbell hardware: %v", err)
	}
	logger.MarkPass("Connected to tinkerbell stack")

	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	tinkerbellClusterSpec := newSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	if err := p.configureSshKeys(); err != nil {
		return err
	}

	// ValidateHardwareConfig performs a lazy load of hardware configuration. Given subsequent steps need the hardware
	// read into memory it needs to be done first. It also needs connection to
	// Tinkerbell steps to verify hardware availability on the stack
	if err := p.validator.ValidateHardwareConfig(ctx, p.hardwareConfigFile, hardware, p.skipPowerActions, p.force); err != nil {
		return err
	}

	if p.force {
		if err := p.setHardwareStateToProvisining(ctx); err != nil {
			return err
		}

		if !p.skipPowerActions {
			if err := p.setMachinesToPXEBoot(ctx); err != nil {
				return err
			}
		}

		if err := p.scrubWorkflowsFromTinkerbell(ctx, p.validator.hardwareConfig.Hardwares, hardware); err != nil {
			return err
		}
	} else if !p.skipPowerActions {
		if err := p.validator.ValidateMachinesPoweredOff(ctx); err != nil {
			return fmt.Errorf("validating machines are powered off: %w", err)
		}
	}

	if err := p.validator.ValidateTinkerbellConfig(ctx, tinkerbellClusterSpec.datacenterConfig); err != nil {
		return err
	}

	if err := p.validator.ValidateClusterMachineConfigs(ctx, tinkerbellClusterSpec); err != nil {
		return err
	}

	if err := p.validator.ValidateAndPopulateTemplate(ctx, tinkerbellClusterSpec.datacenterConfig, tinkerbellClusterSpec.Spec.TinkerbellTemplateConfigs[tinkerbellClusterSpec.controlPlaneMachineConfig().Spec.TemplateRef.Name]); err != nil {
		return fmt.Errorf("failed validating control plane template config: %v", err)
	}

	if err := p.validator.ValidateAndPopulateTemplate(ctx, tinkerbellClusterSpec.datacenterConfig, tinkerbellClusterSpec.Spec.TinkerbellTemplateConfigs[tinkerbellClusterSpec.firstWorkerMachineConfig().Spec.TemplateRef.Name]); err != nil {
		return fmt.Errorf("failed validating worker node template config: %v", err)
	}

	if err := p.validator.ValidateMinimumRequiredTinkerbellHardwareAvailable(tinkerbellClusterSpec.Cluster.Spec); err != nil {
		return fmt.Errorf("minimum hardware not available: %v", err)
	}

	if !p.skipIpCheck {
		if err := p.validator.validateControlPlaneIpUniqueness(tinkerbellClusterSpec); err != nil {
			return err
		}
	} else {
		logger.Info("Skipping check for whether control plane ip is in use")
	}

	return nil
}

func (p *tinkerbellProvider) setHardwareStateToProvisining(ctx context.Context) error {
	for _, hardware := range p.validator.hardwareConfig.Hardwares {
		tinkHardware, err := p.providerTinkClient.GetHardwareByUuid(ctx, hardware.Spec.ID)
		if err != nil {
			return fmt.Errorf("getting hardware with UUID '%s': %v", hardware.Spec.ID, err)
		}

		if tinkHardware.Metadata.State != Provisioning {
			tinkHardware.Metadata.State = Provisioning

			patchedHardware, err := json.Marshal(tinkHardware)
			if err != nil {
				return fmt.Errorf("marshaling hardware %s: %v", tinkHardware.Id, err)
			}

			logger.Info(fmt.Sprintf("Attempting to set state of hardware '%s' to '%s'", tinkHardware.Id, Provisioning))

			if err := p.providerTinkClient.PushHardware(ctx, patchedHardware); err != nil {
				return fmt.Errorf("patching hardware state: %v", err)
			}
		}
	}
	return nil
}

// setMachinesToPXEBoot iterates over all p.validator.hardwareConfig.Bmcs and instructs them to turn off
// and boot from PXE and turn on.
func (p *tinkerbellProvider) setMachinesToPXEBoot(ctx context.Context) error {
	// We're reaching into an unexported member of p.validator because of the lazy loading we're doing with
	// hardware configuration. This effectively defines a concrete tight coupling between the validator and the
	// Tinkerbell construct that desperately needs teething apart.

	secrets := make(map[string]corev1.Secret, len(p.validator.hardwareConfig.Secrets))
	for _, secret := range p.validator.hardwareConfig.Secrets {
		secrets[secret.Name] = secret
	}

	var errs []error
	for _, bmc := range p.validator.hardwareConfig.Bmcs {
		secret, found := secrets[bmc.Spec.AuthSecretRef.Name]
		if !found {
			errs = append(errs, fmt.Errorf("could not find bmc secret for '%v'", bmc.Name))
		}

		conf := pbnj.BmcSecretConfig{
			Host:     bmc.Spec.Host,
			Username: string(secret.Data["username"]),
			Password: string(secret.Data["password"]),
			Vendor:   bmc.Spec.Vendor,
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

		if err := p.pbnj.PowerOff(ctx, conf); err != nil {
			errs = append(errs, err)
		}

		if err := p.pbnj.SetBootDevice(ctx, conf, pbnj.BootDevicePXE); err != nil {
			errs = append(errs, err)
		}

		cancel()
	}

	return errorutil.NewAggregate(errs)
}

// scrubWorkflowsFromTinkerbell removes all workflows in the Tinkerbell stack that feature in hardware by retrieving
// hardware MAC addresses using tinkerbellHardware. tinkerbellHardware is necessary because MAC addresses aren't
// available on the Hardware object type.
func (p *tinkerbellProvider) scrubWorkflowsFromTinkerbell(ctx context.Context, hardware []tinkv1alpha1.Hardware, tinkerbellHardware []*tinkhardware.Hardware) error {
	workflows, err := p.providerTinkClient.GetWorkflow(ctx)
	if err != nil {
		return fmt.Errorf("retrieving workflows: %w", err)
	}

	hardwareMACLookup, err := createHardwareIDToMACMapping(tinkerbellHardware)
	if err != nil {
		return err
	}

	manifestHardwareMACs, err := createMACSetFromHardwareManifests(hardwareMACLookup, hardware)
	if err != nil {
		return err
	}

	workflowIDs, err := getWorkflowsIDsFromMACs(manifestHardwareMACs, workflows)
	if err != nil {
		return err
	}

	if err := p.providerTinkClient.DeleteWorkflow(ctx, workflowIDs...); err != nil {
		return fmt.Errorf("could not delete tinkerbell workflow: %v", err)
	}

	return nil
}

func createHardwareIDToMACMapping(hardware []*tinkhardware.Hardware) (map[string]string, error) {
	hardwareMACLookup := make(map[string]string)
	for _, h := range hardware {
		if len(h.Network.Interfaces) == 0 {
			return nil, fmt.Errorf("hardware manifest without interface: hardware ID = '%v'", h.Id)
		}
		hardwareMACLookup[h.Id] = h.Network.Interfaces[0].Dhcp.Mac
	}

	return hardwareMACLookup, nil
}

func createMACSetFromHardwareManifests(hardwareMACLookup map[string]string, hardware []tinkv1alpha1.Hardware) (macAddressSet, error) {
	manifestHardwareMACs := make(macAddressSet)
	for _, h := range hardware {
		mac, found := hardwareMACLookup[h.Spec.ID]
		if !found {
			return nil, fmt.Errorf("couldn't find mac address for hardware manifest: manifest hardware ID = '%v'", h.Spec.ID)
		}

		manifestHardwareMACs.Insert(mac)
	}

	return manifestHardwareMACs, nil
}

func getWorkflowsIDsFromMACs(hardwareMACs macAddressSet, workflows []*tinkworkflow.Workflow) ([]string, error) {
	var workflowIDs []string
	for _, w := range workflows {
		mac, err := macFromWorkflow(w)
		if err != nil {
			return nil, err
		}

		if hardwareMACs.Contains(mac) {
			workflowIDs = append(workflowIDs, w.Id)
		}
	}

	return workflowIDs, nil
}

func macFromWorkflow(workflow *tinkworkflow.Workflow) (string, error) {
	var data struct {
		Mac string `json:"device_1"` // Assume the hardware device data uses device_1 as the key.
	}

	if err := json.Unmarshal([]byte(workflow.Hardware), &data); err != nil {
		return "", err
	}

	return data.Mac, nil
}

type macAddressSet map[string]struct{}

func (m *macAddressSet) Contains(mac string) bool {
	_, found := (*m)[strings.ToLower(mac)]
	return found
}

func (m *macAddressSet) Insert(mac string) {
	(*m)[strings.ToLower(mac)] = struct{}{}
}

func (p *tinkerbellProvider) SetupAndValidateDeleteCluster(ctx context.Context, cluster *types.Cluster) error {
	// TODO: validations?
	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	hardwares, err := p.providerKubectlClient.GetHardwareWithLabel(ctx, tinkerbellOwnerNameLabel, cluster.KubeconfigFile, constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	filteredHws, err := filterHardwareForCluster(hardwares, cluster.Name)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	p.hardwares = filteredHws

	return nil
}

// filterHardwareForCluster filters hardware with ownerName label that contains cluster name.
func filterHardwareForCluster(hardwares []tinkv1alpha1.Hardware, clusterName string) ([]tinkv1alpha1.Hardware, error) {
	var filteredHardwareList []tinkv1alpha1.Hardware
	for _, hw := range hardwares {
		if strings.Contains(hw.Labels[tinkerbellOwnerNameLabel], clusterName) {
			filteredHardwareList = append(filteredHardwareList, hw)
		}
	}
	// Ensure that there are one or more hardware CRDs presnt in the hardware list for a cluster.
	if len(filteredHardwareList) == 0 {
		return nil, fmt.Errorf("no hardware found for cluster %s", clusterName)
	}
	return filteredHardwareList, nil
}

func (p *tinkerbellProvider) SetupAndValidateUpgradeCluster(ctx context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	// TODO: Add validations when this is supported
	return errors.New("upgrade for tinkerbell provider isn't currently supported")
}

func (p *tinkerbellProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	// TODO: implement
	return nil
}

type TinkerbellTemplateBuilder struct {
	controlPlaneMachineSpec     *v1alpha1.TinkerbellMachineConfigSpec
	datacenterSpec              *v1alpha1.TinkerbellDatacenterConfigSpec
	WorkerNodeGroupMachineSpecs map[string]v1alpha1.TinkerbellMachineConfigSpec
	etcdMachineSpec             *v1alpha1.TinkerbellMachineConfigSpec
	now                         types.NowFunc
}

func NewTinkerbellTemplateBuilder(datacenterSpec *v1alpha1.TinkerbellDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.TinkerbellMachineConfigSpec, workerNodeGroupMachineSpecs map[string]v1alpha1.TinkerbellMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &TinkerbellTemplateBuilder{
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		datacenterSpec:              datacenterSpec,
		WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		etcdMachineSpec:             etcdMachineSpec,
		now:                         now,
	}
}

func (tb *TinkerbellTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	cpTemplateConfig := clusterSpec.TinkerbellTemplateConfigs[tb.controlPlaneMachineSpec.TemplateRef.Name]
	cpTemplateString, err := cpTemplateConfig.ToTemplateString()
	if err != nil {
		return nil, fmt.Errorf("failed to get Control Plane TinkerbellTemplateConfig: %v", err)
	}

	var etcdMachineSpec v1alpha1.TinkerbellMachineConfigSpec
	var etcdTemplateString string
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *tb.etcdMachineSpec
		etcdTemplateConfig := clusterSpec.TinkerbellTemplateConfigs[tb.etcdMachineSpec.TemplateRef.Name]
		etcdTemplateString, err = etcdTemplateConfig.ToTemplateString()
		if err != nil {
			return nil, fmt.Errorf("failed to get ETCD TinkerbellTemplateConfig: %v", err)
		}
	}
	values := buildTemplateMapCP(clusterSpec, *tb.controlPlaneMachineSpec, etcdMachineSpec, cpTemplateString, etcdTemplateString)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}
	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (tb *TinkerbellTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		wTemplateConfig := clusterSpec.TinkerbellTemplateConfigs[tb.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name].TemplateRef.Name]
		wTemplateString, err := wTemplateConfig.ToTemplateString()
		if err != nil {
			return nil, fmt.Errorf("failed to get worker TinkerbellTemplateConfig: %v", err)
		}

		values := buildTemplateMapMD(clusterSpec, tb.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration, wTemplateString)
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]
		values["workerSshAuthorizedKey"] = tb.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name].Users[0].SshAuthorizedKeys[0]
		values["workerReplicas"] = workerNodeGroupConfiguration.Count

		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}
	return templater.AppendYamlResources(workerSpecs...), nil
}

func (p *tinkerbellProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForCreate(ctx, cluster, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("generating cluster api spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *tinkerbellProvider) generateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := clusterSpec.Cluster.Name

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = common.CPMachineTemplateName(clusterName, p.templateBuilder.now)
		values["controlPlaneSshAuthorizedKey"] = p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0].SshAuthorizedKeys[0]
		if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
			values["etcdSshAuthorizedKey"] = p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0].SshAuthorizedKeys[0]
		}
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

func (p *tinkerbellProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	// TODO: implement
	return nil, nil, nil
}

func (p *tinkerbellProvider) GenerateStorageClass() []byte {
	// TODO: determine if we need something else here
	return nil
}

func (p *tinkerbellProvider) GenerateMHC() ([]byte, error) {
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

func (p *tinkerbellProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *tinkerbellProvider) UpgradeNeeded(_ context.Context, _, _ *cluster.Spec, _ *types.Cluster) (bool, error) {
	// TODO: Figure out if something is needed here
	return false, nil
}

func machineDeploymentName(clusterName, nodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, nodeGroupName)
}

func buildTemplateMapCP(clusterSpec *cluster.Spec, controlPlaneMachineSpec, etcdMachineSpec v1alpha1.TinkerbellMachineConfigSpec, cpTemplateOverride, etcdTemplateOverride string) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":                  clusterSpec.Cluster.Name,
		"controlPlaneEndpointIp":       clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":         clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"controlPlaneSshAuthorizedKey": controlPlaneMachineSpec.Users[0].SshAuthorizedKeys,
		"controlPlaneSshUsername":      controlPlaneMachineSpec.Users[0].Name,
		"eksaSystemNamespace":          constants.EksaSystemNamespace,
		"format":                       format,
		"kubernetesVersion":            bundle.KubeDistro.Kubernetes.Tag,
		"kubeVipImage":                 bundle.Tinkerbell.KubeVip.VersionedImage(),
		"podCidrs":                     clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                 clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"baseRegistry":                 "", // TODO: need to get this values for creating template IMAGE_URL
		"osDistro":                     "", // TODO: need to get this values for creating template IMAGE_URL
		"osVersion":                    "", // TODO: need to get this values for creating template IMAGE_URL
		"kubernetesRepository":         bundle.KubeDistro.Kubernetes.Repository,
		"corednsRepository":            bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":               bundle.KubeDistro.CoreDNS.Tag,
		"etcdRepository":               bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                 bundle.KubeDistro.Etcd.Tag,
		"externalEtcdVersion":          bundle.KubeDistro.EtcdVersion,
		"etcdCipherSuites":             crypto.SecureCipherSuitesString(),
		"controlPlanetemplateOverride": cpTemplateOverride,
	}
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
		values["etcdTemplateOverride"] = etcdTemplateOverride
	}

	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.TinkerbellMachineConfigSpec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, workerTemplateOverride string) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":            clusterSpec.Cluster.Name,
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
		"format":                 format,
		"kubernetesVersion":      bundle.KubeDistro.Kubernetes.Tag,
		"workerNodeGroupName":    workerNodeGroupConfiguration.Name,
		"workerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys,
		"workerSshUsername":      workerNodeGroupMachineSpec.Users[0].Name,
		"workertemplateOverride": workerTemplateOverride,
	}
	return values
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
