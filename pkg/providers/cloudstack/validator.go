package cloudstack

import (
	"context"
	"errors"
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

type localAvailabilityZone struct {
	*anywherev1.CloudStackAvailabilityZone
	ZoneId   string
	DomainId string
}

type ProviderCmkClient interface {
	GetManagementApiEndpoint(profile string) (string, error)
	ValidateCloudStackConnection(ctx context.Context, profile string) error
	ValidateServiceOfferingPresent(ctx context.Context, profile string, zoneId string, serviceOffering anywherev1.CloudStackResourceIdentifier) error
	ValidateDiskOfferingPresent(ctx context.Context, profile string, zoneId string, diskOffering anywherev1.CloudStackResourceDiskOffering) error
	ValidateTemplatePresent(ctx context.Context, profile string, domainId string, zoneId string, account string, template anywherev1.CloudStackResourceIdentifier) error
	ValidateAffinityGroupsPresent(ctx context.Context, profile string, domainId string, account string, affinityGroupIds []string) error
	ValidateZoneAndGetId(ctx context.Context, profile string, zone anywherev1.CloudStackZone) (string, error)
	ValidateNetworkPresent(ctx context.Context, profile string, domainId string, network anywherev1.CloudStackResourceIdentifier, zoneId string, account string) error
	ValidateDomainAndGetId(ctx context.Context, profile string, domain string) (string, error)
	ValidateAccountPresent(ctx context.Context, profile string, account string, domainId string) error
}

func (v *Validator) validateCloudStackAccess(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig) error {
	refNamesToCheck := []string{}
	if len(datacenterConfig.Spec.Domain) > 0 {
		refNamesToCheck = append(refNamesToCheck, decoder.CloudStackGlobalAZ)
	}
	for _, az := range datacenterConfig.Spec.AvailabilityZones {
		refNamesToCheck = append(refNamesToCheck, az.CredentialsRef)
	}

	for _, refName := range refNamesToCheck {
		if err := v.cmk.ValidateCloudStackConnection(ctx, refName); err != nil {
			return fmt.Errorf("validating connection to cloudstack %s: %v", refName, err)
		}
	}

	logger.MarkPass(fmt.Sprintf("Connected to servers: %s", strings.Join(refNamesToCheck, ", ")))
	return nil
}

func (v *Validator) ValidateCloudStackDatacenterConfig(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig) error {
	localAvailabilityZones, err := generateLocalAvailabilityZones(ctx, datacenterConfig)
	if err != nil {
		return err
	}

	for _, az := range localAvailabilityZones {
		_, err := getHostnameFromUrl(az.ManagementApiEndpoint)
		if err != nil {
			return fmt.Errorf("checking management api endpoint: %v", err)
		}

		endpoint, err := v.cmk.GetManagementApiEndpoint(az.CredentialsRef)
		if err != nil {
			return err
		}
		if endpoint != az.ManagementApiEndpoint {
			return fmt.Errorf("cloudstack secret management url (%s) differs from cluster spec management url (%s)",
				endpoint, az.ManagementApiEndpoint)
		}

		domainId, err := v.cmk.ValidateDomainAndGetId(ctx, az.CredentialsRef, az.Domain)
		if err != nil {
			return err
		}
		az.DomainId = domainId

		if err := v.cmk.ValidateAccountPresent(ctx, az.CredentialsRef, az.Account, az.DomainId); err != nil {
			return err
		}

		zoneId, err := v.cmk.ValidateZoneAndGetId(ctx, az.CredentialsRef, az.CloudStackAvailabilityZone.Zone)
		if err != nil {
			return err
		}
		if len(az.CloudStackAvailabilityZone.Zone.Network.Id) == 0 && len(az.CloudStackAvailabilityZone.Zone.Network.Name) == 0 {
			return fmt.Errorf("zone network is not set or is empty")
		}
		if err := v.cmk.ValidateNetworkPresent(ctx, az.CredentialsRef, az.DomainId, az.CloudStackAvailabilityZone.Zone.Network, zoneId, az.Account); err != nil {
			return err
		}
	}

	logger.MarkPass("Datacenter validated")
	return nil
}

func generateLocalAvailabilityZones(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig) ([]localAvailabilityZone, error) {
	localAvailabilityZones := []localAvailabilityZone{}

	if datacenterConfig == nil {
		return nil, errors.New("CloudStack Datacenter Config is null")
	}

	if len(datacenterConfig.Spec.Domain) > 0 {
		for index, zone := range datacenterConfig.Spec.Zones {
			availabilityZone := localAvailabilityZone{
				CloudStackAvailabilityZone: &anywherev1.CloudStackAvailabilityZone{
					Name:                  fmt.Sprintf("default-az%d", index),
					CredentialsRef:        decoder.CloudStackGlobalAZ,
					Domain:                datacenterConfig.Spec.Domain,
					Account:               datacenterConfig.Spec.Account,
					ManagementApiEndpoint: datacenterConfig.Spec.ManagementApiEndpoint,
					Zone:                  zone,
				},
			}
			localAvailabilityZones = append(localAvailabilityZones, availabilityZone)
		}
	}
	for _, az := range datacenterConfig.Spec.AvailabilityZones {
		availabilityZone := localAvailabilityZone{
			CloudStackAvailabilityZone: &az,
		}
		localAvailabilityZones = append(localAvailabilityZones, availabilityZone)
	}

	if len(localAvailabilityZones) <= 0 {
		return nil, fmt.Errorf("CloudStackDatacenterConfig domain or availabilityZones is not set or is empty")
	}
	return localAvailabilityZones, nil
}

// TODO: dry out machine configs validations
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, cloudStackClusterSpec *Spec) error {
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
		etcdMachineConfig := cloudStackClusterSpec.etcdMachineConfig()
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
		if len(machineConfig.Spec.ComputeOffering.Id) == 0 && len(machineConfig.Spec.ComputeOffering.Name) == 0 {
			return fmt.Errorf("computeOffering is not set for CloudStackMachineConfig %s. Default computeOffering is not supported in CloudStack, please provide a computeOffering name or ID", machineConfig.Name)
		}
		if len(machineConfig.Spec.Template.Id) == 0 && len(machineConfig.Spec.Template.Name) == 0 {
			return fmt.Errorf("template is not set for CloudStackMachineConfig %s. Default template is not supported in CloudStack, please provide a template name or ID", machineConfig.Name)
		}
		if err, fieldName, fieldValue := machineConfig.Spec.DiskOffering.Validate(); err != nil {
			return fmt.Errorf("machine config %s validation failed: %s: %s invalid, %v", machineConfig.Name, fieldName, fieldValue, err)
		}
		if err = v.validateMachineConfig(ctx, cloudStackClusterSpec.datacenterConfig, machineConfig); err != nil {
			return fmt.Errorf("machine config %s validation failed: %v", machineConfig.Name, err)
		}
		if err = v.validateAffinityConfig(machineConfig); err != nil {
			return err
		}
	}

	logger.MarkPass("Validated cluster Machine Configs")

	return nil
}

