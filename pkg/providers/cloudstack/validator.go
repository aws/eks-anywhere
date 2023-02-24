package cloudstack

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type Validator struct {
	cmk         ProviderCmkClient
	netClient   networkutils.NetClient
	skipIpCheck bool
}

func NewValidator(cmk ProviderCmkClient, netClient networkutils.NetClient, skipIpCheck bool) *Validator {
	return &Validator{
		cmk:         cmk,
		netClient:   netClient,
		skipIpCheck: skipIpCheck,
	}
}

type localAvailabilityZone struct {
	*anywherev1.CloudStackAvailabilityZone
	ZoneId   string
	DomainId string
}

// ProviderCmkClient defines the methods used by Cmk as a separate interface to be mockable when injected into other objects.
type ProviderCmkClient interface {
	GetManagementApiEndpoint(profile string) (string, error)
	ValidateServiceOfferingPresent(ctx context.Context, profile string, zoneId string, serviceOffering anywherev1.CloudStackResourceIdentifier) error
	ValidateDiskOfferingPresent(ctx context.Context, profile string, zoneId string, diskOffering anywherev1.CloudStackResourceDiskOffering) error
	ValidateTemplatePresent(ctx context.Context, profile string, domainId string, zoneId string, account string, template anywherev1.CloudStackResourceIdentifier) error
	ValidateAffinityGroupsPresent(ctx context.Context, profile string, domainId string, account string, affinityGroupIds []string) error
	ValidateZoneAndGetId(ctx context.Context, profile string, zone anywherev1.CloudStackZone) (string, error)
	ValidateNetworkPresent(ctx context.Context, profile string, domainId string, network anywherev1.CloudStackResourceIdentifier, zoneId string, account string) error
	ValidateDomainAndGetId(ctx context.Context, profile string, domain string) (string, error)
	ValidateAccountPresent(ctx context.Context, profile string, account string, domainId string) error
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

// TODO: dry out machine configs validations.
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, cloudStackClusterSpec *Spec) error {
	controlPlaneMachineConfig := cloudStackClusterSpec.controlPlaneMachineConfig()
	if controlPlaneMachineConfig == nil {
		return fmt.Errorf("cannot find CloudStackMachineConfig %v for control plane", cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}

	if cloudStackClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := cloudStackClusterSpec.etcdMachineConfig()
		if etcdMachineConfig == nil {
			return fmt.Errorf("cannot find CloudStackMachineConfig %v for etcd machines", cloudStackClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
		}
		if etcdMachineConfig.Spec.Template != controlPlaneMachineConfig.Spec.Template {
			return fmt.Errorf("control plane and etcd machines must have the same template specified")
		}
	}

	for _, workerNodeGroupConfiguration := range cloudStackClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineConfig, ok := cloudStackClusterSpec.machineConfigsLookup[workerNodeGroupConfiguration.MachineGroupRef.Name]
		if !ok {
			return fmt.Errorf("cannot find CloudStackMachineConfig %v for worker nodes", workerNodeGroupConfiguration.MachineGroupRef.Name)
		}
		if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
			return fmt.Errorf("control plane and worker nodes must have the same template specified")
		}
	}

	err := v.setDefaultAndValidateControlPlaneHostPort(cloudStackClusterSpec)
	if err != nil {
		return fmt.Errorf("validating controlPlaneConfiguration.Endpoint.Host: %v", err)
	}

	for _, machineConfig := range cloudStackClusterSpec.machineConfigsLookup {
		if len(machineConfig.Spec.Users) <= 0 {
			machineConfig.Spec.Users = []anywherev1.UserConfiguration{{}}
		}
		if len(machineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
			machineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
		}
		if err = v.validateMachineConfig(ctx, cloudStackClusterSpec.datacenterConfig, machineConfig); err != nil {
			return fmt.Errorf("machine config %s validation failed: %v", machineConfig.Name, err)
		}
	}

	logger.MarkPass("Validated cluster Machine Configs")

	return nil
}

func (v *Validator) ValidateControlPlaneEndpointUniqueness(endpoint string) error {
	if v.skipIpCheck {
		logger.Info("Skipping control plane endpoint uniqueness check")
		return nil
	}
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint - not in host:port format: %v", err)
	}
	if networkutils.IsPortInUse(v.netClient, host, port) {
		return fmt.Errorf("endpoint <%s> is already in use", endpoint)
	}
	return nil
}

func (v *Validator) validateMachineConfig(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig, machineConfig *anywherev1.CloudStackMachineConfig) error {
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
		if machineConfig.Spec.DiskOffering != nil && (len(machineConfig.Spec.DiskOffering.Id) > 0 || len(machineConfig.Spec.DiskOffering.Name) > 0) {
			if err := v.cmk.ValidateDiskOfferingPresent(ctx, az.CredentialsRef, zoneId, *machineConfig.Spec.DiskOffering); err != nil {
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

// setDefaultAndValidateControlPlaneHostPort checks the input host to see if it is a valid hostname. If it's valid, it checks the port
// to see if the default port should be used and sets it.
func (v *Validator) setDefaultAndValidateControlPlaneHostPort(cloudStackClusterSpec *Spec) error {
	pHost := cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	_, port, err := net.SplitHostPort(pHost)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			port = controlEndpointDefaultPort
			cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = fmt.Sprintf("%s:%s",
				cloudStackClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
				controlEndpointDefaultPort)
		} else {
			return fmt.Errorf("host %s is invalid: %v", pHost, err.Error())
		}
	}
	if !networkutils.IsPortValid(port) {
		return fmt.Errorf("host %s has an invalid port", pHost)
	}
	return nil
}
