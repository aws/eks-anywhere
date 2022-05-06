package tinkerbell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"

	tinkhardware "github.com/tinkerbell/tink/protos/hardware"
	tinkworkflow "github.com/tinkerbell/tink/protos/workflow"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
)

type Validator struct {
	tink      ProviderTinkClient
	netClient networkutils.NetClient
	pbnj      ProviderPbnjClient
}

func NewValidator(tink ProviderTinkClient, netClient networkutils.NetClient, pbnjClient ProviderPbnjClient) *Validator {
	return &Validator{
		tink:      tink,
		netClient: netClient,
		pbnj:      pbnjClient,
	}
}

func (v *Validator) ValidateTinkerbellConfig(ctx context.Context, datacenterConfig *v1alpha1.TinkerbellDatacenterConfig) error {
	if err := v.validateTinkerbellIP(ctx, datacenterConfig.Spec.TinkerbellIP); err != nil {
		return err
	}

	if err := v.validateTinkerbellCertURL(ctx, datacenterConfig.Spec.TinkerbellCertURL); err != nil {
		return err
	}

	if err := v.validateTinkerbellHegelURL(ctx, datacenterConfig.Spec.TinkerbellHegelURL); err != nil {
		return err
	}

	if err := v.validateTinkerbellGRPCAuth(ctx, datacenterConfig.Spec.TinkerbellGRPCAuth); err != nil {
		return err
	}

	if err := v.validateTinkerbellPBnJGRPCAuth(ctx, datacenterConfig.Spec.TinkerbellPBnJGRPCAuth); err != nil {
		return err
	}
	logger.MarkPass("Tinkerbell Config is valid")

	return nil
}

// TODO: dry out machine configs validations
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, tinkerbellClusterSpec *spec) error {
	// TODO: move this to api Cluster validations
	if len(tinkerbellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}

	if tinkerbellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for control plane")
	}

	controlPlaneMachineConfig := tinkerbellClusterSpec.controlPlaneMachineConfig()
	if controlPlaneMachineConfig == nil {
		return fmt.Errorf("cannot find tinkerbellClusterSpec %v for control plane", tinkerbellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}
	if len(controlPlaneMachineConfig.Spec.OSFamily) <= 0 {
		return errors.New("tinkerbellMachineConfig osFamily for control plane is not set or is empty")
	}

	if tinkerbellClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for worker nodes")
	}

	workerNodeGroupMachineConfig := tinkerbellClusterSpec.firstWorkerMachineConfig()
	if workerNodeGroupMachineConfig == nil {
		return fmt.Errorf("cannot find tinkerbellMachineConfig %v for worker nodes", tinkerbellClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name)
	}
	if len(workerNodeGroupMachineConfig.Spec.OSFamily) <= 0 {
		return errors.New("tinkerbellMachineConfig osFamily for worker nodes is not set or is empty")
	}

	// TODO: move this to api Cluster validations
	controlPlaneEndpointIp := tinkerbellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	if err := networkutils.ValidateIP(controlPlaneEndpointIp); err != nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host %s", err)
	}

	if controlPlaneMachineConfig.Spec.OSFamily != v1alpha1.Ubuntu {
		return fmt.Errorf("control plane osFamily: %s is not supported, please use %s", controlPlaneMachineConfig.Spec.OSFamily, v1alpha1.Ubuntu)
	}

	if workerNodeGroupMachineConfig.Spec.OSFamily != v1alpha1.Ubuntu {
		return fmt.Errorf("worker node osFamily: %s is not supported, please use %s", workerNodeGroupMachineConfig.Spec.OSFamily, v1alpha1.Ubuntu)
	}

	if controlPlaneMachineConfig.Spec.OSFamily != workerNodeGroupMachineConfig.Spec.OSFamily {
		return errors.New("control plane and worker nodes must have the same osFamily specified")
	}

	for _, machineConfig := range tinkerbellClusterSpec.machineConfigsLookup {
		if machineConfig.Namespace != tinkerbellClusterSpec.Cluster.Namespace {
			return errors.New("TinkerbellMachineConfig and Cluster objects must have the same namespace specified")
		}
	}

	if tinkerbellClusterSpec.datacenterConfig.Namespace != tinkerbellClusterSpec.Cluster.Namespace {
		return errors.New("TinkerbellDatacenterConfig and Cluster objects must have the same namespace specified")
	}

	logger.MarkPass("Machine Configs are valid")

	return nil
}

