package snow_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
	kubemock "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	"github.com/aws/eks-anywhere/pkg/providers/snow/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
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
	ctrl             *gomock.Controller
	kubeUnAuthClient *mocks.MockKubeUnAuthClient
	kubeconfigClient *kubemock.MockClient
	aws              *mocks.MockAwsClient
	imds             *mocks.MockLocalIMDSClient
	provider         *snow.SnowProvider
	cluster          *types.Cluster
	clusterSpec      *cluster.Spec
	logger           logr.Logger
}

func newSnowTest(t *testing.T) snowTest {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	mockKubeUnAuthClient := mocks.NewMockKubeUnAuthClient(ctrl)
	mockKubeconfigClient := kubemock.NewMockClient(ctrl)
	mockaws := mocks.NewMockAwsClient(ctrl)
	mockimds := mocks.NewMockLocalIMDSClient(ctrl)
	cluster := &types.Cluster{
		Name: "cluster",
	}
	provider := newProvider(ctx, t, mockKubeUnAuthClient, mockaws, mockimds, ctrl)
	return snowTest{
		WithT:            NewWithT(t),
		ctx:              ctx,
		ctrl:             ctrl,
		kubeUnAuthClient: mockKubeUnAuthClient,
		kubeconfigClient: mockKubeconfigClient,
		aws:              mockaws,
		imds:             mockimds,
		provider:         provider,
		cluster:          cluster,
		clusterSpec:      givenClusterSpec(),
		logger:           test.NewNullLogger(),
	}
}

func givenClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = givenClusterConfig()
		s.SnowDatacenter = givenDatacenterConfig()
		s.SnowCredentialsSecret = wantEksaCredentialsSecret()
		s.SnowMachineConfigs = givenMachineConfigs()
		s.SnowIPPools = givenIPPools()
		s.VersionsBundles["1.21"] = givenVersionsBundle("1.21")
		s.VersionsBundles["1.29"] = givenVersionsBundle("1.29")
		s.VersionsBundles["1.33"] = givenVersionsBundle("1.33")
		s.ManagementCluster = givenManagementCluster()
	})
}

