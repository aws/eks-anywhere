package snow_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	kubemock "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
	"github.com/aws/eks-anywhere/pkg/providers/snow/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	expectedSnowProviderName = "snow"
	credsFileEnvVar          = "EKSA_AWS_CREDENTIALS_FILE"
	credsFilePath            = "testdata/credentials"
	certsFileEnvVar          = "EKSA_AWS_CA_BUNDLES_FILE"
	certsFilePath            = "testdata/certificates"
)

type snowTest struct {
	*WithT
	ctx              context.Context
	kubeUnAuthClient *mocks.MockKubeUnAuthClient
	kubeconfigClient *kubemock.MockClient
	aws              *mocks.MockAwsClient
	provider         *snow.SnowProvider
	cluster          *types.Cluster
	clusterSpec      *cluster.Spec
}

func newSnowTest(t *testing.T) snowTest {
	ctrl := gomock.NewController(t)
	mockKubeUnAuthClient := mocks.NewMockKubeUnAuthClient(ctrl)
	mockKubeconfigClient := kubemock.NewMockClient(ctrl)
	mockaws := mocks.NewMockAwsClient(ctrl)
	cluster := &types.Cluster{
		Name: "cluster",
	}
	provider := newProvider(t, mockKubeUnAuthClient, mockaws)
	return snowTest{
		WithT:            NewWithT(t),
		ctx:              context.Background(),
		kubeUnAuthClient: mockKubeUnAuthClient,
		kubeconfigClient: mockKubeconfigClient,
		aws:              mockaws,
		provider:         provider,
		cluster:          cluster,
		clusterSpec:      givenClusterSpec(),
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
						URI:         "public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433",
						ImageDigest: "sha256:cf324971db7696810effd5c6c95e34b2c115893e1fbcaeb8877355dc74768ef1",
						Description: "Container image for kube-vip image",
						Arch:        []string{"amd64"},
					},
					Manager: releasev1alpha1.Image{
						Name:        "cluster-api-snow-controller",
						OS:          "linux",
						URI:         "public.ecr.aws/l0g8r8j6/aws/cluster-api-provider-aws-snow/manager:v0.1.4-eks-a-v0.0.0-dev-build.2216",
						ImageDigest: "sha256:59da9c726c4816c29d119e77956c6391e2dff451daf36aeb60e5d6425eb88018",
						Description: "Container image for cluster-api-snow-controller image",
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
				Devices: []string{
					"1.2.3.4",
					"1.2.3.5",
				},
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
				Devices: []string{
					"1.2.3.4",
					"1.2.3.5",
				},
			},
		},
	}
}

func givenProvider(t *testing.T) *snow.SnowProvider {
	return newProvider(t, nil, nil)
}

func givenEmptyClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.KubeVersion = "1.21"
	})
}

func newProvider(t *testing.T, kubeUnAuthClient snow.KubeUnAuthClient, mockaws *mocks.MockAwsClient) *snow.SnowProvider {
	awsClients := snow.AwsClientMap{
		"1.2.3.4": mockaws,
		"1.2.3.5": mockaws,
	}
	validator := snow.NewValidatorFromAwsClientMap(awsClients)
	defaulters := snow.NewDefaultersFromAwsClientMap(awsClients, nil, nil)
	configManager := snow.NewConfigManager(defaulters, validator)
	return snow.NewProvider(
		kubeUnAuthClient,
		configManager,
		false,
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
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(Succeed())
}

func TestSetupAndValidateCreateClusterNoCredsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(credsFileEnvVar)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("'EKSA_AWS_CREDENTIALS_FILE' is not set or is empty")))
}

func TestSetupAndValidateCreateClusterNoCertsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(certsFileEnvVar)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("'EKSA_AWS_CA_BUNDLES_FILE' is not set or is empty")))
}

func TestSetupAndValidateUpgradeClusterSuccess(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	err := tt.provider.SetupAndValidateUpgradeCluster(tt.ctx, tt.cluster, tt.clusterSpec, tt.clusterSpec)
	tt.Expect(err).To(Succeed())
}

func TestSetupAndValidateUpgradeClusterNoCredsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(credsFileEnvVar)
	err := tt.provider.SetupAndValidateUpgradeCluster(tt.ctx, tt.cluster, tt.clusterSpec, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("'EKSA_AWS_CREDENTIALS_FILE' is not set or is empty")))
}

