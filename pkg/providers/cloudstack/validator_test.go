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
	setupContext()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk)

	cloudstackDatacenter, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	cmk.EXPECT().ValidateZonesPresent(ctx, cloudstackDatacenter.Spec.Zones).Return([]v1alpha1.CloudStackResourceIdentifier{{Name: "zone1", Id: "4e3b338d-87a6-4189-b931-a1747edeea8f"}}, nil)
	cmk.EXPECT().ValidateDomainPresent(ctx, cloudstackDatacenter.Spec.Domain).Return(v1alpha1.CloudStackResourceIdentifier{Id: "5300cdac-74d5-11ec-8696-c81f66d3e965", Name: cloudstackDatacenter.Spec.Domain}, nil)
	cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), gomock.Any()).Return(nil)
	cmk.EXPECT().ValidateNetworkPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false).Return(nil)
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
	cloudStackClusterSpec := &Spec{
		Spec:             clusterSpec,
		datacenterConfig: datacenterConfig,
	}
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = ""

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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)

	thenErrorExpected(t, "multiple worker node groups are not yet supported by the Cloudstack provider", err)
}

func TestValidateDatacenterConfigsNoNetwork(t *testing.T) {
	ctx := context.Background()
	setupContext()
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
	cmk.EXPECT().ValidateZonesPresent(ctx, gomock.Any()).Return([]v1alpha1.CloudStackResourceIdentifier{{Name: "zone1", Id: "4e3b338d-87a6-4189-b931-a1747edeea8f"}}, nil)
	cmk.EXPECT().ValidateDomainPresent(ctx, gomock.Any()).Return(v1alpha1.CloudStackResourceIdentifier{Id: "5300cdac-74d5-11ec-8696-c81f66d3e965", Name: "ROOT"}, nil)
	cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), gomock.Any()).Return(nil)

	datacenterConfig.Spec.Zones[0].Network.Id = ""
	datacenterConfig.Spec.Zones[0].Network.Name = ""
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

	datacenterConfig.Spec.ManagementApiEndpoint = ":1234.5234"
	err = validator.ValidateCloudStackDatacenterConfig(ctx, cloudStackClusterSpec.datacenterConfig)

	thenErrorExpected(t, "error while checking management api endpoint: :1234.5234 is not a valid url", err)
}

func TestValidateDatacenterInconsistentManagementEndpoints(t *testing.T) {
	ctx := context.Background()
	setupContext()
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

	datacenterConfig.Spec.ManagementApiEndpoint = "abcefg.com"
	err = validator.ValidateCloudStackDatacenterConfig(ctx, cloudStackClusterSpec.datacenterConfig)

	thenErrorExpected(t, "cloudstack secret management url (http://127.16.0.1:8080/client/api) differs from cluster spec management url (abcefg.com)", err)
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
	cmk.EXPECT().ValidateDomainPresent(gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateZonesPresent(gomock.Any(), gomock.Any()).Times(3).Return([]v1alpha1.CloudStackResourceIdentifier{{Name: "zone1", Id: "4e3b338d-87a6-4189-b931-a1747edeea8f"}}, nil)
	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
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

	cmk.EXPECT().ValidateDomainPresent(gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateZonesPresent(gomock.Any(), gomock.Any()).Times(3).Return([]v1alpha1.CloudStackResourceIdentifier{{Name: "zone1", Id: "4e3b338d-87a6-4189-b931-a1747edeea8f"}}, nil)
	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
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
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef = nil

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
	cmk.EXPECT().ValidateDomainPresent(gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateZonesPresent(gomock.Any(), gomock.Any()).Times(3).Return([]v1alpha1.CloudStackResourceIdentifier{{Name: "zone1", Id: "4e3b338d-87a6-4189-b931-a1747edeea8f"}}, nil)
	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(),
		gomock.Any(), datacenterConfig.Spec.Account, testTemplate).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), testOffering).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), datacenterConfig.Spec.Account, gomock.Any()).Times(3)
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

	for _, machineConfig := range machineConfigs {
		cmk.EXPECT().ValidateDomainPresent(gomock.Any(), gomock.Any())
		cmk.EXPECT().ValidateZonesPresent(gomock.Any(), gomock.Any()).Return([]v1alpha1.CloudStackResourceIdentifier{{Name: "zone1", Id: "4e3b338d-87a6-4189-b931-a1747edeea8f"}}, nil)
		cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(), "admin", machineConfig.Spec.Template).Return(nil)
		cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), machineConfig.Spec.ComputeOffering).Return(nil)
		if len(machineConfig.Spec.AffinityGroupIds) > 0 {
			cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), "admin", machineConfig.Spec.AffinityGroupIds).Return(nil)
		}
		err := validator.validateMachineConfig(ctx, datacenterConfig.Spec, machineConfig)
		if err != nil {
			t.Fatalf("failed to validate CloudStackMachineConfig: %v", err)
		}
	}
}
