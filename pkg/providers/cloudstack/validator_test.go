package cloudstack

import (
	"context"
	_ "embed"
	"errors"
	"net"
	"path"
	"testing"
	"time"

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

type DummyNetClient struct{}

func (n *DummyNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// add dummy case for coverage
	if address == "255.255.255.255:6443" {
		return &net.IPConn{}, nil
	}
	return nil, errors.New("")
}

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
	validator := NewValidator(cmk, &DummyNetClient{}, true)

	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

	err = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateCloudStackDatacenterConfigWithAZ(t *testing.T) {
	ctx := context.Background()
	setupContext(t)
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, &DummyNetClient{}, true)

	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainWithAZsFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

	err = validator.ValidateCloudStackDatacenterConfig(ctx, datacenterConfig)
	if err != nil {
		t.Fatalf("failed to validate CloudStackDataCenterConfig: %v", err)
	}
}

func TestValidateSkipControlPlaneIpCheck(t *testing.T) {
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	if err := validator.ValidateControlPlaneEndpointUniqueness("invalid_url_skip_check"); err != nil {
		t.Fatalf("expected no error, validation should be skipped")
	}
}

func TestValidateControlPlaneIpCheck(t *testing.T) {
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, &DummyNetClient{}, false)
	err := validator.ValidateControlPlaneEndpointUniqueness("255.255.255.255:6443")
	thenErrorExpected(t, "endpoint <255.255.255.255:6443> is already in use", err)
}

func TestValidateControlPlaneIpCheckInvalidPort(t *testing.T) {
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, &DummyNetClient{}, false)
	err := validator.ValidateControlPlaneEndpointUniqueness("255.255.255.255")
	thenErrorExpected(t, "invalid endpoint - not in host:port format: address 255.255.255.255: missing port in address", err)
}

func TestValidateControlPlaneIpCheckUniqueIpSuccess(t *testing.T) {
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, &DummyNetClient{}, false)
	if err := validator.ValidateControlPlaneEndpointUniqueness("1.1.1.1:6443"); err != nil {
		t.Fatalf("Expected endpoint to be valid and unused")
	}
}

func TestValidateMachineConfigsInvalidControlPlanePort(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, false)
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}
	cloudStackClusterSpec := &Spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigs,
	}
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "255.255.255.255:0"

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)

	thenErrorExpected(t, "validating controlPlaneConfiguration.Endpoint.Host: host 255.255.255.255:0 has an invalid port", err)
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
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	datacenterConfig.Spec.AvailabilityZones[0].Zone.Network.Id = ""
	datacenterConfig.Spec.AvailabilityZones[0].Zone.Network.Name = ""
	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

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
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

	datacenterConfig.Spec.AvailabilityZones[0].ManagementApiEndpoint = ":1234.5234"
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
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

	datacenterConfig.Spec.AvailabilityZones[0].ManagementApiEndpoint = "abcefg.com"
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
	validator := NewValidator(cmk, &DummyNetClient{}, true)
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

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

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

func TestSetupAndValidateSshAuthorizedKeysNil(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
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

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

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

//func setupMockForAvailabilityZonesValidation(cmk *mocks.MockProviderCmkClient, ctx context.Context, datacenterConfig *v1alpha1.CloudStackDatacenterConfig) {
//	if len(datacenterConfig.Spec.Zones) > 0 {
//		cmk.EXPECT().ValidateZoneAndGetId(ctx, gomock.Any(), datacenterConfig.Spec.Zones[0]).AnyTimes().Return("4e3b338d-87a6-4189-b931-a1747edeea8f", nil)
//	}
//	cmk.EXPECT().ValidateDomainAndGetId(ctx, gomock.Any(), datacenterConfig.Spec.Domain).AnyTimes().Return("5300cdac-74d5-11ec-8696-c81f66d3e965", nil)
//	cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
//	cmk.EXPECT().ValidateNetworkPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
//	cmk.EXPECT().GetManagementApiEndpoint(gomock.Any()).AnyTimes().MaxTimes(1).Return("http://127.16.0.1:8080/client/api", nil)
//}

func setupMockForAvailabilityZonesValidation(cmk *mocks.MockProviderCmkClient, ctx context.Context, azs []v1alpha1.CloudStackAvailabilityZone) {
	for _, az := range azs {
		cmk.EXPECT().ValidateZoneAndGetId(ctx, gomock.Any(), az.Zone).AnyTimes().Return("4e3b338d-87a6-4189-b931-a1747edeea82", nil)
		cmk.EXPECT().ValidateDomainAndGetId(ctx, gomock.Any(), az.Domain).AnyTimes().Return("5300cdac-74d5-11ec-8696-c81f66d3e962", nil)
		cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), az.Account, gomock.Any()).AnyTimes().Return(nil)
		cmk.EXPECT().ValidateNetworkPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		cmk.EXPECT().GetManagementApiEndpoint(az.CredentialsRef).AnyTimes().Return(az.ManagementApiEndpoint, nil)
	}
}

func TestSetupAndValidateCreateClusterCPMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file %s", testClusterConfigMainFilename)
	}
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
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

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

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
	validator := NewValidator(cmk, &DummyNetClient{}, true)
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

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

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
	validator := NewValidator(cmk, &DummyNetClient{}, true)
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

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

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
	validator := NewValidator(cmk, &DummyNetClient{}, true)
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

	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

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
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	setupMockForAvailabilityZonesValidation(cmk, ctx, datacenterConfig.Spec.AvailabilityZones)

	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, testTemplate).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, gomock.Any()).Times(3)

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
	validator := NewValidator(cmk, &DummyNetClient{}, true)

	cmk.EXPECT().ValidateZoneAndGetId(ctx, gomock.Any(), gomock.Any()).Times(3).Return("4e3b338d-87a6-4189-b931-a1747edeea82", nil)
	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, testTemplate).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, gomock.Any()).Times(3)

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

	validator := NewValidator(cmk, &DummyNetClient{}, true)
	cmk.EXPECT().ValidateZoneAndGetId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("4e3b338d-87a6-4189-b931-a1747edeea8f", nil)
	cmk.EXPECT().ValidateDomainAndGetId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	cmk.EXPECT().ValidateNetworkPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	cmk.EXPECT().GetManagementApiEndpoint(gomock.Any()).AnyTimes().Return("http://127.16.0.1:8080/client/api", nil)

	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, testTemplate).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).AnyTimes()
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, gomock.Any()).AnyTimes()

	// Valid affinity types
	err = validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec)
	assert.Nil(t, err)
}
