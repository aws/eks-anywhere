package nutanix

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nutanix-cloud-native/prism-go-client/utils"
	"github.com/nutanix-cloud-native/prism-go-client/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

//go:embed testdata/machineConfigUUID.yaml
var nutanixMachineConfigSpecUUID string

//go:embed testdata/machineConfigInvalidIdentifier.yaml
var nutanixMachineConfigSpecInvalidIdentifier string

//go:embed testdata/machineConfigEmptyName.yaml
var nutanixMachineConfigSpecEmptyName string

//go:embed testdata/machineConfigEmptyUUID.yaml
var nutanixMachineConfigSpecEmptyUUID string

func TestNutanixValidatorNoResourcesFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockClient := NewMockClient(ctrl)
	mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no clusters found"))
	mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no subnets found"))
	mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no images found"))
	validator := NewValidator(mockClient)
	require.NotNil(t, validator)

	machineConfig := &anywherev1.NutanixMachineConfig{}
	err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConfig)
	require.NoError(t, err)

	err = validator.ValidateMachineConfig(context.Background(), machineConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no clusters found")
	assert.Contains(t, err.Error(), "no subnets found")
	assert.Contains(t, err.Error(), "no images found")
}

func TestNutanixValidatorDuplicateResourcesFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockClient := NewMockClient(ctrl)
	clusters := &v3.ClusterListIntentResponse{
		Entities: []*v3.ClusterIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cda"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster"),
				},
				Status: &v3.ClusterDefStatus{
					Resources: &v3.ClusterObj{
						Config: &v3.ClusterConfig{
							ServiceList: []*string{utils.StringPtr("AOS")},
						},
					},
				},
			},
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster"),
				},
				Status: &v3.ClusterDefStatus{
					Resources: &v3.ClusterObj{
						Config: &v3.ClusterConfig{
							ServiceList: []*string{utils.StringPtr("AOS")},
						},
					},
				},
			},
		},
	}
	mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(clusters, nil)
	subnets := &v3.SubnetListIntentResponse{
		Entities: []*v3.SubnetIntentResponse{
			{
				Spec: &v3.Subnet{
					Name: utils.StringPtr("prism-subnet"),
				},
			},
			{
				Spec: &v3.Subnet{
					Name: utils.StringPtr("prism-subnet"),
				},
			},
		},
	}
	mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(subnets, nil)
	images := &v3.ImageListIntentResponse{
		Entities: []*v3.ImageIntentResponse{
			{
				Spec: &v3.Image{
					Name: utils.StringPtr("prism-image"),
				},
			},
			{
				Spec: &v3.Image{
					Name: utils.StringPtr("prism-image"),
				},
			},
		},
	}
	mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(images, nil)
	validator := NewValidator(mockClient)
	require.NotNil(t, validator)

	machineConfig := &anywherev1.NutanixMachineConfig{}
	err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConfig)
	require.NoError(t, err)

	err = validator.ValidateMachineConfig(context.Background(), machineConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "found more than one (2) cluster")
	assert.Contains(t, err.Error(), "found more than one (2) subnet")
	assert.Contains(t, err.Error(), "found more than one (2) image")
}

func TestNutanixValidatorIdentifierTypeUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockClient := NewMockClient(ctrl)
	mockClient.EXPECT().GetCluster(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no clusters found"))
	mockClient.EXPECT().GetSubnet(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no subnets found"))
	mockClient.EXPECT().GetImage(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no images found"))
	validator := NewValidator(mockClient)
	require.NotNil(t, validator)

	machineConfig := &anywherev1.NutanixMachineConfig{}
	err := yaml.Unmarshal([]byte(nutanixMachineConfigSpecUUID), machineConfig)
	require.NoError(t, err)

	err = validator.ValidateMachineConfig(context.Background(), machineConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no clusters found")
	assert.Contains(t, err.Error(), "no subnets found")
	assert.Contains(t, err.Error(), "no images found")
}

func TestNutanixValidatorInvalidIdentifierType(t *testing.T) {
	ctrl := gomock.NewController(t)

	machineConfig := &anywherev1.NutanixMachineConfig{}
	err := yaml.Unmarshal([]byte(nutanixMachineConfigSpecInvalidIdentifier), machineConfig)
	require.NoError(t, err)

	mockClient := NewMockClient(ctrl)
	validator := NewValidator(mockClient)
	require.NotNil(t, validator)

	err = validator.ValidateMachineConfig(context.Background(), machineConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cluster identifier type")
	assert.Contains(t, err.Error(), "invalid subnet identifier type")
	assert.Contains(t, err.Error(), "invalid image identifier type")
}

func TestNutanixValidatorEmptyResourceName(t *testing.T) {
	ctrl := gomock.NewController(t)

	machineConfig := &anywherev1.NutanixMachineConfig{}
	err := yaml.Unmarshal([]byte(nutanixMachineConfigSpecEmptyName), machineConfig)
	require.NoError(t, err)

	mockClient := NewMockClient(ctrl)
	validator := NewValidator(mockClient)
	require.NotNil(t, validator)

	err = validator.ValidateMachineConfig(context.Background(), machineConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing cluster name")
	assert.Contains(t, err.Error(), "missing subnet name")
	assert.Contains(t, err.Error(), "missing image name")
}

func TestNutanixValidatorEmptyResourceUUID(t *testing.T) {
	ctrl := gomock.NewController(t)

	machineConfig := &anywherev1.NutanixMachineConfig{}
	err := yaml.Unmarshal([]byte(nutanixMachineConfigSpecEmptyUUID), machineConfig)
	require.NoError(t, err)

	mockClient := NewMockClient(ctrl)
	validator := NewValidator(mockClient)
	require.NotNil(t, validator)

	err = validator.ValidateMachineConfig(context.Background(), machineConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing cluster uuid")
	assert.Contains(t, err.Error(), "missing subnet uuid")
	assert.Contains(t, err.Error(), "missing image uuid")
}
