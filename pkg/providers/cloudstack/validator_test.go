package cloudstack

import (
	"context"
	_ "embed"
	"errors"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/mocks"
)

const (
	testClusterConfigMainFilename        = "cluster_main.yaml"
	testClusterConfigMainWithAZsFilename = "cluster_main_with_availability_zones.yaml"
	testDataDir                          = "testdata"
)

var testTemplate = v1alpha1.CloudStackResourceIdentifier{
	Name: "centos7-k8s-118",
}

var testOffering = v1alpha1.CloudStackResourceIdentifier{
	Name: "m4-large",
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
	setupContext(t)
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk)

	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	err = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateCloudStackDatacenterConfigWithAZ(t *testing.T) {
	ctx := context.Background()
	setupContext(t)
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk)

	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainWithAZsFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)
	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

	for _, az := range datacenterConfig.Spec.AvailabilityZones {
		cmk.EXPECT().GetManagementApiEndpoint(gomock.Any()).Times(1).Return(az.ManagementApiEndpoint, nil)
	}

	err = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateCloudStackConnection(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	cmk.EXPECT().ValidateCloudStackConnection(ctx, "global").Return(nil)
	if err := validator.validateCloudStackAccess(ctx, datacenterConfig); err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateCloudStackConnectionFailure(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	cmk.EXPECT().ValidateCloudStackConnection(ctx, "global").Return(errors.New("exception"))
	err = validator.validateCloudStackAccess(ctx, datacenterConfig)
	thenErrorExpected(t, "validating connection to cloudstack global: exception", err)
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
	cloudStackClusterSpec := &Spec{
		Spec:             clusterSpec,
		datacenterConfig: datacenterConfig,
	}
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = ""

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)

	thenErrorExpected(t, "cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty", err)
}

func TestValidateDatacenterConfigsNoNetwork(t *testing.T) {
	ctx := context.Background()
	setupContext(t)
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: nil,
	}
	validator := NewValidator(cmk)
	datacenterConfig.Spec.Zones[0].Network.Id = ""
	datacenterConfig.Spec.Zones[0].Network.Name = ""
	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	err = validator.ValidateCloudStackDatacenterConfig(ctx, cloudStackClusterSpec.datacenterConfig)

	thenErrorExpected(t, "zone network is not set or is empty", err)
}

func TestValidateDatacenterBadManagementEndpoint(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: nil,
	}
	validator := NewValidator(cmk)
	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	datacenterConfig.Spec.ManagementApiEndpoint = ":1234.5234"
	err = validator.ValidateCloudStackDatacenterConfig(ctx, cloudStackClusterSpec.datacenterConfig)

	thenErrorExpected(t, "checking management api endpoint: :1234.5234 is not a valid url", err)
}

func TestValidateDatacenterInconsistentManagementEndpoints(t *testing.T) {
	ctx := context.Background()
	setupContext(t)
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: nil,
	}
	validator := NewValidator(cmk)
	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	datacenterConfig.Spec.ManagementApiEndpoint = "abcefg.com"
	err = validator.ValidateCloudStackDatacenterConfig(ctx, cloudStackClusterSpec.datacenterConfig)

	thenErrorExpected(t, "cloudstack secret management url (http://127.16.0.1:8080/client/api) differs from cluster spec management url (abcefg.com)", err)
}