func TestSetupAndValidateUpgradeClusterNoCertsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(certsFileEnvVar)
	err := tt.provider.SetupAndValidateUpgradeCluster(tt.ctx, tt.cluster, tt.clusterSpec, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("'EKSA_AWS_CA_BUNDLES_FILE' is not set or is empty")))
}

func TestSetupAndValidateDeleteClusterSuccess(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx, tt.cluster)
	tt.Expect(err).To(Succeed())
}

func TestSetupAndValidateDeleteClusterNoCredsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(credsFileEnvVar)
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("'EKSA_AWS_CREDENTIALS_FILE' is not set or is empty")))
}

func TestSetupAndValidateDeleteClusterNoCertsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(certsFileEnvVar)
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("'EKSA_AWS_CA_BUNDLES_FILE' is not set or is empty")))
}

// TODO: add more tests (multi worker node groups, unstacked etcd, etc.)
func TestGenerateCAPISpecForCreate(t *testing.T) {
	tt := newSnowTest(t)
	tt.kubeUnAuthClient.EXPECT().KubeconfigClient(tt.cluster.KubeconfigFile).Return(tt.kubeconfigClient)
	tt.kubeconfigClient.EXPECT().
		Get(
			tt.ctx,
			"snow-test",
			constants.EksaSystemNamespace,
			&controlplanev1.KubeadmControlPlane{},
		).
		Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.kubeconfigClient.EXPECT().
		Get(
			tt.ctx,
			"snow-test-md-0",
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))

	cp, md, err := tt.provider.GenerateCAPISpecForCreate(tt.ctx, tt.cluster, tt.clusterSpec)

	tt.Expect(err).To(Succeed())
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md.yaml")
}

func TestGenerateCAPISpecForUpgrade(t *testing.T) {
	tt := newSnowTest(t)
	mt := wantSnowMachineTemplate()
	tt.kubeUnAuthClient.EXPECT().KubeconfigClient(tt.cluster.KubeconfigFile).Return(tt.kubeconfigClient)
	tt.kubeconfigClient.EXPECT().
		Get(
			tt.ctx,
			"snow-test",
			constants.EksaSystemNamespace,
			&controlplanev1.KubeadmControlPlane{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *controlplanev1.KubeadmControlPlane) error {
			obj.Spec.MachineTemplate.InfrastructureRef.Name = "snow-test-control-plane-1"
			return nil
		})
	tt.kubeconfigClient.EXPECT().
		Get(
			tt.ctx,
			"snow-test-control-plane-1",
			constants.EksaSystemNamespace,
			&snowv1.AWSSnowMachineTemplate{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *snowv1.AWSSnowMachineTemplate) error {
			wantSnowMachineTemplate().DeepCopyInto(obj)
			obj.SetName("snow-test-control-plane-1")
			obj.Spec.Template.Spec.InstanceType = "sbe-c.large"
			return nil
		})
	tt.kubeconfigClient.EXPECT().
		Get(
			tt.ctx,
			"snow-test-md-0",
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *clusterv1.MachineDeployment) error {
			wantMachineDeployment().DeepCopyInto(obj)
			obj.Spec.Template.Spec.InfrastructureRef.Name = "snow-test-md-0-1"
			obj.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "snow-test-md-0-1"
			return nil
		})
	tt.kubeconfigClient.EXPECT().
		Get(
			tt.ctx,
			"snow-test-md-0-1",
			constants.EksaSystemNamespace,
			&bootstrapv1.KubeadmConfigTemplate{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *bootstrapv1.KubeadmConfigTemplate) error {
			wantKubeadmConfigTemplate().DeepCopyInto(obj)
			return nil
		})
	tt.kubeconfigClient.EXPECT().
		Get(
			tt.ctx,
			"snow-test-md-0-1",
			constants.EksaSystemNamespace,
			&snowv1.AWSSnowMachineTemplate{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *snowv1.AWSSnowMachineTemplate) error {
			mt.DeepCopyInto(obj)
			obj.SetName("snow-test-md-0-1")
			return nil
		})

	gotCp, gotMd, err := tt.provider.GenerateCAPISpecForUpgrade(tt.ctx, tt.cluster, nil, nil, tt.clusterSpec)
	tt.Expect(err).To(Succeed())
	test.AssertContentToFile(t, string(gotCp), "testdata/expected_results_main_cp.yaml")
	test.AssertContentToFile(t, string(gotMd), "testdata/expected_results_main_md.yaml")
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

