package cloudstack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/types"
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
// Cyclomatic complexity is high. The exception below can probably be removed once the above todo is done.
// nolint:gocyclo
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, clusterSpec *cluster.Spec) error {
	controlPlaneMachineConfig := controlPlaneMachineConfig(clusterSpec)
	if controlPlaneMachineConfig == nil {
		return fmt.Errorf("cannot find CloudStackMachineConfig %v for control plane", clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}

	// validate template field name contains cluster kubernetes version for the control plane machine.
	if err := v.validateTemplateMatchesKubernetesVersion(ctx, controlPlaneMachineConfig.Spec.Template.Name, string(clusterSpec.Cluster.Spec.KubernetesVersion)); err != nil {
		return fmt.Errorf("machine config %s validation failed: %v", controlPlaneMachineConfig.Name, err)
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := etcdMachineConfig(clusterSpec)
		if etcdMachineConfig == nil {
			return fmt.Errorf("cannot find CloudStackMachineConfig %v for etcd machines", clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
		}
		// validate template field name contains cluster kubernetes version for the external etcd machine.
		if err := v.validateTemplateMatchesKubernetesVersion(ctx, etcdMachineConfig.Spec.Template.Name, string(clusterSpec.Cluster.Spec.KubernetesVersion)); err != nil {
			return fmt.Errorf("machine config %s validation failed: %v", etcdMachineConfig.Name, err)
		}
	}

	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		_, ok := clusterSpec.CloudStackMachineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		if !ok {
			return fmt.Errorf("cannot find CloudStackMachineConfig %v for worker nodes", workerNodeGroupConfiguration.MachineGroupRef.Name)
		}

		version := string(clusterSpec.Cluster.Spec.KubernetesVersion)
		// validate template field of worker group spec with the kubernetes version of each workerNodeGroup - in case of modular upgrade.
		if workerNodeGroupConfiguration.KubernetesVersion != nil {
			version = string(*workerNodeGroupConfiguration.KubernetesVersion)
		}
		templateName := clusterSpec.CloudStackMachineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec.Template.Name
		if err := v.validateTemplateMatchesKubernetesVersion(ctx, templateName, version); err != nil {
			return fmt.Errorf("machine config %s validation failed: %v", workerNodeGroupConfiguration.Name, err)
		}
	}

	for _, machineConfig := range clusterSpec.CloudStackMachineConfigs {
		if err := v.validateMachineConfig(ctx, clusterSpec.CloudStackDatacenter, machineConfig); err != nil {
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
	host, port, err := getValidControlPlaneHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %v", err)
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

func (v *Validator) validateTemplateMatchesKubernetesVersion(ctx context.Context, templateName string, kubernetesVersionName string) error {
	// Replace 1.23, 1-23, 1_23 to 123 in the template name string.
	templateReplacer := strings.NewReplacer("-", "", ".", "", "_", "")
	template := templateReplacer.Replace(templateName)
	// Replace 1-23 to 123 in the kubernetesversion string.
	replacer := strings.NewReplacer(".", "")
	kubernetesVersion := replacer.Replace(string(kubernetesVersionName))
	// This will return an error if the template name does not contain specified kubernetes version.
	// For ex if the kubernetes version is 1.23,
	// the template name should include 1.23 or 1-23, 1_23 or 123 i.e. kubernetes-1-23-eks in the string.
	if !strings.Contains(template, kubernetesVersion) {
		return fmt.Errorf("missing kube version from the machine config template name: template=%s, version=%s", templateName, string(kubernetesVersionName))
	}
	return nil
}

// ValidateSecretsUnchanged checks the secret to see if it has not been changed.
func (v *Validator) ValidateSecretsUnchanged(ctx context.Context, cluster *types.Cluster, execConfig *decoder.CloudStackExecConfig, client ProviderKubectlClient) error {
	for _, profile := range execConfig.Profiles {
		secret, err := client.GetSecretFromNamespace(ctx, cluster.KubeconfigFile, profile.Name, constants.EksaSystemNamespace)
		if apierrors.IsNotFound(err) {
			// When the secret is not found we allow for new secrets
			continue
		}
		if err != nil {
			return fmt.Errorf("getting secret for profile %s: %v", profile.Name, err)
		}
		if secretDifferentFromProfile(secret, profile) {
			return fmt.Errorf("profile '%s' is different from the secret", profile.Name)
		}
	}
	return nil
}

func secretDifferentFromProfile(secret *corev1.Secret, profile decoder.CloudStackProfileConfig) bool {
	return string(secret.Data[decoder.APIUrlKey]) != profile.ManagementUrl ||
		string(secret.Data[decoder.APIKeyKey]) != profile.ApiKey ||
		string(secret.Data[decoder.SecretKeyKey]) != profile.SecretKey ||
		string(secret.Data[decoder.VerifySslKey]) != profile.VerifySsl
}
