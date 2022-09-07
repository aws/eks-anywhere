package nutanix

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	"github.com/nutanix-cloud-native/prism-go-client/utils"
	"github.com/nutanix-cloud-native/prism-go-client/v3"
	"go.uber.org/multierr"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// Validator is a client to validate nutanix resources
type Validator struct {
	v3Client client
}

// NewValidator returns a new validator client
func NewValidator(config *anywherev1.NutanixDatacenterConfig, creds basicAuthCreds) (*Validator, error) {
	url := fmt.Sprintf("%s:%d", config.Spec.Endpoint, config.Spec.Port)
	nutanixCreds := prismgoclient.Credentials{
		URL:      url,
		Username: creds.username,
		Password: creds.password,
		Endpoint: config.Spec.Endpoint,
	}
	client, err := v3.NewV3Client(nutanixCreds)
	if err != nil {
		return nil, err
	}

	return &Validator{v3Client: client.V3}, nil
}

// ValidateMachineConfig validates the Prism Element cluster, subnet, and image for the machine
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
			clusterUUID, err := findClusterUUIDByName(ctx, v.v3Client, clusterName)
			if err != nil {
				return fmt.Errorf("failed to find cluster with name %q: %v", clusterName, err)
			}

			identifier.Type = anywherev1.NutanixIdentifierUUID
			identifier.UUID = clusterUUID
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing cluster uuid")
		} else {
			clusterUUID := *identifier.UUID
			if _, err := v.v3Client.GetCluster(ctx, clusterUUID); err != nil {
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
			imageUUID, err := findImageUUIDByName(ctx, v.v3Client, imageName)
			if err != nil {
				return fmt.Errorf("failed to find image with name %q: %v", imageName, err)
			}

			identifier.Type = anywherev1.NutanixIdentifierUUID
			identifier.UUID = imageUUID
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing image uuid")
		} else {
			imageUUID := *identifier.UUID
			if _, err := v.v3Client.GetImage(ctx, imageUUID); err != nil {
				return fmt.Errorf("failed to find image with uuid %s: %v", imageUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid cluster identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
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
			subnetUUID, err := findSubnetUUIDByName(ctx, v.v3Client, subnetName)
			if err != nil {
				return fmt.Errorf("failed to find subnet with name %s: %v", subnetName, err)
			} else {
				identifier.Type = anywherev1.NutanixIdentifierUUID
				identifier.UUID = subnetUUID
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing subnet uuid")
		} else {
			subnetUUID := *identifier.UUID
			if _, err := v.v3Client.GetSubnet(ctx, subnetUUID); err != nil {
				return fmt.Errorf("failed to find subnet with uuid %s: %v", subnetUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid cluster identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

// findSubnetUUIDByName retrieves the subnet uuid by the given subnet name
func findSubnetUUIDByName(ctx context.Context, v3Client client, subnetName string) (*string, error) {
	res, err := v3Client.ListSubnet(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", subnetName)),
	})
	if err != nil || len(res.Entities) == 0 {
		return nil, fmt.Errorf("failed to find subnet by name %q: %v", subnetName, err)
	}

	if len(res.Entities) > 1 {
		logr.FromContextOrDiscard(ctx).Info("Found more than one (%v) subnet with name %q.", len(res.Entities), subnetName)
	}

	return res.Entities[0].Metadata.UUID, nil
}

// findClusterUuidByName retrieves the cluster uuid by the given cluster name
func findClusterUUIDByName(ctx context.Context, v3Client client, clusterName string) (*string, error) {
	res, err := v3Client.ListCluster(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", clusterName)),
	})
	if err != nil || len(res.Entities) == 0 {
		return nil, fmt.Errorf("failed to find cluster by name %q: %v", clusterName, err)
	}

	if len(res.Entities) > 1 {
		logr.FromContextOrDiscard(ctx).Info("Found more than one (%v) cluster with name %q.", len(res.Entities), clusterName)
	}

	return res.Entities[0].Metadata.UUID, nil
}

// findImageByName retrieves the image uuid by the given image name
func findImageUUIDByName(ctx context.Context, v3Client client, imageName string) (*string, error) {
	res, err := v3Client.ListImage(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", imageName)),
	})
	if err != nil || len(res.Entities) == 0 {
		return nil, fmt.Errorf("failed to find image by name %q: %v", imageName, err)
	}

	if len(res.Entities) > 1 {
		logr.FromContextOrDiscard(ctx).Info("Found more than one (%v) image with name %q.", len(res.Entities), imageName)
	}

	return res.Entities[0].Metadata.UUID, nil
}