func TestGenerateMHC(t *testing.T) {
	tt := newSnowTest(t)
	got, err := tt.provider.GenerateMHC(tt.clusterSpec)
	tt.Expect(err).To(Succeed())
	test.AssertContentToFile(t, string(got), "testdata/expected_results_machine_health_check.yaml")
}

func TestDeleteResources(t *testing.T) {
	tt := newSnowTest(t)
	tt.kubeUnAuthClient.EXPECT().Delete(
		tt.ctx,
		tt.clusterSpec.SnowDatacenter.Name,
		tt.clusterSpec.SnowDatacenter.Namespace,
		tt.clusterSpec.ManagementCluster.KubeconfigFile,
		tt.clusterSpec.SnowDatacenter,
	).Return(nil)
	tt.kubeUnAuthClient.EXPECT().Delete(
		tt.ctx,
		tt.clusterSpec.SnowMachineConfigs["test-cp"].Name,
		tt.clusterSpec.SnowMachineConfigs["test-cp"].Namespace,
		tt.clusterSpec.ManagementCluster.KubeconfigFile,
		tt.clusterSpec.SnowMachineConfigs["test-cp"],
	).Return(nil)
	tt.kubeUnAuthClient.EXPECT().Delete(
		tt.ctx,
		tt.clusterSpec.SnowMachineConfigs["test-wn"].Name,
		tt.clusterSpec.SnowMachineConfigs["test-wn"].Namespace,
		tt.clusterSpec.ManagementCluster.KubeconfigFile,
		tt.clusterSpec.SnowMachineConfigs["test-wn"],
	).Return(nil)

	err := tt.provider.DeleteResources(tt.ctx, tt.clusterSpec)
	tt.Expect(err).To(Succeed())
}

func TestUpgradeNeededFalse(t *testing.T) {
	tt := newSnowTest(t)
	got, err := tt.provider.UpgradeNeeded(tt.ctx, tt.clusterSpec, tt.clusterSpec, tt.cluster)
	tt.Expect(err).To(Succeed())
	tt.Expect(got).To(Equal(false))
}