func givenVersionsBundle(kubeVersion v1alpha1.KubernetesVersion) *cluster.VersionsBundle {
	switch kubeVersion {
	case "1.21":
		return &cluster.VersionsBundle{
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
				EtcdImage: releasev1alpha1.Image{
					URI: "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
				},
				Pause: releasev1alpha1.Image{
					URI: "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
				},
				EtcdVersion: "3.4.16",
				EtcdURL:     "https://distro.eks.amazonaws.com/kubernetes-1-21/releases/4/artifacts/etcd/v3.4.16/etcd-linux-amd64-v3.4.16.tar.gz",
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
					BottlerocketBootstrapSnow: releasev1alpha1.Image{
						Name:        "bottlerocket-bootstrap-snow",
						OS:          "linux",
						URI:         "public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap-snow:v1-20-22-eks-a-v0.0.0-dev-build.4984",
						ImageDigest: "sha256:59da9c726c4816c29d119e77956c6391e2dff451daf36aeb60e5d6425eb88018",
						Description: "Container image for bottlerocket-bootstrap-snow image",
						Arch:        []string{"amd64"},
					},
				},
				BottleRocketHostContainers: releasev1alpha1.BottlerocketHostContainersBundle{
					Admin: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-admin:0.0.1",
					},
					Control: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-control:0.0.1",
					},
					KubeadmBootstrap: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
					},
				},
			},
		}
	case "1.29":
		return &cluster.VersionsBundle{
			KubeDistro: &cluster.KubeDistro{
				Kubernetes: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/kubernetes",
					Tag:        "v1.29.0-eks-1-29-3",
				},
				CoreDNS: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/coredns",
					Tag:        "v1.11.1-eks-1-29-3",
				},
				Etcd: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/etcd-io",
					Tag:        "v3.5.10-eks-1-29-3",
				},
				EtcdImage: releasev1alpha1.Image{
					URI: "public.ecr.aws/eks-distro/etcd-io/etcd:v3.5.10-eks-1-29-3",
				},
				EtcdURL:     "https://distro.eks.amazonaws.com/kubernetes-1-29/releases/3/artifacts/etcd/v3.5.10/etcd-linux-amd64-v3.5.10.tar.gz",
				EtcdVersion: "3.5.10",
				Pause: releasev1alpha1.Image{
					URI: "public.ecr.aws/eks-distro/kubernetes/pause:v1.29.0-eks-1-29-3",
				},
			},
			VersionsBundle: &releasev1alpha1.VersionsBundle{
				KubeVersion: "1.29",
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
					BottlerocketBootstrapSnow: releasev1alpha1.Image{
						Name:        "bottlerocket-bootstrap-snow",
						OS:          "linux",
						URI:         "public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap-snow:v1-20-22-eks-a-v0.0.0-dev-build.4984",
						ImageDigest: "sha256:59da9c726c4816c29d119e77956c6391e2dff451daf36aeb60e5d6425eb88018",
						Description: "Container image for bottlerocket-bootstrap-snow image",
						Arch:        []string{"amd64"},
					},
				},
				BottleRocketHostContainers: releasev1alpha1.BottlerocketHostContainersBundle{
					Admin: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-admin:0.0.1",
					},
					Control: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-control:0.0.1",
					},
					KubeadmBootstrap: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
					},
				},
			},
		}
	case "1.33":
		return &cluster.VersionsBundle{
			KubeDistro: &cluster.KubeDistro{
				Kubernetes: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/kubernetes",
					Tag:        "v1.33.0-eks-1-33-1",
				},
				CoreDNS: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/coredns",
					Tag:        "v1.11.1-eks-1-33-1",
				},
				Etcd: cluster.VersionedRepository{
					Repository: "public.ecr.aws/eks-distro/etcd-io",
					Tag:        "v3.5.15-eks-1-33-1",
				},
				EtcdImage: releasev1alpha1.Image{
					URI: "public.ecr.aws/eks-distro/etcd-io/etcd:v3.5.15-eks-1-33-1",
				},
				EtcdURL:     "https://distro.eks.amazonaws.com/kubernetes-1-33/releases/1/artifacts/etcd/v3.5.15/etcd-linux-amd64-v3.5.15.tar.gz",
				EtcdVersion: "3.5.15",
				Pause: releasev1alpha1.Image{
					URI: "public.ecr.aws/eks-distro/kubernetes/pause:v1.33.0-eks-1-33-1",
				},
			},
			VersionsBundle: &releasev1alpha1.VersionsBundle{
				KubeVersion: "1.33",
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
					BottlerocketBootstrapSnow: releasev1alpha1.Image{
						Name:        "bottlerocket-bootstrap-snow",
						OS:          "linux",
						URI:         "public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap-snow:v1-20-22-eks-a-v0.0.0-dev-build.4984",
						ImageDigest: "sha256:59da9c726c4816c29d119e77956c6391e2dff451daf36aeb60e5d6425eb88018",
						Description: "Container image for bottlerocket-bootstrap-snow image",
						Arch:        []string{"amd64"},
					},
				},
				BottleRocketHostContainers: releasev1alpha1.BottlerocketHostContainersBundle{
					Admin: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-admin:0.0.1",
					},
					Control: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-control:0.0.1",
					},
					KubeadmBootstrap: releasev1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
					},
				},
			},
		}
	}
	return nil
}

func givenManagementComponents() *cluster.ManagementComponents {
	return &cluster.ManagementComponents{
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
			BottlerocketBootstrapSnow: releasev1alpha1.Image{
				Name:        "bottlerocket-bootstrap-snow",
				OS:          "linux",
				URI:         "public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap-snow:v1-20-22-eks-a-v0.0.0-dev-build.4984",
				ImageDigest: "sha256:59da9c726c4816c29d119e77956c6391e2dff451daf36aeb60e5d6425eb88018",
				Description: "Container image for bottlerocket-bootstrap-snow image",
				Arch:        []string{"amd64"},
			},
		},
	}
}

func givenManagementCluster() *types.Cluster {
	return &types.Cluster{
		Name:           "test-snow",
		KubeconfigFile: "management.kubeconfig",
	}
}

func givenClusterConfig() *v1alpha1.Cluster {
	devVersion := test.DevEksaVersion()
	return &v1alpha1.Cluster{
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
				UpgradeRolloutStrategy: &v1alpha1.ControlPlaneUpgradeRolloutStrategy{
					Type: "RollingUpdate",
					RollingUpdate: &v1alpha1.ControlPlaneRollingUpdateParams{
						MaxSurge: 1,
					},
				},
			},
			KubernetesVersion: "1.21",
			EksaVersion:       &devVersion,
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Name:  "md-0",
					Count: ptr.Int(3),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "SnowMachineConfig",
						Name: "test-wn",
					},
					UpgradeRolloutStrategy: &v1alpha1.WorkerNodesUpgradeRolloutStrategy{
						Type: "RollingUpdate",
						RollingUpdate: &v1alpha1.WorkerNodesRollingUpdateParams{
							MaxSurge:       1,
							MaxUnavailable: 0,
						},
					},
				},
			},
			DatacenterRef: v1alpha1.Ref{
				Kind: "SnowDatacenterConfig",
				Name: "test",
			},
		},
	}
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
		Spec: v1alpha1.SnowDatacenterConfigSpec{
			IdentityRef: v1alpha1.Ref{
				Kind: "Secret",
				Name: "test-snow-credentials",
			},
		},
	}
}

