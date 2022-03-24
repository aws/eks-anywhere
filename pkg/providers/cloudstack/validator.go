package cloudstack

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

type Validator struct {
	cmk ProviderCmkClient
}

// Taken from https://github.com/shapeblue/cloudstack/blob/08bb4ad9fea7e422c3d3ac6d52f4670b1e89eed7/api/src/main/java/com/cloud/vm/VmDetailConstants.java
// These fields should be modeled separately in eks-a and not used by the additionalDetails cloudstack VM field
var restrictedUserCustomDetails = [...]string{
	"keyboard", "cpu.corespersocket", "rootdisksize", "boot.mode", "nameonhypervisor",
	"nicAdapter", "rootDiskController", "dataDiskController", "svga.vramSize", "nestedVirtualizationFlag", "ramReservation",
	"hypervisortoolsversion", "platform", "timeoffset", "kvm.vnc.port", "kvm.vnc.address", "video.hardware", "video.ram",
	"smc.present", "firmware", "cpuNumber", "cpuSpeed", "memory", "cpuOvercommitRatio", "memoryOvercommitRatio",
	"Message.ReservedCapacityFreed.Flag", "deployvm", "SSH.PublicKey", "SSH.KeyPairNames", "password", "Encrypted.Password",
	"configDriveLocation", "nic", "network", "ip4Address", "ip6Address", "disk", "diskOffering", "configurationId",
	"keypairnames", "controlNodeLoginUser",
}

func NewValidator(cmk ProviderCmkClient) *Validator {
	return &Validator{
		cmk: cmk,
	}
}

type ProviderCmkClient interface {
	ValidateCloudStackConnection(ctx context.Context) error
	ValidateServiceOfferingPresent(ctx context.Context, zoneId string, serviceOffering anywherev1.CloudStackResourceIdentifier) error
	ValidateTemplatePresent(ctx context.Context, domainId string, zoneId string, account string, template anywherev1.CloudStackResourceIdentifier) error
	ValidateAffinityGroupsPresent(ctx context.Context, domainId string, account string, affinityGroupIds []string) error
	ValidateZonesPresent(ctx context.Context, zones []anywherev1.CloudStackZone) ([]anywherev1.CloudStackResourceIdentifier, error)
	ValidateNetworkPresent(ctx context.Context, domainId string, zoneRef anywherev1.CloudStackZone, zones []anywherev1.CloudStackResourceIdentifier, account string, multipleZone bool) error
	ValidateDomainPresent(ctx context.Context, domain string) (anywherev1.CloudStackResourceIdentifier, error)
	ValidateAccountPresent(ctx context.Context, account string, domainId string) error
}

func (v *Validator) validateCloudStackAccess(ctx context.Context) error {
	if err := v.cmk.ValidateCloudStackConnection(ctx); err != nil {
		return fmt.Errorf("failed validating connection to cloudstack: %v", err)
	}
	logger.MarkPass("Connected to server")

	return nil
}