func TestUpgradeNeededBundle(t *testing.T) {
	tests := []struct {
		name   string
		bundle releasev1alpha1.SnowBundle
		want   bool
	}{
		{
			name: "non compared fields diff",
			bundle: releasev1alpha1.SnowBundle{
				Version: "v1.0.2-diff",
				KubeVip: releasev1alpha1.Image{
					Name:        "kube-vip-diff",
					OS:          "linux-diff",
					URI:         "public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433-diff",
					ImageDigest: "sha256:cf324971db7696810effd5c6c95e34b2c115893e1fbcaeb8877355dc74768ef1",
					Description: "Container image for kube-vip image-diff",
					Arch:        []string{"amd64-diff"},
				},
				Manager: releasev1alpha1.Image{
					Name:        "cluster-api-snow-controller-diff",
					OS:          "linux-diff",
					URI:         "public.ecr.aws/l0g8r8j6/aws/cluster-api-provider-aws-snow/manager:v0.1.4-eks-a-v0.0.0-dev-build.2216-diff",
					ImageDigest: "sha256:59da9c726c4816c29d119e77956c6391e2dff451daf36aeb60e5d6425eb88018",
					Description: "Container image for cluster-api-snow-controller image-diff",
					Arch:        []string{"amd64-diff"},
				},
			},
			want: false,
		},
		{
			name: "kube-vip image digest diff",
			bundle: releasev1alpha1.SnowBundle{
				Version: "v1.0.2",
				KubeVip: releasev1alpha1.Image{
					Name:        "kube-vip",
					OS:          "linux",
					URI:         "public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433",
					ImageDigest: "sha256:diff",
					Description: "Container image for kube-vip image",
					Arch:        []string{"amd64"},
				},
				Manager: releasev1alpha1.Image{
					Name:        "cluster-api-snow-controller",
					OS:          "linux",
					URI:         "public.ecr.aws/l0g8r8j6/aws/cluster-api-provider-aws-snow/manager:v0.1.4-eks-a-v0.0.0-dev-build.2216",
					ImageDigest: "sha256:59da9c726c4816c29d119e77956c6391e2dff451daf36aeb60e5d6425eb88018",
					Description: "Container image for cluster-api-snow-controller image",
					Arch:        []string{"amd64"},
				},
			},
			want: true,
		},
		{
			name: "manager image digest diff",
			bundle: releasev1alpha1.SnowBundle{
				Version: "v1.0.2",
				KubeVip: releasev1alpha1.Image{
					Name:        "kube-vip",
					OS:          "linux",
					URI:         "public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433",
					ImageDigest: "sha256:cf324971db7696810effd5c6c95e34b2c115893e1fbcaeb8877355dc74768ef1",
					Description: "Container image for kube-vip image",
					Arch:        []string{"amd64"},
				},
				Manager: releasev1alpha1.Image{
					Name:        "cluster-api-snow-controller",
					OS:          "linux",
					URI:         "public.ecr.aws/l0g8r8j6/aws/cluster-api-provider-aws-snow/manager:v0.1.4-eks-a-v0.0.0-dev-build.2216",
					ImageDigest: "sha256:diff",
					Description: "Container image for cluster-api-snow-controller image",
					Arch:        []string{"amd64"},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newSnowTest(t)
			new := g.clusterSpec.DeepCopy()
			new.VersionsBundle.Snow = tt.bundle
			new.SnowMachineConfigs = givenMachineConfigs()
			got, err := g.provider.UpgradeNeeded(g.ctx, new, g.clusterSpec, g.cluster)
			g.Expect(err).To(Succeed())
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestUpgradeNeededMachineConfigs(t *testing.T) {
	tests := []struct {
		name           string
		machineConfigs map[string]*v1alpha1.SnowMachineConfig
		want           bool
	}{
		{
			name: "non compared fields diff",
			machineConfigs: map[string]*v1alpha1.SnowMachineConfig{
				"test-cp": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cp-diff",
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
						Namespace: "test-namespace-diff",
					},
					Spec: v1alpha1.SnowMachineConfigSpec{
						AMIID:                    "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea",
						InstanceType:             "sbe-c.xlarge",
						SshKeyName:               "default",
						PhysicalNetworkConnector: "SFP_PLUS",
					},
				},
			},
			want: false,
		},
		{
			name: "spec diff",
			machineConfigs: map[string]*v1alpha1.SnowMachineConfig{
				"test-cp": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cp",
						Namespace: "test-namespace",
					},
					Spec: v1alpha1.SnowMachineConfigSpec{
						AMIID:                    "eks-d-v1-21-5-ubuntu-ami-0",
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
						AMIID:                    "eks-d-v1-21-5-ubuntu-ami-1",
						InstanceType:             "sbe-c.xlarge",
						SshKeyName:               "default",
						PhysicalNetworkConnector: "SFP_PLUS",
					},
				},
			},
			want: true,
		},
		{
			name: "length diff",
			machineConfigs: map[string]*v1alpha1.SnowMachineConfig{
				"test-cp": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cp-diff",
						Namespace: "test-namespace",
					},
					Spec: v1alpha1.SnowMachineConfigSpec{
						AMIID:                    "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea",
						InstanceType:             "sbe-c.large",
						SshKeyName:               "default",
						PhysicalNetworkConnector: "SFP_PLUS",
					},
				},
			},
			want: true,
		},
		{
			name: "key diff",
			machineConfigs: map[string]*v1alpha1.SnowMachineConfig{
				"test-cp-diff": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cp-diff",
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
						Namespace: "test-namespace-diff",
					},
					Spec: v1alpha1.SnowMachineConfigSpec{
						AMIID:                    "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea",
						InstanceType:             "sbe-c.xlarge",
						SshKeyName:               "default",
						PhysicalNetworkConnector: "SFP_PLUS",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newSnowTest(t)
			new := g.clusterSpec.DeepCopy()
			new.SnowMachineConfigs = tt.machineConfigs
			got, err := g.provider.UpgradeNeeded(g.ctx, new, g.clusterSpec, g.cluster)
			g.Expect(err).To(Succeed())
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestChangeDiffNoChange(t *testing.T) {
	g := NewWithT(t)
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	g.Expect(provider.ChangeDiff(clusterSpec, clusterSpec)).To(BeNil())
}

func TestChangeDiffWithChange(t *testing.T) {
	g := NewWithT(t)
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	newClusterSpec := clusterSpec.DeepCopy()
	clusterSpec.VersionsBundle.Snow.Version = "v1.0.2"
	newClusterSpec.VersionsBundle.Snow.Version = "v1.0.3"
	want := &types.ComponentChangeDiff{
		ComponentName: "snow",
		NewVersion:    "v1.0.3",
		OldVersion:    "v1.0.2",
	}
	g.Expect(provider.ChangeDiff(clusterSpec, newClusterSpec)).To(Equal(want))
}