func TestSetupAndValidateDiskOfferingEmpty(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	if err != nil {
		t.Fatalf("validator.ValidateClusterMachineConfigs() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateValidDiskOffering(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	if err != nil {
		t.Fatalf("validator.ValidateClusterMachineConfigs() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateInvalidDiskOfferingNotPresent(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("match me"))
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	wantErrMsg := "validating disk offering: match me"
	assert.Contains(t, err.Error(), wantErrMsg, "expected error containing %q, got %v", wantErrMsg, err)
}

func TestSetupAndValidateInValidDiskOfferingBadMountPath(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	wantErrMsg := "machine config test validation failed: mountPath: / invalid, must be non-empty and starts with /"
	assert.Contains(t, err.Error(), wantErrMsg, "expected error containing %q, got %v", wantErrMsg, err)
}

func TestSetupAndValidateInValidDiskOfferingEmptyDevice(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	wantErrMsg := "machine config test validation failed: device:  invalid, empty device"
	assert.Contains(t, err.Error(), wantErrMsg, "expected error containing %q, got %v", wantErrMsg, err)
}

func TestSetupAndValidateInValidDiskOfferingEmptyFilesystem(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "",
		Label:      "data_disk",
	}
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	wantErrMsg := "machine config test validation failed: filesystem:  invalid, empty filesystem"
	assert.Contains(t, err.Error(), wantErrMsg, "expected error containing %q, got %v", wantErrMsg, err)
}

func TestSetupAndValidateInValidDiskOfferingEmptyLabel(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "",
	}
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.DiskOffering = v1alpha1.CloudStackResourceDiskOffering{}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	wantErrMsg := "machine config test validation failed: label:  invalid, empty label"
	assert.Contains(t, err.Error(), wantErrMsg, "expected error containing %q, got %v", wantErrMsg, err)
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.Users = nil
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.Users = nil
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.Users = nil

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	if err != nil {
		t.Fatalf("validator.ValidateClusterMachineConfigs() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateRestrictedUserDetails(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.UserCustomDetails = map[string]string{"keyboard": "test"}
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.UserCustomDetails = map[string]string{"keyboard": "test"}
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.UserCustomDetails = map[string]string{"keyboard": "test"}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	if err == nil {
		t.Fatalf("expected error like 'validation failed: restricted key keyboard found in custom user details' but no error was thrown")
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	if err != nil {
		t.Fatalf("validator.ValidateClusterMachineConfigs() err = %v, want err = nil", err)
	}
}

func setupMockForDatacenterConfigValidation(cmk *mocks.MockProviderCmkClient, ctx context.Context, datacenterConfig *v1alpha1.CloudStackDatacenterConfig) {
	if len(datacenterConfig.Spec.Zones) > 0 {
		cmk.EXPECT().ValidateZoneAndGetId(ctx, gomock.Any(), datacenterConfig.Spec.Zones[0]).AnyTimes().Return("4e3b338d-87a6-4189-b931-a1747edeea8f", nil)
	}
	cmk.EXPECT().ValidateDomainAndGetId(ctx, gomock.Any(), datacenterConfig.Spec.Domain).AnyTimes().Return("5300cdac-74d5-11ec-8696-c81f66d3e965", nil)
	cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	cmk.EXPECT().ValidateNetworkPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	cmk.EXPECT().GetManagementApiEndpoint(gomock.Any()).AnyTimes().MaxTimes(1).Return("http://127.16.0.1:8080/client/api", nil)
}

func setupMockForAvailabilityZonesValidation(cmk *mocks.MockProviderCmkClient, ctx context.Context, azs []v1alpha1.CloudStackAvailabilityZone) {
	for _, az := range azs {
		cmk.EXPECT().ValidateZoneAndGetId(ctx, gomock.Any(), az.Zone).AnyTimes().Return("4e3b338d-87a6-4189-b931-a1747edeea82", nil)
		cmk.EXPECT().ValidateDomainAndGetId(ctx, gomock.Any(), az.Domain).AnyTimes().Return("5300cdac-74d5-11ec-8696-c81f66d3e962", nil)
		cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), az.Account, gomock.Any()).AnyTimes().Return(nil)
		cmk.EXPECT().ValidateNetworkPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef = nil

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef = nil

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef = nil

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "nonexistent"

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "nonexistent"

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "nonexistent"

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cloudStackClusterSpec.machineConfigsLookup[controlPlaneMachineConfigName].Spec.Template = v1alpha1.CloudStackResourceIdentifier{Name: "different"}

	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	validator := NewValidator(cmk)
	setupMockForDatacenterConfigValidation(cmk, ctx, datacenterConfig)

	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), datacenterConfig.Spec.Account, testTemplate).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), datacenterConfig.Spec.Account, gomock.Any()).Times(3)

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	assert.Nil(t, err)
	assert.Equal(t, "1.2.3.4:6443", clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
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

	cmk.EXPECT().ValidateZoneAndGetId(ctx, gomock.Any(), gomock.Any()).Times(3).Return("4e3b338d-87a6-4189-b931-a1747edeea82", nil)
	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), datacenterConfig.Spec.Account, testTemplate).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), datacenterConfig.Spec.Account, gomock.Any()).Times(3)

	for _, machineConfig := range machineConfigs {
		err := validator.validateMachineConfig(ctx, datacenterConfig, machineConfig)
		if err != nil {
			t.Fatalf("failed to validate CloudStackMachineConfig: %v", err)
		}
	}
}

func TestValidateMachineConfigsWithAffinity(t *testing.T) {
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	cloudStackClusterSpec.controlPlaneMachineConfig().Spec.Affinity = "pro"
	cloudStackClusterSpec.controlPlaneMachineConfig().Spec.AffinityGroupIds = []string{}
	cloudStackClusterSpec.etcdMachineConfig().Spec.Affinity = "anti"
	cloudStackClusterSpec.etcdMachineConfig().Spec.AffinityGroupIds = []string{}
	for _, machineConfig := range machineConfigs {
		machineConfig.Spec.Affinity = "no"
		machineConfig.Spec.AffinityGroupIds = []string{}
	}

	validator := NewValidator(cmk)
	cmk.EXPECT().ValidateZoneAndGetId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("4e3b338d-87a6-4189-b931-a1747edeea8f", nil)
	cmk.EXPECT().ValidateDomainAndGetId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	cmk.EXPECT().ValidateNetworkPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	cmk.EXPECT().GetManagementApiEndpoint(gomock.Any()).AnyTimes().Return("http://127.16.0.1:8080/client/api", nil)

	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), datacenterConfig.Spec.Account, testTemplate).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).AnyTimes()
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), datacenterConfig.Spec.Account, gomock.Any()).AnyTimes()

	// Valid affinity types
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	assert.Nil(t, err)

	// Bad affinity type
	originalValue := cloudStackClusterSpec.controlPlaneMachineConfig().Spec.Affinity
	cloudStackClusterSpec.controlPlaneMachineConfig().Spec.Affinity = "xxx"
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	assert.NotNil(t, err)
	cloudStackClusterSpec.controlPlaneMachineConfig().Spec.Affinity = originalValue

	// Both affinity and affinityGroupIds are defined
	cloudStackClusterSpec.controlPlaneMachineConfig().Spec.AffinityGroupIds = []string{"affinity-group-1"}
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	assert.NotNil(t, err)
	cloudStackClusterSpec.controlPlaneMachineConfig().Spec.Affinity = originalValue
}