func (v *Validator) ValidateCloudStackDatacenterConfig(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig) error {
	if len(datacenterConfig.Spec.Domain) <= 0 {
		return fmt.Errorf("CloudStackDatacenterConfig domain is not set or is empty")
	}
	if datacenterConfig.Spec.ManagementApiEndpoint == "" {
		return fmt.Errorf("CloudStackDatacenterConfig managementApiEndpoint is not set or is empty")
	}
	_, err := getHostnameFromUrl(datacenterConfig.Spec.ManagementApiEndpoint)
	if err != nil {
		return fmt.Errorf("checking management api endpoint: %v", err)
	}
	execConfig, err := decoder.ParseCloudStackSecret()
	if err != nil {
		return fmt.Errorf("parsing cloudstack secret: %v", err)
	}
	if execConfig.ManagementUrl != datacenterConfig.Spec.ManagementApiEndpoint {
		return fmt.Errorf("cloudstack secret management url (%s) differs from cluster spec management url (%s)",
			execConfig.ManagementUrl, datacenterConfig.Spec.ManagementApiEndpoint)
	}

	domain, errDomain := v.cmk.ValidateDomainPresent(ctx, datacenterConfig.Spec.Domain)
	if errDomain != nil {
		return fmt.Errorf("checking domain: %v", errDomain)
	}

	if len(datacenterConfig.Spec.Account) > 0 {
		err := v.cmk.ValidateAccountPresent(ctx, datacenterConfig.Spec.Account, domain.Id)
		if err != nil {
			return fmt.Errorf("checking account %v", err)
		}
	}

	zones, errZone := v.cmk.ValidateZonesPresent(ctx, datacenterConfig.Spec.Zones)
	if errZone != nil {
		return fmt.Errorf("checking zones %v", errZone)
	}

	for _, zone := range datacenterConfig.Spec.Zones {
		if len(zone.Network.Id) == 0 && len(zone.Network.Name) == 0 {
			return fmt.Errorf("zone network is not set or is empty")
		}
		err := v.cmk.ValidateNetworkPresent(ctx, domain.Id, zone, zones, datacenterConfig.Spec.Account, len(zones) > 1)
		if err != nil {
			return fmt.Errorf("checking network %v", err)
		}
	}

	logger.MarkPass("Datacenter validated")
	return nil
}

// TODO: dry out machine configs validations
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, cloudStackClusterSpec *Spec) error {
	var etcdMachineConfig *anywherev1.CloudStackMachineConfig

	if len(cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}
	if cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
		return fmt.Errorf("must specify machineGroupRef for control plane")
	}

	controlPlaneMachineConfig := cloudStackClusterSpec.controlPlaneMachineConfig()
	if controlPlaneMachineConfig == nil {
		return fmt.Errorf("cannot find CloudStackMachineConfig %v for control plane", cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}

	if cloudStackClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		if cloudStackClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
			return fmt.Errorf("must specify machineGroupRef for etcd machines")
		}
		etcdMachineConfig = cloudStackClusterSpec.etcdMachineConfig()
		if etcdMachineConfig == nil {
			return fmt.Errorf("cannot find CloudStackMachineConfig %v for etcd machines", cloudStackClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
		}
		if etcdMachineConfig.Spec.Template != controlPlaneMachineConfig.Spec.Template {
			return fmt.Errorf("control plane and etcd machines must have the same template specified")
		}
	}

	if cloudStackClusterSpec.datacenterConfig.Namespace != cloudStackClusterSpec.Cluster.Namespace {
		return fmt.Errorf(
			"CloudStackDatacenterConfig and Cluster objects must have the same namespace: CloudstackDatacenterConfig namespace=%s; Cluster namespace=%s",
			cloudStackClusterSpec.datacenterConfig.Namespace,
			cloudStackClusterSpec.Cluster.Namespace,
		)
	}
	if len(cloudStackClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations) > 1 {
		return fmt.Errorf("multiple worker node groups are not yet supported by the Cloudstack provider")
	}

	for _, workerNodeGroupConfiguration := range cloudStackClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if workerNodeGroupConfiguration.MachineGroupRef == nil {
			return fmt.Errorf("must specify machineGroupRef for worker nodes")
		}
		workerNodeGroupMachineConfig, ok := cloudStackClusterSpec.machineConfigsLookup[workerNodeGroupConfiguration.MachineGroupRef.Name]
		if !ok {
			return fmt.Errorf("cannot find CloudStackMachineConfig %v for worker nodes", workerNodeGroupConfiguration.MachineGroupRef.Name)
		}
		if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
			return fmt.Errorf("control plane and worker nodes must have the same template specified")
		}
	}

	isPortSpecified, err := v.validateControlPlaneHost(cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
	if err != nil {
		return fmt.Errorf("failed to validate controlPlaneConfiguration.Endpoint.Host: %v", err)
	}
	if !isPortSpecified {
		v.setDefaultControlPlanePort(cloudStackClusterSpec)
	}

	for _, machineConfig := range cloudStackClusterSpec.machineConfigsLookup {
		if machineConfig.Namespace != cloudStackClusterSpec.Cluster.Namespace {
			return fmt.Errorf(
				"CloudStackMachineConfig %s and Cluster objects must have the same namespace: CloudStackMachineConfig namespace=%s; Cluster namespace=%s",
				machineConfig.Name,
				machineConfig.Namespace,
				cloudStackClusterSpec.Cluster.Namespace,
			)
		}
		if len(machineConfig.Spec.Users) <= 0 {
			machineConfig.Spec.Users = []anywherev1.UserConfiguration{{}}
		}
		if len(machineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
			machineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
		}
		if len(machineConfig.Spec.Template.Id) == 0 && len(machineConfig.Spec.Template.Name) == 0 {
			return fmt.Errorf("template is not set for CloudStackMachineConfig %s. Default template is not supported in CloudStack, please provide a template name", machineConfig.Name)
		}
		if err = v.validateMachineConfig(ctx, cloudStackClusterSpec.datacenterConfig.Spec, machineConfig); err != nil {
			return fmt.Errorf("machine config %s validation failed: %v", machineConfig.Name, err)
		}
	}
	logger.MarkPass("Validated cluster Machine Configs")

	return nil
}