func givenMachineConfigs() map[string]*v1alpha1.SnowMachineConfig {
	return map[string]*v1alpha1.SnowMachineConfig{
		"test-cp": {
			TypeMeta: metav1.TypeMeta{
				Kind: "SnowMachineConfig",
			},
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
				OSFamily: v1alpha1.Ubuntu,
				Network: v1alpha1.SnowNetwork{
					DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
						{
							Index:   1,
							DHCP:    true,
							Primary: true,
						},
					},
				},
			},
		},
		"test-wn": {
			TypeMeta: metav1.TypeMeta{
				Kind: "SnowMachineConfig",
			},
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
				OSFamily: v1alpha1.Ubuntu,
				Network: v1alpha1.SnowNetwork{
					DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
						{
							Index:   1,
							DHCP:    true,
							Primary: true,
						},
					},
				},
			},
		},
	}
}

func givenIPPools() map[string]*v1alpha1.SnowIPPool {
	return map[string]*v1alpha1.SnowIPPool{
		"ip-pool-1": {
			TypeMeta: metav1.TypeMeta{
				Kind: snow.SnowIPPoolKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ip-pool-1",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.SnowIPPoolSpec{
				Pools: []v1alpha1.IPPool{
					{
						IPStart: "start",
						IPEnd:   "end",
						Gateway: "gateway",
						Subnet:  "subnet",
					},
				},
			},
		},
	}
}

func givenProvider(t *testing.T) *snow.SnowProvider {
	return newProvider(context.Background(), t, nil, nil, nil, gomock.NewController(t))
}

func givenEmptyClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].KubeVersion = "1.21"
	})
}

func newProvider(ctx context.Context, t *testing.T, kubeUnAuthClient snow.KubeUnAuthClient, mockaws *mocks.MockAwsClient, mockimds *mocks.MockLocalIMDSClient, ctrl *gomock.Controller) *snow.SnowProvider {
	awsClients := snow.AwsClientMap{
		"1.2.3.4": mockaws,
		"1.2.3.5": mockaws,
	}
	mockClientRegistry := mocks.NewMockClientRegistry(ctrl)
	mockClientRegistry.EXPECT().Get(ctx).Return(awsClients, nil).AnyTimes()
	validator := snow.NewValidator(mockClientRegistry, snow.WithIMDS(mockimds))
	defaulters := snow.NewDefaulters(mockClientRegistry, nil)
	configManager := snow.NewConfigManager(defaulters, validator)
	return snow.NewProvider(
		kubeUnAuthClient,
		configManager,
		false,
	)
}

func setupContext(t *testing.T) {
	t.Setenv(credsFileEnvVar, credsFilePath)
	t.Setenv(certsFileEnvVar, certsFilePath)
}

func TestName(t *testing.T) {
	tt := newSnowTest(t)
	tt.Expect(tt.provider.Name()).To(Equal(expectedSnowProviderName))
}

