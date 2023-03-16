package nutanix

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	"github.com/nutanix-cloud-native/prism-go-client/utils"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
	"k8s.io/apimachinery/pkg/api/resource"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

const (
	minNutanixCPUSockets   = 1
	minNutanixCPUPerSocket = 1
	minNutanixMemoryMiB    = 2048
	minNutanixDiskGiB      = 20
)

// Validator is a client to validate nutanix resources.
type Validator struct {
	httpClient    *http.Client
	certValidator crypto.TlsValidator
	clientCache   *ClientCache
}

// NewValidator returns a new validator client.
func NewValidator(clientCache *ClientCache, certValidator crypto.TlsValidator, httpClient *http.Client) *Validator {
	return &Validator{
		clientCache:   clientCache,
		certValidator: certValidator,
		httpClient:    httpClient,
	}
}

// ValidateClusterSpec validates the cluster spec.
func (v *Validator) ValidateClusterSpec(ctx context.Context, spec *cluster.Spec, creds credentials.BasicAuthCredential) error {
	fmt.Printf("ValidateClusterSpec for Nutanix datacenter %s\n", spec.NutanixDatacenter.Name)
	client, err := v.clientCache.GetNutanixClient(spec.NutanixDatacenter, creds)
	if err != nil {
		return err
	}

	if err := v.ValidateDatacenterConfig(ctx, client, spec.NutanixDatacenter); err != nil {
		return err
	}

	for _, conf := range spec.NutanixMachineConfigs {
		if err := v.ValidateMachineConfig(ctx, client, conf); err != nil {
			return fmt.Errorf("failed to validate machine config: %v", err)
		}
	}

	return nil
}