func (v *Validator) ValidateHardwareCatalogue(ctx context.Context, catalogue hardware.Catalogue, hardwares []*tinkhardware.Hardware, skipPowerActions, force bool) error {
	tinkHardwareMap := getHardwareMap(hardwares)

	workflows, err := v.tink.GetWorkflow(ctx)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	tinkWorkflowMap, err := getWorkflowMap(workflows)
	if err != nil {
		return fmt.Errorf("validating if the workflow exist for the given list of hardwares %v", err)
	}

	if err := catalogue.ValidateHardware(skipPowerActions, force, tinkHardwareMap, tinkWorkflowMap); err != nil {
		return fmt.Errorf("failed validating Hardware BMC refs in hardware config: %v", err)
	}

	if !skipPowerActions {
		if err := catalogue.ValidateBMC(); err != nil {
			return fmt.Errorf("failed validating BMCs in hardware config: %v", err)
		}

		if err := catalogue.ValidateBmcSecretRefs(); err != nil {
			return fmt.Errorf("failed validating Secrets in hardware config: %v", err)
		}

		if err := v.ValidateBMCSecretCreds(ctx, catalogue); err != nil {
			return fmt.Errorf("failed validating BMC connection in hardware config: %v", err)
		}
		logger.MarkPass("BMC connectivity validated")
	}

	logger.MarkPass("Hardware Config file validated")

	return nil
}

func (v *Validator) ValidateBMCSecretCreds(ctx context.Context, catalogue hardware.Catalogue) error {
	for index, bmc := range catalogue.BMCs {
		bmcInfo := pbnj.BmcSecretConfig{
			Host:     bmc.Spec.Host,
			Username: string(catalogue.Secrets[index].Data["username"]),
			Password: string(catalogue.Secrets[index].Data["password"]),
			Vendor:   bmc.Spec.Vendor,
		}
		if _, err := v.pbnj.GetPowerState(ctx, bmcInfo); err != nil {
			return fmt.Errorf("failed to connect to BMC (address=%s): %v", bmcInfo.Host, err)
		}
	}

	return nil
}

func (v *Validator) ValidateAndPopulateTemplateForUpgrade(ctx context.Context, datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, templateConfig *v1alpha1.TinkerbellTemplateConfig, imgUrl string, k8sVersion string) error {
	// @TODO: Add any additional template validations here

	err := v.ValidateAndPopulateTemplate(ctx, datacenterConfig, templateConfig)
	if err != nil {
		return err
	}

	return nil
}

func (v *Validator) ValidateAndPopulateTemplate(ctx context.Context, datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, templateConfig *v1alpha1.TinkerbellTemplateConfig) error {
	for _, task := range templateConfig.Spec.Template.Tasks {
		for _, action := range task.Actions {
			// set metadata_urls to Hegel URL provided in the cluster config file
			// TODO(pjshah):: add more validations around metadata_url field and parse
			if action.Name == "add-tink-cloud-init-config" {
				metadataList := make([]string, 0)
				metadataList = append(metadataList, datacenterConfig.Spec.TinkerbellHegelURL)

				var content map[string]interface{}
				err := yaml.Unmarshal([]byte(action.Environment["CONTENTS"]), &content)
				if err != nil {
					return fmt.Errorf("failed to unmarshal %s action data: %v", action.Name, err)
				}

				if _, ok := content["datasource"]; ok {
					ec2 := content["datasource"].(map[string]interface{})
					if _, ok := ec2["Ec2"]; ok {
						metadataUrl := ec2["Ec2"].(map[string]interface{})
						if _, ok := metadataUrl["metadata_urls"]; ok {
							metadataUrl["metadata_urls"] = metadataList

							patchedContent, err := yaml.Marshal(&content)
							if err != nil {
								return fmt.Errorf("failed to marshal %s action data: %v", action.Name, err)
							}
							action.Environment["CONTENTS"] = string(patchedContent)
						}
					}
				}

			}
		}
	}

	return nil
}