func wantEksaCredentialsSecretWithEnvCreds() *v1.Secret {
	secret := wantEksaCredentialsSecret()
	secret.Data["credentials"] = []byte("WzEuMi4zLjRdCmF3c19hY2Nlc3Nfa2V5X2lkID0gQUJDREVGR0hJSktMTU5PUFFSMlQKYXdzX3NlY3JldF9hY2Nlc3Nfa2V5ID0gQWZTRDdzWXovVEJadHprUmVCbDZQdXVJU3pKMld0TmtlZVB3K25OekoKcmVnaW9uID0gc25vdw==")
	secret.Data["ca-bundle"] = []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURYakNDQWthZ0F3SUJBZ0lJYjVtMFJsakpDTUV3RFFZSktvWklodmNOQVFFTkJRQXdPREUyTURRR0ExVUUKQXd3dFNrbEVMVEl3TmpnME16UXlNREF3TWkweE9USXRNVFk0TFRFdE1qTTFMVEl5TFRBeExUQTJMVEl5TFRBMApNQjRYRFRJeE1ERXhNVEl5TURjMU9Gb1hEVEkxTVRJeE5qSXlNRGMxT0Zvd09ERTJNRFFHQTFVRUF3d3RTa2xFCkxUSXdOamcwTXpReU1EQXdNaTB4T1RJdE1UWTRMVEV0TWpNMUxUSXlMVEF4TFRBMkxUSXlMVEEwTUlJQklqQU4KQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBbU9UUURCZkJ0UGNWREZnL2E1OWRrK3JZclBSVQpmNXpsN0pnRkFFdzFuODJTa2JObTRzcndsb2o4cEN1RDFuSkFsTiszTEtvaWJ5OWpVOFpxb1FLcXBwSmFLMVFLCmR2MjdKWU5sV29yRzlyNktyRmtpRVRuMmN4dUF3Y1JCdnE0VUY3NldkTnI3ekZqSTEwOGJ5UHA5UGQwbXhLaVEKNldWYXhjS1g5QUVjYXJCL0dmaWRITzk1QWF5NnRpQlUxU1F2QkpybzNMMS9VRnU1U1RTcFphaTl6eCtWa1dUSgpEMEpYaDdlTEY0eUwwTjFvVTBoWDJDR0R4RHo0VmxKbUJPdmJuUnV3c09ydVJNdFVGUlV5NTljUHpyLy80ZmpkCjRTN0FZYmVPVlB3RVA3cTE5Tlo2K1A3RTcxalRxMXJ6OFJoQW5XL0pjYlRLUzBLcWdCVVB6MFU0cVFJREFRQUIKbzJ3d2FqQU1CZ05WSFJNRUJUQURBUUgvTUIwR0ExVWREZ1FXQkJRVGFaekwyZ29xcTcvTWJKRWZOUnV6YndpaAprVEE3QmdOVkhSRUVOREF5aGpCSlJEcEtTVVF0TWpBMk9EUXpOREl3TURBeUxURTVNaTB4TmpndE1TMHlNelV0Ck1qSXRNREV0TURZdE1qSXRNRFF3RFFZSktvWklodmNOQVFFTkJRQURnZ0VCQUV6ZWwrVXNwaFV4NDlFVkF5V0IKUHpTem9FN1g2MmZnL2I0Z1U3aWZGSHBXcFlwQVBzYmFwejkvVHl3YzRUR1JJdGZjdFhZWnNqY2hKS2l1dEdVMgp6WDRydDFOU0hreDcyaU1sM29iUTJqUW1URDhmOUx5Q3F5YStRTTRDQTc0a2s2djJuZzFFaXdNWXZRbFR2V1k0CkZFV3YyMXlOUnMyeWlSdUhXalJZSDRURjU0Y0NvRFFHcEZwc09GaTBMNFYveW8xWHVpbVNMeDJ2dktaMGxDTnQKS3hDMW9DZ0N4eE5rT2EvNmlMazZxVkFOb1g1S0lWc2F0YVZodkdLKzltd1duOCtkbk1GbmVNaVdkL2p2aStkaApleXdsZFZFTEJXUktFTERkQmM5WGI0aTVCRVRGNmRVbG12cFdncE9YWE8zdUpsSVJHWkNWRkxzZ1E1MTFvTXhNCnJFQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQ==")
	return secret
}

func supportedInstanceTypes() []aws.EC2InstanceType {
	return []aws.EC2InstanceType{
		{
			Name:        "sbe-c.large",
			DefaultVCPU: ptr.Int32(2),
		},
		{
			Name:        "sbe-c.xlarge",
			DefaultVCPU: ptr.Int32(4),
		},
	}
}

func TestSetupAndValidateCreateClusterSuccess(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2InstanceTypes(tt.ctx).Return(supportedInstanceTypes(), nil).Times(4)
	tt.aws.EXPECT().IsSnowballDeviceUnlocked(tt.ctx).Return(true, nil).Times(4)
	tt.aws.EXPECT().SnowballDeviceSoftwareVersion(tt.ctx).Return("102", nil).Times(4)
	tt.imds.EXPECT().EC2InstanceIP(tt.ctx).Return("1.2.3.5", nil)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
	tt.Expect(err).To(Succeed())
}

