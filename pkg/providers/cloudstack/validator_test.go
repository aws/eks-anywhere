package cloudstack

import (
	"context"
	_ "embed"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
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

var testTemplate = v1alpha1.CloudStackResourceRef{
	Type:  "name",
	Value: "centos7-k8s-118",
}

var testOffering = v1alpha1.CloudStackResourceRef{
	Type:  "name",
	Value: "m4-large",
}

func thenErrorExpected(t *testing.T, expected string, err error) {
	if err == nil {
		t.Fatalf("Expected=<%s> actual=<nil>", expected)
	}
	actual := err.Error()
	if expected != actual {
		t.Fatalf("Expected=<%s> actual=<%s>", expected, actual)
	}
}

func TestValidateCloudStackDatacenterConfig(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, nil)

	cloudstackDatacenter, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	cmk.EXPECT().ValidateZonePresent(ctx, cloudstackDatacenter.Spec.Zone).Return(nil)
	err = validator.ValidateCloudStackDatacenterConfig(ctx, cloudstackDatacenter)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateCloudStackConnection(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, nil)

	cmk.EXPECT().ValidateCloudStackConnection(ctx).Return(nil)
	err := validator.validateCloudStackAccess(ctx)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateMachineConfigsNoControlPlaneEndpointIP(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, nil)
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:             clusterSpec,
		datacenterConfig: datacenterConfig,
	}
	clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host = ""

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)

	thenErrorExpected(t, "cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty", err)
}

func TestValidateMachineConfigsNoNetwork(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	cloudStackClusterSpec := &spec{
		Spec:             clusterSpec,
		datacenterConfig: datacenterConfig,
	}
	validator := NewValidator(cmk, machineConfigs)
	cloudStackClusterSpec.datacenterConfig.Spec.Network = v1alpha1.CloudStackResourceRef{
		Value: "",
		Type:  "id",
	}

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)

	thenErrorExpected(t, "CloudStackDatacenterConfig network is not set or is empty", err)
}

func TestValidateMachineConfigsHappyCase(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	validator := NewValidator(cmk, machineConfigs)
	cmk.EXPECT().ValidateTemplatePresent(ctx, datacenterConfig.Spec.Domain,
		datacenterConfig.Spec.Zone, datacenterConfig.Spec.Account, testTemplate).Times(2)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, datacenterConfig.Spec.Domain,
		datacenterConfig.Spec.Zone, datacenterConfig.Spec.Account, testOffering).Times(2)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, datacenterConfig.Spec.Domain,
		datacenterConfig.Spec.Zone, datacenterConfig.Spec.Account, gomock.Any()).Times(2)
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	assert.Nil(t, err)
	assert.Equal(t, "1.2.3.4:6443", clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host)
}

func TestValidateCloudStackMachineConfig(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
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
