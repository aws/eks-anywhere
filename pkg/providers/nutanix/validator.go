package nutanix

import (
	"context"
	"fmt"
	"strings"

	"github.com/nutanix-cloud-native/prism-go-client/utils"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
	"go.uber.org/multierr"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/logger"
)

// Validator is a client to validate nutanix resources.
type Validator struct {
	client        Client
	certValidator crypto.TlsValidator
}

// NewValidator returns a new validator client.
func NewValidator(client Client, certValidator crypto.TlsValidator) *Validator {
	return &Validator{
		client:        client,
		certValidator: certValidator,
	}
}

// ValidateDatacenterConfig validates the datacenter config.
func (v *Validator) ValidateDatacenterConfig(ctx context.Context, config *anywherev1.NutanixDatacenterConfig) error {
	if config.Spec.Insecure {
		logger.Info("Warning: Skipping TLS validation for insecure connection to Nutanix Prism Central; this is not recommended for production use")
	}
	return v.validateTrustBundleConfig(config.Spec)
}

func (v *Validator) validateTrustBundleConfig(dcConf anywherev1.NutanixDatacenterConfigSpec) error {
	if dcConf.AdditionalTrustBundle == "" {
		return nil
	}
	return v.certValidator.ValidateCert(dcConf.Endpoint, fmt.Sprintf("%d", dcConf.Port), dcConf.AdditionalTrustBundle)
}

// ValidateMachineConfig validates the Prism Element cluster, subnet, and image for the machine.
func (v *Validator) ValidateMachineConfig(ctx context.Context, config *anywherev1.NutanixMachineConfig) error {
	var errors error
	if err := v.validateClusterConfig(ctx, config.Spec.Cluster); err != nil {
		errors = multierr.Append(errors, err)
	}

	if err := v.validateSubnetConfig(ctx, config.Spec.Subnet); err != nil {
		errors = multierr.Append(errors, err)
	}

	if err := v.validateImageConfig(ctx, config.Spec.Image); err != nil {
		errors = multierr.Append(errors, err)
	}

	return errors
}

func (v *Validator) validateClusterConfig(ctx context.Context, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing cluster name")
		} else {
			clusterName := *identifier.Name
			if _, err := findClusterUUIDByName(ctx, v.client, clusterName); err != nil {
				return fmt.Errorf("failed to find cluster with name %q: %v", clusterName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing cluster uuid")
		} else {
			clusterUUID := *identifier.UUID
			if _, err := v.client.GetCluster(ctx, clusterUUID); err != nil {
				return fmt.Errorf("failed to find cluster with uuid %v: %v", clusterUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid cluster identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

func (v *Validator) validateImageConfig(ctx context.Context, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing image name")
		} else {
			imageName := *identifier.Name
			if _, err := findImageUUIDByName(ctx, v.client, imageName); err != nil {
				return fmt.Errorf("failed to find image with name %q: %v", imageName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing image uuid")
		} else {
			imageUUID := *identifier.UUID
			if _, err := v.client.GetImage(ctx, imageUUID); err != nil {
				return fmt.Errorf("failed to find image with uuid %s: %v", imageUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid image identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

func (v *Validator) validateSubnetConfig(ctx context.Context, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing subnet name")
		} else {
			subnetName := *identifier.Name
			if _, err := findSubnetUUIDByName(ctx, v.client, subnetName); err != nil {
				return fmt.Errorf("failed to find subnet with name %s: %v", subnetName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing subnet uuid")
		} else {
			subnetUUID := *identifier.UUID
			if _, err := v.client.GetSubnet(ctx, subnetUUID); err != nil {
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