// ValidateDatacenterConfig validates the datacenter config.
func (v *Validator) ValidateDatacenterConfig(ctx context.Context, client Client, config *anywherev1.NutanixDatacenterConfig) error {
	if config.Spec.Insecure {
		logger.Info("Warning: Skipping TLS validation for insecure connection to Nutanix Prism Central; this is not recommended for production use")
	}

	if err := v.validateEndpointAndPort(config.Spec); err != nil {
		return err
	}

	if err := v.validateCredentials(ctx, client); err != nil {
		return err
	}

	if err := v.validateTrustBundleConfig(config.Spec); err != nil {
		return err
	}

	if err := v.validateCredentialRef(config); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateCredentialRef(config *anywherev1.NutanixDatacenterConfig) error {
	if config.Spec.CredentialRef == nil {
		return fmt.Errorf("credentialRef must be provided")
	}

	if config.Spec.CredentialRef.Kind != constants.SecretKind {
		return fmt.Errorf("credentialRef kind must be %s", constants.SecretKind)
	}

	if config.Spec.CredentialRef.Name == "" {
		return fmt.Errorf("credentialRef name must be provided")
	}

	return nil
}

func (v *Validator) validateEndpointAndPort(dcConf anywherev1.NutanixDatacenterConfigSpec) error {
	if !networkutils.IsPortValid(strconv.Itoa(dcConf.Port)) {
		return fmt.Errorf("nutanix prism central port %d out of range", dcConf.Port)
	}

	if dcConf.Endpoint == "" {
		return fmt.Errorf("nutanix prism central endpoint must be provided")
	}
	server := fmt.Sprintf("%s:%d", dcConf.Endpoint, dcConf.Port)
	if !strings.HasPrefix(server, "https://") {
		server = fmt.Sprintf("https://%s", server)
	}

	if _, err := v.httpClient.Get(server); err != nil {
		return fmt.Errorf("failed to reach server %s: %v", server, err)
	}

	return nil
}

func (v *Validator) validateCredentials(ctx context.Context, client Client) error {
	_, err := client.GetCurrentLoggedInUser(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateTrustBundleConfig(dcConf anywherev1.NutanixDatacenterConfigSpec) error {
	if dcConf.AdditionalTrustBundle == "" {
		return nil
	}

	return v.certValidator.ValidateCert(dcConf.Endpoint, fmt.Sprintf("%d", dcConf.Port), dcConf.AdditionalTrustBundle)
}

func (v *Validator) validateMachineSpecs(machineSpec anywherev1.NutanixMachineConfigSpec) error {
	if machineSpec.VCPUSockets < minNutanixCPUSockets {
		return fmt.Errorf("vCPU sockets %d must be greater than or equal to %d", machineSpec.VCPUSockets, minNutanixCPUSockets)
	}

	if machineSpec.VCPUsPerSocket < minNutanixCPUPerSocket {
		return fmt.Errorf("vCPUs per socket %d must be greater than or equal to %d", machineSpec.VCPUsPerSocket, minNutanixCPUPerSocket)
	}

	minNutanixMemory, err := resource.ParseQuantity(fmt.Sprintf("%dMi", minNutanixMemoryMiB))
	if err != nil {
		return err
	}

	if machineSpec.MemorySize.Cmp(minNutanixMemory) < 0 {
		return fmt.Errorf("MemorySize must be greater than or equal to %dMi", minNutanixMemoryMiB)
	}

	minNutanixDisk, err := resource.ParseQuantity(fmt.Sprintf("%dGi", minNutanixDiskGiB))
	if err != nil {
		return err
	}

	if machineSpec.SystemDiskSize.Cmp(minNutanixDisk) < 0 {
		return fmt.Errorf("SystemDiskSize must be greater than or equal to %dGi", minNutanixDiskGiB)
	}

	return nil
}

// ValidateMachineConfig validates the Prism Element cluster, subnet, and image for the machine.
func (v *Validator) ValidateMachineConfig(ctx context.Context, client Client, config *anywherev1.NutanixMachineConfig) error {
	if err := v.validateMachineSpecs(config.Spec); err != nil {
		return err
	}

	if err := v.validateClusterConfig(ctx, client, config.Spec.Cluster); err != nil {
		return err
	}

	if err := v.validateSubnetConfig(ctx, client, config.Spec.Subnet); err != nil {
		return err
	}

	if err := v.validateImageConfig(ctx, client, config.Spec.Image); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateClusterConfig(ctx context.Context, client Client, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing cluster name")
		} else {
			clusterName := *identifier.Name
			if _, err := findClusterUUIDByName(ctx, client, clusterName); err != nil {
				return fmt.Errorf("failed to find cluster with name %q: %v", clusterName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing cluster uuid")
		} else {
			clusterUUID := *identifier.UUID
			if _, err := client.GetCluster(ctx, clusterUUID); err != nil {
				return fmt.Errorf("failed to find cluster with uuid %v: %v", clusterUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid cluster identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

func (v *Validator) validateImageConfig(ctx context.Context, client Client, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing image name")
		} else {
			imageName := *identifier.Name
			if _, err := findImageUUIDByName(ctx, client, imageName); err != nil {
				return fmt.Errorf("failed to find image with name %q: %v", imageName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing image uuid")
		} else {
			imageUUID := *identifier.UUID
			if _, err := client.GetImage(ctx, imageUUID); err != nil {
				return fmt.Errorf("failed to find image with uuid %s: %v", imageUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid image identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

func (v *Validator) validateSubnetConfig(ctx context.Context, client Client, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing subnet name")
		} else {
			subnetName := *identifier.Name
			if _, err := findSubnetUUIDByName(ctx, client, subnetName); err != nil {
				return fmt.Errorf("failed to find subnet with name %s: %v", subnetName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing subnet uuid")
		} else {
			subnetUUID := *identifier.UUID
			if _, err := client.GetSubnet(ctx, subnetUUID); err != nil {
				return fmt.Errorf("failed to find subnet with uuid %s: %v", subnetUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid subnet identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

// findSubnetUUIDByName retrieves the subnet uuid by the given subnet name.
func findSubnetUUIDByName(ctx context.Context, v3Client Client, subnetName string) (*string, error) {
	res, err := v3Client.ListSubnet(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", subnetName)),
	})
	if err != nil || len(res.Entities) == 0 {
		return nil, fmt.Errorf("failed to find subnet by name %q: %v", subnetName, err)
	}

	if len(res.Entities) > 1 {
		return nil, fmt.Errorf("found more than one (%v) subnet with name %q", len(res.Entities), subnetName)
	}

	return res.Entities[0].Metadata.UUID, nil
}

// findClusterUUIDByName retrieves the cluster uuid by the given cluster name.
func findClusterUUIDByName(ctx context.Context, v3Client Client, clusterName string) (*string, error) {
	res, err := v3Client.ListCluster(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", clusterName)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find cluster by name %q: %v", clusterName, err)
	}
	entities := make([]*v3.ClusterIntentResponse, 0)
	for _, entity := range res.Entities {
		if entity.Status != nil && entity.Status.Resources != nil && entity.Status.Resources.Config != nil {
			serviceList := entity.Status.Resources.Config.ServiceList
			isPrismCentral := false
			for _, svc := range serviceList {
				// Prism Central is also internally a cluster, but we filter that out here as we only care about prism element clusters
				if svc != nil && strings.ToUpper(*svc) == "PRISM_CENTRAL" {
					isPrismCentral = true
				}
			}
			if !isPrismCentral && *entity.Spec.Name == clusterName {
				entities = append(entities, entity)
			}
		}
	}
	if len(entities) == 0 {
		return nil, fmt.Errorf("failed to find cluster by name %q: %v", clusterName, err)
	}

	if len(entities) > 1 {
		return nil, fmt.Errorf("found more than one (%v) cluster with name %q", len(entities), clusterName)
	}

	return entities[0].Metadata.UUID, nil
}

// findImageUUIDByName retrieves the image uuid by the given image name.
func findImageUUIDByName(ctx context.Context, v3Client Client, imageName string) (*string, error) {
	res, err := v3Client.ListImage(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", imageName)),
	})
	if err != nil || len(res.Entities) == 0 {
		return nil, fmt.Errorf("failed to find image by name %q: %v", imageName, err)
	}

	if len(res.Entities) > 1 {
		return nil, fmt.Errorf("found more than one (%v) image with name %q", len(res.Entities), imageName)
	}

	return res.Entities[0].Metadata.UUID, nil
}

func (v *Validator) validateUpgradeRolloutStrategy(clusterSpec *cluster.Spec) error {
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
		return fmt.Errorf("Upgrade rollout strategy customization is not supported for nutanix provider")
	}
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if workerNodeGroupConfiguration.UpgradeRolloutStrategy != nil {
			return fmt.Errorf("Upgrade rollout strategy customization is not supported for nutanix provider")
		}
	}
	return nil
}