func (v *Validator) validateMachineConfig(ctx context.Context, datacenterConfigSpec anywherev1.CloudStackDatacenterConfigSpec, machineConfig *anywherev1.CloudStackMachineConfig) error {
	if len(datacenterConfigSpec.Domain) <= 0 {
		return fmt.Errorf("CloudStackDatacenterConfig domain is not set or is empty")
	}
	for _, restrictedKey := range restrictedUserCustomDetails {
		if _, found := machineConfig.Spec.UserCustomDetails[restrictedKey]; found {
			return fmt.Errorf("restricted key %s found in custom user details", restrictedKey)
		}
	}
	domain, errDomain := v.cmk.ValidateDomainPresent(ctx, datacenterConfigSpec.Domain)
	if errDomain != nil {
		return fmt.Errorf("checking domain: %v", errDomain)
	}

	zones, err := v.cmk.ValidateZonesPresent(ctx, datacenterConfigSpec.Zones)
	if err != nil {
		return fmt.Errorf("checking zones %v", err)
	}
	domainId := domain.Id
	account := datacenterConfigSpec.Account

	for _, zone := range zones {
		if err = v.cmk.ValidateTemplatePresent(ctx, domainId, zone.Id, account, machineConfig.Spec.Template); err != nil {
			return fmt.Errorf("validating template: %v", err)
		}
		if err = v.cmk.ValidateServiceOfferingPresent(ctx, zone.Id, machineConfig.Spec.ComputeOffering); err != nil {
			return fmt.Errorf("validating service offering: %v", err)
		}
	}

	if len(machineConfig.Spec.AffinityGroupIds) > 0 {
		if err = v.cmk.ValidateAffinityGroupsPresent(ctx, domainId, account, machineConfig.Spec.AffinityGroupIds); err != nil {
			return fmt.Errorf("validating affinity group ids: %v", err)
		}
	}

	return nil
}

// validateControlPlaneHost checks the input host to see if it is a valid hostname. If it's valid, it checks the port
// or returns a boolean indicating that there was no port specified, in which case the default port should be used
func (v *Validator) validateControlPlaneHost(pHost string) (bool, error) {
	_, port, err := net.SplitHostPort(pHost)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			return false, nil
		} else {
			return false, fmt.Errorf("host %s is invalid: %v", pHost, err.Error())
		}
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		return false, fmt.Errorf("host %s has an invalid port: %v", pHost, err.Error())
	}
	return true, nil
}

func (v *Validator) setDefaultControlPlanePort(cloudStackClusterSpec *Spec) {
	cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = fmt.Sprintf("%s:%s",
		cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		controlEndpointDefaultPort)
}
