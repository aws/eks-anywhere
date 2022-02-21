package cloudstack

import (
	"context"
	_ "embed"
	"path"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/mocks"
)

const (
	testClusterConfigMainFilename = "cluster_main.yaml"
	testDataDir                   = "testdata"
)

var testZone = v1alpha1.CloudStackResourceRef{
	Type:  "name",
	Value: "zone1",
}

func givenCloudMonkeyMock(t *testing.T) *mocks.MockProviderCmkClient {
	ctrl := gomock.NewController(t)
	return mocks.NewMockProviderCmkClient(ctrl)
}

func givenDatacenterConfig(t *testing.T, fileName string) *v1alpha1.CloudStackDatacenterConfig {
	deploymentConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file: %v", err)
	}
	return deploymentConfig
}

func givenMachineConfigs(t *testing.T, fileName string) map[string]*v1alpha1.CloudStackMachineConfig {
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get machine configs from file")
	}
	return machineConfigs
}

func TestValidateCloudStackDatacenterConfig(t *testing.T) {
	ctx := context.Background()
	cmk := givenCloudMonkeyMock(t)
	validator := NewValidator(cmk, nil)

	cloudstackDatacenter := givenDatacenterConfig(t, testClusterConfigMainFilename)

	cmk.EXPECT().ValidateZonePresent(ctx, cloudstackDatacenter.Spec.Zone).Return(nil)
	err := validator.ValidateCloudStackDatacenterConfig(ctx, cloudstackDatacenter)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateCloudStackConnection(t *testing.T) {
	ctx := context.Background()
	cmk := givenCloudMonkeyMock(t)
	validator := NewValidator(cmk, nil)

	cmk.EXPECT().ValidateCloudStackConnection(ctx).Return(nil)
	err := validator.validateCloudStackAccess(ctx)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateCloudStackMachineConfig(t *testing.T) {
	ctx := context.Background()
	cmk := givenCloudMonkeyMock(t)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	validator := NewValidator(cmk, machineConfigs)

	for _, machineConfig := range machineConfigs {
		cmk.EXPECT().ValidateTemplatePresent(ctx, "domain1", testZone, "admin", machineConfig.Spec.Template).Return(nil)
		cmk.EXPECT().ValidateServiceOfferingPresent(ctx, "domain1", testZone, "admin", machineConfig.Spec.ComputeOffering).Return(nil)
		if len(machineConfig.Spec.AffinityGroupIds) > 0 {
			cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, "domain1", testZone, "admin", machineConfig.Spec.AffinityGroupIds).Return(nil)
		}
		err := validator.validateMachineConfig(ctx, datacenterConfig.Spec, machineConfig)
		if err != nil {
			t.Fatalf("failed to validate CloudStackMachineConfig: %v", err)
		}
	}
}
