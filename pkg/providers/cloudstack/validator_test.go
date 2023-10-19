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
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
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
	Name: "kubernetes_1_21",
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

func TestValidateControlPlaneIpCheckUniqueIpSuccess(t *testing.T) {
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, &DummyNetClient{}, false)
	if err := validator.ValidateControlPlaneEndpointUniqueness("1.1.1.1:6443"); err != nil {
		t.Fatalf("Expected endpoint to be valid and unused")
	}
}

func TestValidateControlPlaneIpCheckUniqueIpInvalidEndpointPort(t *testing.T) {
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, &DummyNetClient{}, false)
	err := validator.ValidateControlPlaneEndpointUniqueness("1.1.1.1:")
	thenErrorExpected(t, "invalid endpoint: host 1.1.1.1: has an invalid port", err)
}

func TestValidateControlPlaneIpCheckUniqueIpInvalidEndpointHost(t *testing.T) {
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	validator := NewValidator(cmk, &DummyNetClient{}, false)
	err := validator.ValidateControlPlaneEndpointUniqueness("invalid::host")
	thenErrorExpected(t, "invalid endpoint: host invalid::host is invalid: address invalid::host: too many colons in address", err)
}

func TestValidateDatacenterInconsistentManagementEndpoints(t *testing.T) {
	ctx := context.Background()
	setupContext(t)
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	setupMockForAvailabilityZonesValidation(cmk, ctx, clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones)

	clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones[0].ManagementApiEndpoint = "abcefg.com"
	err := validator.ValidateCloudStackDatacenterConfig(ctx, clusterSpec.CloudStackDatacenter)

	thenErrorExpected(t, "cloudstack secret management url (http://127.16.0.1:8080/client/api) differs from cluster spec management url (abcefg.com)", err)
}

func TestSetupAndValidateUsersNil(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users = nil
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[workerNodeMachineConfigName].Spec.Users = nil
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[etcdMachineConfigName].Spec.Users = nil

	setupMockForAvailabilityZonesValidation(cmk, ctx, clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, clusterSpec.CloudStackDatacenter)
	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("validator.ValidateClusterMachineConfigs() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateSshAuthorizedKeysNil(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil

	setupMockForAvailabilityZonesValidation(cmk, ctx, clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones)

	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, clusterSpec.CloudStackDatacenter)
	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
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
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "nonexistent"

	setupMockForAvailabilityZonesValidation(cmk, ctx, clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones)

	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for control plane", err)
}

func TestSetupAndValidateCreateClusterWorkerMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "nonexistent"

	setupMockForAvailabilityZonesValidation(cmk, ctx, clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones)

	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for worker nodes", err)
}

func TestSetupAndValidateCreateClusterEtcdMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "nonexistent"

	setupMockForAvailabilityZonesValidation(cmk, ctx, clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones)

	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for etcd machines", err)
}

func TestValidateMachineConfigsHappyCase(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	validator := NewValidator(cmk, &DummyNetClient{}, true)
	setupMockForAvailabilityZonesValidation(cmk, ctx, clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones)

	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(),
		gomock.Any(), clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones[0].Account, testTemplate).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones[0].Account, gomock.Any()).Times(3)

	_ = validator.ValidateCloudStackDatacenterConfig(ctx, clusterSpec.CloudStackDatacenter)
	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	assert.Nil(t, err)
}

func TestValidateCloudStackMachineConfig(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))

	config, err := cluster.ParseConfigFromFile(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get machine configs from file: %v", err)
	}
	machineConfigs := config.CloudStackMachineConfigs
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

func TestValidateClusterMachineConfigsError(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	clusterSpec.Cluster.Spec.KubernetesVersion = "1.22"

	validator := NewValidator(cmk, &DummyNetClient{}, true)

	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsCPError(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	clusterSpec.CloudStackMachineConfigs["test-cp"].Spec.Template.Name = "kubernetes_1_22"

	validator := NewValidator(cmk, &DummyNetClient{}, true)

	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsEtcdError(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	clusterSpec.CloudStackMachineConfigs["test-etcd"].Spec.Template.Name = "kubernetes_1_22"

	validator := NewValidator(cmk, &DummyNetClient{}, true)

	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsModularUpgradeError(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	kube122 := v1alpha1.KubernetesVersion("1.22")
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122

	validator := NewValidator(cmk, &DummyNetClient{}, true)

	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsSuccess(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))

	clusterSpec.Cluster.Spec.KubernetesVersion = "1.22"
	clusterSpec.CloudStackMachineConfigs["test-cp"].Spec.Template.Name = "kubernetes_1_22"
	clusterSpec.CloudStackMachineConfigs["test-etcd"].Spec.Template.Name = "kubernetes_1_22"
	clusterSpec.CloudStackMachineConfigs["test"].Spec.Template.Name = "kubernetes_1_22"

	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	validator := NewValidator(cmk, &DummyNetClient{}, true)

	cmk.EXPECT().ValidateZoneAndGetId(ctx, gomock.Any(), gomock.Any()).Times(3).Return("4e3b338d-87a6-4189-b931-a1747edeea82", nil)
	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, v1alpha1.CloudStackResourceIdentifier{Name: "kubernetes_1_22"}).Times(3)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, gomock.Any()).Times(3)

	err = validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("validation should pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsModularUpgradeSuccess(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))

	kube122 := v1alpha1.KubernetesVersion("1.22")
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122
	clusterSpec.CloudStackMachineConfigs["test"].Spec.Template.Name = "kubernetes_1_22"

	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, testClusterConfigMainFilename))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file")
	}

	validator := NewValidator(cmk, &DummyNetClient{}, true)

	cmk.EXPECT().ValidateZoneAndGetId(ctx, gomock.Any(), gomock.Any()).Times(3).Return("4e3b338d-87a6-4189-b931-a1747edeea82", nil)
	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, v1alpha1.CloudStackResourceIdentifier{Name: "kubernetes_1_22"})
	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, testTemplate).Times(2)
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).Times(3)
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), datacenterConfig.Spec.AvailabilityZones[0].Account, gomock.Any()).Times(3)

	err = validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("validation should pass: %v", err)
	}
}