func TestSetupAndValidateCreateClusterIMDSNotInitialized(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	tt.provider = newProvider(tt.ctx, t, tt.kubeUnAuthClient, tt.aws, nil, tt.ctrl)
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2InstanceTypes(tt.ctx).Return(supportedInstanceTypes(), nil).Times(4)
	tt.aws.EXPECT().IsSnowballDeviceUnlocked(tt.ctx).Return(true, nil).Times(4)
	tt.aws.EXPECT().SnowballDeviceSoftwareVersion(tt.ctx).Return("102", nil).Times(4)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
	tt.Expect(err).To(MatchError(ContainSubstring("imds client is not initialized")))
}

func TestSetupAndValidateCreateClusterCPIPInvalid(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2InstanceTypes(tt.ctx).Return(supportedInstanceTypes(), nil).Times(4)
	tt.aws.EXPECT().IsSnowballDeviceUnlocked(tt.ctx).Return(true, nil).Times(4)
	tt.aws.EXPECT().SnowballDeviceSoftwareVersion(tt.ctx).Return("102", nil).Times(4)
	tt.imds.EXPECT().EC2InstanceIP(tt.ctx).Return("1.2.3.4", nil)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
	tt.Expect(err).To(MatchError(ContainSubstring("control plane host ip cannot be same as the admin instance ip")))
}

func TestSetupAndValidateCreateClusterGetInstanceIPError(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2InstanceTypes(tt.ctx).Return(supportedInstanceTypes(), nil).Times(4)
	tt.aws.EXPECT().IsSnowballDeviceUnlocked(tt.ctx).Return(true, nil).Times(4)
	tt.aws.EXPECT().SnowballDeviceSoftwareVersion(tt.ctx).Return("102", nil).Times(4)
	tt.imds.EXPECT().EC2InstanceIP(tt.ctx).Return("", errors.New("fetch instance ip error"))
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
	tt.Expect(err).To(MatchError(ContainSubstring("fetch instance ip error")))
}

func TestSetupAndValidateCreateClusterGetEC2InstanceTypesError(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2InstanceTypes(tt.ctx).Return(nil, errors.New("get instance types error"))
	tt.aws.EXPECT().IsSnowballDeviceUnlocked(tt.ctx).Return(true, nil).Times(4)
	tt.aws.EXPECT().SnowballDeviceSoftwareVersion(tt.ctx).Return("102", nil).Times(4)
	tt.imds.EXPECT().EC2InstanceIP(tt.ctx).Return("1.2.3.5", nil)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
	tt.Expect(err).To(MatchError(ContainSubstring("fetching supported instance types for device [1.2.3.4]: get instance types error")))
}

func TestSetupAndValidateCreateClusterUnsupportedInstanceTypeError(t *testing.T) {
	tt := newSnowTest(t)
	instanceTypes := []aws.EC2InstanceType{
		{
			Name: "new-instance-type",
		},
	}
	setupContext(t)
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2InstanceTypes(tt.ctx).Return(instanceTypes, nil)
	tt.aws.EXPECT().IsSnowballDeviceUnlocked(tt.ctx).Return(true, nil).Times(4)
	tt.aws.EXPECT().SnowballDeviceSoftwareVersion(tt.ctx).Return("102", nil).Times(4)
	tt.imds.EXPECT().EC2InstanceIP(tt.ctx).Return("1.2.3.5", nil)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
	tt.Expect(err).To(MatchError(ContainSubstring("not supported in device [1.2.3.4]")))
}

func TestSetupAndValidateCreateClusterInstanceTypeVCPUError(t *testing.T) {
	tt := newSnowTest(t)
	instanceTypes := []aws.EC2InstanceType{
		{
			Name:        "sbe-c.large",
			DefaultVCPU: ptr.Int32(1),
		},
		{
			Name:        "sbe-c.xlarge",
			DefaultVCPU: ptr.Int32(1),
		},
	}
	setupContext(t)
	tt.aws.EXPECT().EC2ImageExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2KeyNameExists(tt.ctx, gomock.Any()).Return(true, nil).Times(4)
	tt.aws.EXPECT().EC2InstanceTypes(tt.ctx).Return(instanceTypes, nil)
	tt.aws.EXPECT().IsSnowballDeviceUnlocked(tt.ctx).Return(true, nil).Times(4)
	tt.aws.EXPECT().SnowballDeviceSoftwareVersion(tt.ctx).Return("102", nil).Times(4)
	tt.imds.EXPECT().EC2InstanceIP(tt.ctx).Return("1.2.3.5", nil)
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
	tt.Expect(err).To(MatchError(ContainSubstring("has 1 vCPU. Please choose an instance type with at least 2 default vCPU")))
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
	tt.aws.EXPECT().EC2InstanceTypes(tt.ctx).Return(supportedInstanceTypes(), nil).Times(4)
	tt.aws.EXPECT().IsSnowballDeviceUnlocked(tt.ctx).Return(true, nil).Times(4)
	tt.aws.EXPECT().SnowballDeviceSoftwareVersion(tt.ctx).Return("102", nil).Times(4)
	tt.imds.EXPECT().EC2InstanceIP(tt.ctx).Return("1.2.3.5", nil)
	err := tt.provider.SetupAndValidateUpgradeCluster(tt.ctx, tt.cluster, tt.clusterSpec, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
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
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx, tt.cluster, tt.clusterSpec)
	tt.Expect(tt.clusterSpec.SnowCredentialsSecret).To(Equal(wantEksaCredentialsSecretWithEnvCreds()))
	tt.Expect(err).To(Succeed())
}