func (v *Validator) validateAffinityConfig(machineConfig *anywherev1.CloudStackMachineConfig) error {
	if len(machineConfig.Spec.Affinity) > 0 && len(machineConfig.Spec.AffinityGroupIds) > 0 {
		return fmt.Errorf("affinity and affinityGroupIds cannot be set at the same time for CloudStackMachineConfig %s. Please provide either one of them or none", machineConfig.Name)
	}
	if len(machineConfig.Spec.Affinity) > 0 {
		if machineConfig.Spec.Affinity != "pro" && machineConfig.Spec.Affinity != "anti" && machineConfig.Spec.Affinity != "no" {
			return fmt.Errorf("invalid affinity type %s for CloudStackMachineConfig %s. Please provide \"pro\", \"anti\" or \"no\"", machineConfig.Spec.Affinity, machineConfig.Name)
		}
	}
	return nil
}

func (v *Validator) validateMachineConfig(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig, machineConfig *anywherev1.CloudStackMachineConfig) error {
	for _, restrictedKey := range restrictedUserCustomDetails {
		if _, found := machineConfig.Spec.UserCustomDetails[restrictedKey]; found {
			return fmt.Errorf("restricted key %s found in custom user details", restrictedKey)
		}
	}

	localAvailabilityZones, err := generateLocalAvailabilityZones(ctx, datacenterConfig)
	if err != nil {
		return err
	}

	for _, az := range localAvailabilityZones {
		zoneId, err := v.cmk.ValidateZoneAndGetId(ctx, az.CredentialsRef, az.CloudStackAvailabilityZone.Zone)
		if err != nil {
			return err
		}

		if err := v.cmk.ValidateTemplatePresent(ctx, az.CredentialsRef, az.DomainId, zoneId, az.Account, machineConfig.Spec.Template); err != nil {
			return fmt.Errorf("validating template: %v", err)
		}
		if err := v.cmk.ValidateServiceOfferingPresent(ctx, az.CredentialsRef, zoneId, machineConfig.Spec.ComputeOffering); err != nil {
			return fmt.Errorf("validating service offering: %v", err)
		}
		if len(machineConfig.Spec.DiskOffering.Id) > 0 || len(machineConfig.Spec.DiskOffering.Name) > 0 {
			if err := v.cmk.ValidateDiskOfferingPresent(ctx, az.CredentialsRef, zoneId, machineConfig.Spec.DiskOffering); err != nil {
				return fmt.Errorf("validating disk offering: %v", err)
			}
		}
		if len(machineConfig.Spec.AffinityGroupIds) > 0 {
			if err := v.cmk.ValidateAffinityGroupsPresent(ctx, az.CredentialsRef, az.DomainId, az.Account, machineConfig.Spec.AffinityGroupIds); err != nil {
				return fmt.Errorf("validating affinity group ids: %v", err)
			}
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