// ValidateMinHardwareAvailableForCreate ensures there is sufficient hardware registered relative to the
// sum of requested control plane, etcd and worker node counts.
// The system requires hardware >= to requested provisioning.
// ValidateMinHardwareAvailableForCreate requires v.ValidateHardwareConfig() to be called first.
func (v *Validator) ValidateMinHardwareAvailableForCreate(spec v1alpha1.ClusterSpec, catalogue hardware.Catalogue) error {
	// ValidateMinHardwareAvailableForCreate relies on v.hardwareConfig being valid. A call to
	// v.ValidateHardwareConfig parses the hardware config file. Consequently, we need to validate the hardware config
	// prior to calling ValidateMinHardwareAvailableForCreate. We should decouple validation including
	// isolation of io in the parsing of hardware config.

	requestedNodesCount := spec.ControlPlaneConfiguration.Count +
		sumWorkerNodeCounts(spec.WorkerNodeGroupConfigurations)

	// Optional external etcd configuration.
	if spec.ExternalEtcdConfiguration != nil {
		requestedNodesCount += spec.ExternalEtcdConfiguration.Count
	}

	if len(catalogue.Hardware) < requestedNodesCount {
		return fmt.Errorf(
			"have %v tinkerbell hardware; cluster spec requires >= %v hardware",
			len(catalogue.Hardware),
			requestedNodesCount,
		)
	}

	return nil
}

// ValidateMinHardwareAvailableForUpgrade ensures there is sufficient hardware registered relative to the
// sum of requested control plane, etcd and worker node counts.
// The system requires hardware >= to requested provisioning.
// ValidateMinHardwareAvailableForUpgrade requires v.ValidateHardwareConfig() to be called first.
func (v *Validator) ValidateMinHardwareAvailableForUpgrade(spec v1alpha1.ClusterSpec, maxSurge int) error {
	// @TODO: Add # of available hardware validation here

	// ValidateMinHardwareAvailableForUpgrade relies on v.hardwareConfig being valid. A call to
	// v.ValidateHardwareConfig parses the hardware config file. Consequently, we need to validate the hardware config
	// prior to calling ValidateMinTinkerbellHardwareAvailableForUpgrade. We should decouple validation including
	// isolation of io in the parsing of hardware config.

	/* requestedNodesCount := spec.ControlPlaneConfiguration.Count + maxSurge +
		sumWorkerNodeCounts(spec.WorkerNodeGroupConfigurations) + maxSurge

	// Optional external etcd configuration.
	if spec.ExternalEtcdConfiguration != nil {
		requestedNodesCount += spec.ExternalEtcdConfiguration.Count + maxSurge
	}

	if len(v.hardwareConfig.Hardwares) < requestedNodesCount {
		return fmt.Errorf(
			"have %v tinkerbell hardware; cluster spec requires >= %v hardware",
			len(v.hardwareConfig.Hardwares),
			requestedNodesCount,
		)
	} */

	return nil
}

// ValidateMachinesPoweredOff validates the hardware submitted for provisioning is powered off.
func (v *Validator) ValidateMachinesPoweredOff(ctx context.Context, catalogue hardware.Catalogue) error {
	secrets := make(map[string]corev1.Secret)
	for _, s := range catalogue.Secrets {
		secrets[s.Name] = s
	}

	var poweredOnHosts []string
	for _, bmc := range catalogue.BMCs {
		secret := secrets[bmc.Spec.AuthSecretRef.Name]
		state, err := v.pbnj.GetPowerState(ctx, pbnj.BmcSecretConfig{
			Host:     bmc.Spec.Host,
			Username: string(secret.Data["username"]),
			Password: string(secret.Data["password"]),
			Vendor:   bmc.Spec.Vendor,
		})
		if err != nil {
			return fmt.Errorf("retrieving power state (bmc host = '%v'): %w", bmc.Spec.Host, err)
		}

		if state != pbnj.PowerStateOff {
			poweredOnHosts = append(poweredOnHosts, bmc.Spec.Host)
		}
	}

	if len(poweredOnHosts) > 0 {
		return fmt.Errorf(
			"expected host machines to be powered off; use --force-cleanup flag to power off machine: bmc hosts = %v",
			poweredOnHosts,
		)
	}

	return nil
}

func (v *Validator) validateControlPlaneIpUniqueness(tinkerBellClusterSpec *spec) error {
	ip := tinkerBellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	if networkutils.IsIPInUse(v.netClient, ip) {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host <%s> is already in use, please provide a unique IP", ip)
	}

	logger.MarkPass("Cluster  controlPlaneConfiguration host IP available")
	return nil
}

