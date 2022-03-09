package snow

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/docker/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	expectedSnowProviderName  = "snow"
	testDataDir               = "testdata"
	testClusterConfigFilename = "cluster_main_1_21.yaml"
	credsFileEnvVar           = "EKSA_SNOW_DEVICES_CREDENTIALS_FILE"
	credsFilePath             = "testdata/credentials"
	certsFileEnvVar           = "EKSA_SNOW_DEVICES_CA_BUNDLES_FILE"
	certsFilePath             = "testdata/certificates"
)

type snowTest struct {
	*WithT
	ctx         context.Context
	kubectl     *mocks.MockProviderKubectlClient
	provider    *snowProvider
	cluster     *types.Cluster
	clusterSpec *cluster.Spec
}

func newSnowTest(t *testing.T) snowTest {
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(ctrl)
	o := givenFullAPIObjects(t, testClusterConfigFilename)
	cluster := &types.Cluster{
		Name: "cluster",
	}
	provider := newProvider(t, o.datacenterConfig, o.machineConfigs, o.cluster, kubectl)
	return snowTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		kubectl:     kubectl,
		provider:    provider,
		cluster:     cluster,
		clusterSpec: o.clusterSpec,
	}
}

type apiObjects struct {
	cluster          *v1alpha1.Cluster
	clusterSpec      *cluster.Spec
	datacenterConfig *v1alpha1.SnowDatacenterConfig
	machineConfigs   map[string]*v1alpha1.SnowMachineConfig
}

func givenFullAPIObjects(t *testing.T, fileName string) apiObjects {
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, fileName))
	datacenterConfig, err := v1alpha1.GetSnowDatacenterConfig(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get snow datacenter config from file: %v", err)
	}
	machineConfigs, err := v1alpha1.GetSnowMachineConfigs(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get snow machine configs from file")
	}

	return apiObjects{
		cluster:          clusterSpec.Cluster,
		clusterSpec:      clusterSpec,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
	}
}

func givenProvider(t *testing.T) *snowProvider {
	return newProvider(t, nil, nil, nil, nil)
}

func givenEmptyClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.KubeVersion = "1.21"
	})
}

func newProvider(t *testing.T, datacenterConfig *v1alpha1.SnowDatacenterConfig, machineConfigs map[string]*v1alpha1.SnowMachineConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient) *snowProvider {
	return NewProvider(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		kubectl,
		nil,
		test.FakeNow,
	)
}

func setupContext(t *testing.T, envVar string) {
	envVarOrgValue, isSet := os.LookupEnv(envVar)
	t.Cleanup(func() {
		if isSet {
			os.Setenv(envVar, envVarOrgValue)
		} else {
			os.Unsetenv(envVar)
		}
	})
}

func TestName(t *testing.T) {
	tt := newSnowTest(t)
	tt.Expect(tt.provider.Name()).To(Equal(expectedSnowProviderName))
}

func TestSetupAndValidateCreateClusterSuccess(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t, credsFileEnvVar)
	setupContext(t, certsFileEnvVar)
	os.Setenv(credsFileEnvVar, credsFilePath)
	os.Setenv(certsFileEnvVar, certsFilePath)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(Succeed())
}

func TestSetupAndValidateCreateClusterNoCredsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t, credsFileEnvVar)
	setupContext(t, certsFileEnvVar)
	os.Setenv(certsFileEnvVar, certsFilePath)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("EKSA_SNOW_DEVICES_CREDENTIALS_FILE is not set or is empty")))
}

func TestSetupAndValidateCreateClusterNoCertsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t, credsFileEnvVar)
	os.Setenv(credsFileEnvVar, credsFilePath)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("EKSA_SNOW_DEVICES_CA_BUNDLES_FILE is not set or is empty")))
}

// TODO: add more tests (multi worker node groups, unstacked etcd, etc.)
func TestGenerateCAPISpecForCreate(t *testing.T) {
	tt := newSnowTest(t)
	cp, md, err := tt.provider.GenerateCAPISpecForCreate(tt.ctx, tt.cluster, tt.clusterSpec)
	tt.Expect(err).To(Succeed())
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md.yaml")
}

func TestVersion(t *testing.T) {
	snowVersion := "v1.0.2"
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	clusterSpec.VersionsBundle.Snow.Version = snowVersion
	g := NewWithT(t)
	result := provider.Version(clusterSpec)
	g.Expect(result).To(Equal(snowVersion))
}

func TestGetInfrastructureBundle(t *testing.T) {
	tt := newSnowTest(t)
	want := &types.InfrastructureBundle{
		FolderName: "infrastructure-snow/v1.0.2/",
		Manifests: []releasev1alpha1.Manifest{
			tt.clusterSpec.VersionsBundle.Snow.Components,
			tt.clusterSpec.VersionsBundle.Snow.Metadata,
		},
	}
	got := tt.provider.GetInfrastructureBundle(tt.clusterSpec)
	tt.Expect(got).To(Equal(want))
}

func TestGetDatacenterConfig(t *testing.T) {
	tt := newSnowTest(t)
	tt.Expect(tt.provider.DatacenterConfig().Kind()).To(Equal("SnowDatacenterConfig"))
}

func TestDatacenterResourceType(t *testing.T) {
	g := NewWithT(t)
	provider := givenProvider(t)
	g.Expect(provider.DatacenterResourceType()).To(Equal("snowdatacenterconfigs.anywhere.eks.amazonaws.com"))
}

func TestMachineResourceType(t *testing.T) {
	g := NewWithT(t)
	provider := givenProvider(t)
	g.Expect(provider.MachineResourceType()).To(Equal("snowmachineconfigs.anywhere.eks.amazonaws.com"))
}

func TestMachineConfigs(t *testing.T) {
	tt := newSnowTest(t)
	want := tt.provider.MachineConfigs()
	tt.Expect(len(want)).To(Equal(2))
}