func TestValidateMachineConfigsWithAffinity(t *testing.T) {
	ctx := context.Background()
	cmk := mocks.NewMockProviderCmkClient(gomock.NewController(t))
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	controlPlaneMachineConfig(clusterSpec).Spec.Affinity = "pro"
	controlPlaneMachineConfig(clusterSpec).Spec.AffinityGroupIds = []string{}
	etcdMachineConfig(clusterSpec).Spec.Affinity = "anti"
	etcdMachineConfig(clusterSpec).Spec.AffinityGroupIds = []string{}
	for _, machineConfig := range clusterSpec.CloudStackMachineConfigs {
		machineConfig.Spec.Affinity = "no"
		machineConfig.Spec.AffinityGroupIds = []string{}
	}

	validator := NewValidator(cmk, &DummyNetClient{}, true)
	cmk.EXPECT().ValidateZoneAndGetId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("4e3b338d-87a6-4189-b931-a1747edeea8f", nil)
	cmk.EXPECT().ValidateDomainAndGetId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAccountPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	cmk.EXPECT().ValidateNetworkPresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	cmk.EXPECT().GetManagementApiEndpoint(gomock.Any()).AnyTimes().Return("http://127.16.0.1:8080/client/api", nil)

	cmk.EXPECT().ValidateTemplatePresent(ctx, gomock.Any(), gomock.Any(), gomock.Any(), clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones[0].Account, testTemplate).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(ctx, gomock.Any(), gomock.Any(), testOffering).AnyTimes()
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(ctx, gomock.Any(), gomock.Any(), clusterSpec.CloudStackDatacenter.Spec.AvailabilityZones[0].Account, gomock.Any()).AnyTimes()

	// Valid affinity types
	err := validator.ValidateClusterMachineConfigs(ctx, clusterSpec)
	assert.Nil(t, err)
}

func TestValidateSecretsUnchangedSuccess(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cmk := mocks.NewMockProviderCmkClient(mockCtrl)
	validator := NewValidator(cmk, &DummyNetClient{}, true)

	cluster := &types.Cluster{
		Name: "test",
	}

	kubectl.EXPECT().GetSecretFromNamespace(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedSecret, nil)
	err := validator.ValidateSecretsUnchanged(ctx, cluster, testExecConfig, kubectl)
	assert.Nil(t, err)
}

func TestValidateSecretsUnchangedFailureSecretChanged(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cmk := mocks.NewMockProviderCmkClient(mockCtrl)
	validator := NewValidator(cmk, &DummyNetClient{}, true)

	cluster := &types.Cluster{
		Name: "test",
	}

	modifiedSecret := expectedSecret.DeepCopy()
	modifiedSecret.Data["api-key"] = []byte("updated-api-key")

	kubectl.EXPECT().GetSecretFromNamespace(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(modifiedSecret, nil)
	err := validator.ValidateSecretsUnchanged(ctx, cluster, testExecConfig, kubectl)
	thenErrorExpected(t, "profile 'global' is different from the secret", err)
}

func TestValidateSecretsUnchangedFailureGettingSecret(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cmk := mocks.NewMockProviderCmkClient(mockCtrl)
	validator := NewValidator(cmk, &DummyNetClient{}, true)

	cluster := &types.Cluster{
		Name: "test",
	}

	kubectl.EXPECT().GetSecretFromNamespace(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, apierrors.NewBadRequest("test-error"))
	err := validator.ValidateSecretsUnchanged(ctx, cluster, testExecConfig, kubectl)
	thenErrorExpected(t, "getting secret for profile global: test-error", err)
}

func TestValidateSecretsUnchangedFailureSecretNotFound(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cmk := mocks.NewMockProviderCmkClient(mockCtrl)
	validator := NewValidator(cmk, &DummyNetClient{}, true)

	cluster := &types.Cluster{
		Name: "test",
	}

	kubectl.EXPECT().GetSecretFromNamespace(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, notFoundError)
	err := validator.ValidateSecretsUnchanged(ctx, cluster, testExecConfig, kubectl)
	assert.Nil(t, err)
}

var testProfiles = []decoder.CloudStackProfileConfig{
	{
		Name:          "global",
		ApiKey:        "test-key1",
		SecretKey:     "test-secret1",
		ManagementUrl: "http://127.16.0.1:8080/client/api",
		VerifySsl:     "false",
	},
}

var testExecConfig = &decoder.CloudStackExecConfig{
	Profiles: testProfiles,
}