func (v *Validator) validateTinkerbellIP(ctx context.Context, ip string) error {
	// check if tinkerbellIP is valid
	if err := networkutils.ValidateIP(ip); err != nil {
		return fmt.Errorf("cluster tinkerbellDatacenterConfig.tinkerbellIP %s", err)
	}
	return nil
}

func (v *Validator) validateTinkerbellCertURL(ctx context.Context, tinkerbellCertURL string) error {
	if tinkerbellCertURL == "" {
		return fmt.Errorf("tinkerbellCertURL is required")
	}

	if err := validateURLHostIp(tinkerbellCertURL); err != nil {
		return fmt.Errorf("invalid tinkerbellCertURL URL: %v", err)
	}

	return nil
}

func (v *Validator) validateTinkerbellHegelURL(ctx context.Context, tinkerbellHegelURL string) error {
	if tinkerbellHegelURL == "" {
		return fmt.Errorf("tinkerbellHegelURL is required")
	}

	if err := validateURLHostIp(tinkerbellHegelURL); err != nil {
		return fmt.Errorf("invalid tinkerbellHegelURL URL: %v", err)
	}

	return nil
}

func (v *Validator) validateTinkerbellGRPCAuth(ctx context.Context, tinkerbellGRPCAuth string) error {
	if tinkerbellGRPCAuth == "" {
		return fmt.Errorf("tinkerbellGRPCAuth is required")
	}

	if err := validateAddressWithPort(tinkerbellGRPCAuth); err != nil {
		return fmt.Errorf("invalid grpc authority: %v", err)
	}

	return nil
}

func (v *Validator) validateTinkerbellPBnJGRPCAuth(ctx context.Context, tinkerbellPBnJGRPCAuth string) error {
	if tinkerbellPBnJGRPCAuth == "" {
		return fmt.Errorf("tinkerbellPBnJGRPCAuth is required")
	}

	if err := validateAddressWithPort(tinkerbellPBnJGRPCAuth); err != nil {
		return fmt.Errorf("invalid pbnj authority: %v", err)
	}

	return nil
}

func sumWorkerNodeCounts(nodes []v1alpha1.WorkerNodeGroupConfiguration) int {
	var requestedNodesCount int
	for _, workerSpec := range nodes {
		requestedNodesCount += workerSpec.Count
	}
	return requestedNodesCount
}

func validateURLHostIp(urlAdress string) error {
	if err := validateURL(urlAdress); err != nil {
		return err
	}

	hostPort, err := url.Parse(urlAdress)
	if err != nil {
		return fmt.Errorf("invalid url: %v", err)
	}

	if err := validateAddressWithPort(hostPort.Host); err != nil {
		return fmt.Errorf("invalid url: %v", err)
	}

	return nil
}

func validateURL(urlAddress string) error {
	_, err := url.ParseRequestURI(urlAddress)
	if err != nil {
		return fmt.Errorf("invalid url: %v", urlAddress)
	}

	return nil
}

// validateAddressWithPort validates address is a hostname:port combination. The port is required.
func validateAddressWithPort(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	if host == "" {
		return fmt.Errorf("missing host: %v", address)
	}

	if !networkutils.IsPortValid(port) {
		return fmt.Errorf("invalid port: %v", address)
	}

	return nil
}

// getHardwareMap returns all the hardwares on the tinkerbell stack in the form of map with hardware uuid as a key
func getHardwareMap(hardwareList []*tinkhardware.Hardware) map[string]*tinkhardware.Hardware {
	hardwareMap := make(map[string]*tinkhardware.Hardware)

	for _, data := range hardwareList {
		hardwareMap[data.GetId()] = data
	}

	return hardwareMap
}

// getWorkflowMap returns all the workflows on the tinkerbell stack in the form of map with mac address as a key
func getWorkflowMap(workflowList []*tinkworkflow.Workflow) (map[string]*tinkworkflow.Workflow, error) {
	workflowMap := make(map[string]*tinkworkflow.Workflow)

	for _, data := range workflowList {
		var macAddress map[string]string

		if err := json.Unmarshal([]byte(data.GetHardware()), &macAddress); err != nil {
			return nil, fmt.Errorf("unmarshling workflow data: %v", err)
		}
		for _, mac := range macAddress {
			workflowMap[mac] = data
		}
	}

	return workflowMap, nil
}
