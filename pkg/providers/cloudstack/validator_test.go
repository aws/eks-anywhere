package cloudstack

import (
	"context"
	_ "embed"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
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
	validator := NewValidator(cmk)

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
	validator := NewValidator(cmk)

	cmk.EXPECT().ValidateCloudStackConnection(ctx).Return(nil)
	err := validator.validateCloudStackAccess(ctx)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateMachineConfigsNoControlPlaneEndpointIP(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk)
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

func TestValidateMachineConfigsMultipleWorkerNodeGroupsUnsupported(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk)
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = append(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0])
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)

	thenErrorExpected(t, "multiple worker node groups are not yet supported by the Cloudstack provider", err)
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
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	validator := NewValidator(cmk)
	cloudStackClusterSpec.datacenterConfig.Spec.Network = v1alpha1.CloudStackResourceRef{
		Value: "",
		Type:  "id",
	}

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)

	thenErrorExpected(t, "CloudStackDatacenterConfig network is not set or is empty", err)
}

func TestSetupAndValidateUsersNil(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.Users = nil
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.Users = nil
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.Users = nil
	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateSshAuthorizedKeysNil(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateCreateClusterCPMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef = nil

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	thenErrorExpected(t, "must specify machineGroupRef for control plane", err)
}

func TestSetupAndValidateCreateClusterWorkerMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef = nil

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	thenErrorExpected(t, "must specify machineGroupRef for worker nodes", err)
}

func TestSetupAndValidateCreateClusterEtcdMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef = nil

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	thenErrorExpected(t, "must specify machineGroupRef for etcd machines", err)
}

func TestSetupAndValidateCreateClusterCPMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "nonexistent"

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for control plane", err)
}

func TestSetupAndValidateCreateClusterWorkerMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "nonexistent"

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for worker nodes", err)
}

func TestSetupAndValidateCreateClusterEtcdMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "nonexistent"

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for etcd machines", err)
}

func TestSetupAndValidateCreateClusterTemplateDifferent(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.Template = v1alpha1.CloudStackResourceRef{Value: "different", Type: "name"}

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	thenErrorExpected(t, "control plane and etcd machines must have the same template specified", err)
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
	validator := NewValidator(cmk)
	cmk.EXPECT().ValidateTemplatePresent(ctx, datacenterConfig.Spec.Domain,
		datacenterConfig.Spec.Zone, datacenterConfig.Spec.Account, testTemplate).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, datacenterConfig.Spec.Domain,
		datacenterConfig.Spec.Zone, datacenterConfig.Spec.Account, testOffering).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, datacenterConfig.Spec.Domain,
		datacenterConfig.Spec.Zone, datacenterConfig.Spec.Account, gomock.Any()).Times(3)
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
	validator := NewValidator(cmk)

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