func TestSetupAndValidateDeleteClusterNoCredsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(credsFileEnvVar)
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx, tt.cluster, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("'EKSA_AWS_CREDENTIALS_FILE' is not set or is empty")))
}

func TestSetupAndValidateDeleteClusterNoCertsEnv(t *testing.T) {
	tt := newSnowTest(t)
	setupContext(t)
	os.Unsetenv(certsFileEnvVar)
	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx, tt.cluster, tt.clusterSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("'EKSA_AWS_CA_BUNDLES_FILE' is not set or is empty")))
}

func TestVersion(t *testing.T) {
	snowVersion := "v1.0.2"
	provider := givenProvider(t)
	managementComponents := givenManagementComponents()
	managementComponents.Snow.Version = snowVersion
	g := NewWithT(t)
	result := provider.Version(managementComponents)
	g.Expect(result).To(Equal(snowVersion))
}

func TestGetInfrastructureBundle(t *testing.T) {
	tt := newSnowTest(t)
	managementComponents := givenManagementComponents()

	want := &types.InfrastructureBundle{
		FolderName: "infrastructure-snow/v1.0.2/",
		Manifests: []releasev1alpha1.Manifest{
			managementComponents.Snow.Components,
			managementComponents.Snow.Metadata,
		},
	}
	got := tt.provider.GetInfrastructureBundle(managementComponents)
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

func TestChangeDiffNoChange(t *testing.T) {
	g := NewWithT(t)
	provider := givenProvider(t)
	managementComponents := givenManagementComponents()
	g.Expect(provider.ChangeDiff(managementComponents, managementComponents)).To(BeNil())
}

func TestChangeDiffWithChange(t *testing.T) {
	g := NewWithT(t)
	provider := givenProvider(t)
	managementComponents := givenManagementComponents()
	newManagementComponents := givenManagementComponents()

	managementComponents.Snow.Version = "v1.0.2"
	newManagementComponents.Snow.Version = "v1.0.3"
	want := &types.ComponentChangeDiff{
		ComponentName: "snow",
		NewVersion:    "v1.0.3",
		OldVersion:    "v1.0.2",
	}
	g.Expect(provider.ChangeDiff(managementComponents, newManagementComponents)).To(Equal(want))
}

func TestUpdateSecrets(t *testing.T) {
	tt := newSnowTest(t)
	tt.kubeUnAuthClient.EXPECT().Apply(
		tt.ctx,
		tt.cluster.KubeconfigFile,
		tt.clusterSpec.SnowCredentialsSecret,
	).Return(nil)

	tt.Expect(tt.provider.UpdateSecrets(tt.ctx, tt.cluster, tt.clusterSpec)).To(Succeed())
}

func TestPreCAPIInstall(t *testing.T) {
	tt := newSnowTest(t)
	tt.kubeUnAuthClient.EXPECT().Apply(
		tt.ctx,
		tt.cluster.KubeconfigFile,
		tt.clusterSpec.SnowCredentialsSecret,
	).Return(nil)

	tt.Expect(tt.provider.PreCAPIInstallOnBootstrap(tt.ctx, tt.cluster, tt.clusterSpec)).To(Succeed())
}

func TestUpdateSecretsApplyError(t *testing.T) {
	tt := newSnowTest(t)
	tt.kubeUnAuthClient.EXPECT().Apply(
		tt.ctx,
		tt.cluster.KubeconfigFile,
		tt.clusterSpec.SnowCredentialsSecret,
	).Return(errors.New("error"))

	tt.Expect(tt.provider.UpdateSecrets(tt.ctx, tt.cluster, tt.clusterSpec)).NotTo(Succeed())
}
