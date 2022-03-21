package snow

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/snow/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	expectedSnowProviderName = "snow"
	credsFileEnvVar          = "EKSA_SNOW_DEVICES_CREDENTIALS_FILE"
	credsFilePath            = "testdata/credentials"
	certsFileEnvVar          = "EKSA_SNOW_DEVICES_CA_BUNDLES_FILE"
	certsFilePath            = "testdata/certificates"
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
	cluster := &types.Cluster{
		Name: "cluster",
	}
	provider := newProvider(t, kubectl)
	return snowTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		kubectl:     kubectl,
		provider:    provider,
		cluster:     cluster,
		clusterSpec: givenClusterSpec(),
	}
}

func givenClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "snow-test",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.ClusterSpec{
				ClusterNetwork: v1alpha1.ClusterNetwork{
					CNI: v1alpha1.Cilium,
					Pods: v1alpha1.Pods{
						CidrBlocks: []string{
							"10.1.0.0/16",
						},
					},
					Services: v1alpha1.Services{
						CidrBlocks: []string{
							"10.96.0.0/12",
						},
					},
				},
				ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
					Count: 3,
					Endpoint: &v1alpha1.Endpoint{
						Host: "1.2.3.4",
					},
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "SnowMachineConfig",
						Name: "test-cp",
					},
				},
				KubernetesVersion: "1.21",
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
					{
						Name:  "md-0",
						Count: 3,
						MachineGroupRef: &v1alpha1.Ref{
							Kind: "SnowMachineConfig",
							Name: "test-wn",
						},
					},
				},
				DatacenterRef: v1alpha1.Ref{
					Kind: "SnowDatacenterConfig",
					Name: "test",
				},
			},
		}
		s.SnowDatacenter = givenDatacenterConfig()
		s.SnowMachineConfigs = givenMachineConfigs()
		s.VersionsBundle = &cluster.VersionsBundle{
			KubeDistro: &cluster.KubeDistro{
				Kubernetes: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/kubernetes",
					Tag:        "v1.21.5-eks-1-21-9",
				},
				CoreDNS: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/coredns",
					Tag:        "v1.8.4-eks-1-21-9",
				},
				Etcd: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/etcd-io",
					Tag:        "v3.4.16-eks-1-21-9",
				},
			},
			VersionsBundle: &releasev1alpha1.VersionsBundle{
				KubeVersion: "1.21",
				Snow: releasev1alpha1.SnowBundle{
					Version: "v1.0.2",
					KubeVip: releasev1alpha1.Image{
						Name:        "kube-vip",
						OS:          "linux",
						URI:         "public.ecr.aws/l0g8r8j6/plunder-app/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433",
						ImageDigest: "sha256:cf324971db7696810effd5c6c95e34b2c115893e1fbcaeb8877355dc74768ef1",
						Description: "Container image for kube-vip image",
						Arch:        []string{"amd64"},
					},
				},
			},
		}
		s.ManagementCluster = &types.Cluster{
			Name:           "test-snow",
			KubeconfigFile: "management.kubeconfig",
		}
	})
}

func givenDatacenterConfig() *v1alpha1.SnowDatacenterConfig {
	return &v1alpha1.SnowDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SnowDatacenterConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-namespace",
		},
		Spec: v1alpha1.SnowDatacenterConfigSpec{},
	}
}

func givenMachineConfigs() map[string]*v1alpha1.SnowMachineConfig {
	return map[string]*v1alpha1.SnowMachineConfig{
		"test-cp": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cp",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.SnowMachineConfigSpec{
				AMIID:                    "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea",
				InstanceType:             "sbe-c.large",
				SshKeyName:               "default",
				PhysicalNetworkConnector: "SFP_PLUS",
			},
		},
		"test-wn": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-wn",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.SnowMachineConfigSpec{
				AMIID:                    "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea",
				InstanceType:             "sbe-c.xlarge",
				SshKeyName:               "default",
				PhysicalNetworkConnector: "SFP_PLUS",
			},
		},
	}
}

func givenProvider(t *testing.T) *snowProvider {
	return newProvider(t, nil)
}

func givenEmptyClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.KubeVersion = "1.21"
	})
}

func newProvider(t *testing.T, kubectl ProviderKubectlClient) *snowProvider {
	return NewProvider(
		kubectl,
		nil,
		test.FakeNow,
	)
}

func setupContext(t *testing.T) {
	credsFileOrgVal, isSet := os.LookupEnv(credsFileEnvVar)
	os.Setenv(credsFileEnvVar, credsFilePath)
	t.Cleanup(func() {
		if isSet {
			os.Setenv(credsFileEnvVar, credsFileOrgVal)
		} else {
			os.Unsetenv(credsFileEnvVar)
		}
	})

	certsFileOrgVal, isSet := os.LookupEnv(certsFileEnvVar)
	os.Setenv(certsFileEnvVar, certsFilePath)
	t.Cleanup(func() {
		if isSet {
			os.Setenv(certsFileEnvVar, certsFileOrgVal)
		} else {
			os.Unsetenv(certsFileEnvVar)
		}
	})
}

func TestName(t *testing.T) {
	tt := newSnowTest(t)
	tt.Expect(tt.provider.Name()).To(Equal(expectedSnowProviderName))
}

func TestSetupAndValidateCreateClusterSuccess(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(Succeed())
}

func TestSetupAndValidateCreateClusterNoCredsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(credsFileEnvVar)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("EKSA_SNOW_DEVICES_CREDENTIALS_FILE is not set or is empty")))
}

func TestSetupAndValidateCreateClusterNoCertsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(certsFileEnvVar)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("EKSA_SNOW_DEVICES_CA_BUNDLES_FILE is not set or is empty")))
}

func TestSetupAndValidateDeleteClusterSuccess(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx)
	tt.Expect(err).To(Succeed())
}

func TestSetupAndValidateDeleteClusterNoCredsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(credsFileEnvVar)
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx)
	tt.Expect(err).To(MatchError(ContainSubstring("EKSA_SNOW_DEVICES_CREDENTIALS_FILE is not set or is empty")))
}

func TestSetupAndValidateDeleteClusterNoCertsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(certsFileEnvVar)
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx)
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
	tt.Expect(tt.provider.DatacenterConfig(tt.clusterSpec).Kind()).To(Equal("SnowDatacenterConfig"))
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
	want := tt.provider.MachineConfigs(tt.clusterSpec)
	tt.Expect(len(want)).To(Equal(2))
}

func TestDeleteResources(t *testing.T) {
	tt := newSnowTest(t)
	tt.kubectl.EXPECT().DeleteEksaDatacenterConfig(
		tt.ctx,
		snowDatacenterResourceType,
		tt.clusterSpec.SnowDatacenter.Name,
		tt.clusterSpec.ManagementCluster.KubeconfigFile,
		tt.clusterSpec.SnowDatacenter.Namespace,
	).Return(nil)
	tt.kubectl.EXPECT().DeleteEksaMachineConfig(
		tt.ctx,
		snowMachineResourceType,
		gomock.Any(),
		tt.clusterSpec.ManagementCluster.KubeconfigFile,
		gomock.Any(),
	).Times(2).Return(nil)

	err := tt.provider.DeleteResources(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(Succeed())
}
