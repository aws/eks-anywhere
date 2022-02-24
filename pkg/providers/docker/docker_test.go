package docker_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	dockerMocks "github.com/aws/eks-anywhere/pkg/providers/docker/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type dockerTest struct {
	*WithT
	dockerClient *dockerMocks.MockProviderClient
	kubectl      *dockerMocks.MockProviderKubectlClient
	provider     providers.Provider
}

func newTest(t *testing.T) *dockerTest {
	ctrl := gomock.NewController(t)
	client := dockerMocks.NewMockProviderClient(ctrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(ctrl)
	return &dockerTest{
		WithT:        NewWithT(t),
		dockerClient: client,
		kubectl:      kubectl,
		provider:     docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow),
	}
}

func TestProviderUpdateKubeConfig(t *testing.T) {
	input := []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJ
    server: https://172.18.0.3:6443
  name: capi-quickstart`)
	expected := `
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://127.0.0.1:4332
  name: capi-quickstart`

	mockCtrl := gomock.NewController(t)

	type fields struct {
		clusterName string
	}
	type args struct {
		content     *[]byte
		clusterName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test updates for docker config file",
			fields: fields{
				clusterName: "capi-quickstart",
			},
			args: args{
				content:     &input,
				clusterName: "capi-quickstart",
			},
			want: expected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			client.EXPECT().GetDockerLBPort(gomock.Any(), tt.args.clusterName).Return("4332", nil)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)

			if err := p.UpdateKubeConfig(tt.args.content, tt.args.clusterName); (err != nil) != tt.wantErr {
				t.Errorf("UpdateKubeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if string(*tt.args.content) != tt.want {
				t.Errorf("updateKubeConfigFile() got = %v, want %v", string(*tt.args.content), tt.want)
			}
		})
	}
}

func TestProviderGenerateDeploymentFileSuccessUpdateMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	var cpTaints, wnTaints, wnTaints2 []v1.Taint

	cpTaints = append(cpTaints, v1.Taint{Key: "key1", Value: "val1", Effect: "NoSchedule", TimeAdded: nil})
	cpTaints = append(cpTaints, v1.Taint{Key: "key2", Value: "val2", Effect: "PreferNoSchedule", TimeAdded: nil})
	cpTaints = append(cpTaints, v1.Taint{Key: "key3", Value: "val3", Effect: "NoExecute", TimeAdded: nil})
	wnTaints = append(wnTaints, v1.Taint{Key: "key2", Value: "val2", Effect: "PreferNoSchedule", TimeAdded: nil})
	wnTaints2 = append(wnTaints2, v1.Taint{Key: "wnTaitns2", Value: "true", Effect: "PreferNoSchedule", TimeAdded: nil})

	nodeLabels := map[string]string{"label1": "foo", "label2": "bar"}

	tests := []struct {
		testName    string
		clusterSpec *cluster.Spec
		wantCPFile  string
		wantMDFile  string
	}{
		{
			testName: "valid config",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_expected.yaml",
			wantMDFile: "testdata/valid_deployment_md_expected.yaml",
		},
		{
			testName: "valid config with cp taints",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.ControlPlaneConfiguration.Taints = cpTaints
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_taints_expected.yaml",
			wantMDFile: "testdata/valid_deployment_md_expected.yaml",
		},
		{
			testName: "valid config with md taints",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.ControlPlaneConfiguration.Taints = cpTaints
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, Taints: wnTaints, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_taints_expected.yaml",
			wantMDFile: "testdata/valid_deployment_md_taints_expected.yaml",
		},
		{
			testName: "valid config multiple worker node groups with machine deplyoment taints",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.ControlPlaneConfiguration.Taints = cpTaints
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, Taints: wnTaints, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}, {Count: 3, Taints: wnTaints2, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-1"}}
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
			}),
			wantCPFile: "testdata/valid_deployment_cp_taints_expected.yaml",
			wantMDFile: "testdata/valid_deployment_multiple_md_taints_expected.yaml",
		},
		{
			testName: "valid config with node labels",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Labels: nodeLabels, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_expected.yaml",
			wantMDFile: "testdata/valid_deployment_node_labels_md_expected.yaml",
		},
		{
			testName: "valid config with cp node labels",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.ControlPlaneConfiguration.Labels = nodeLabels
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_node_labels_cp_expected.yaml",
			wantMDFile: "testdata/valid_deployment_md_expected.yaml",
		},
		{
			testName: "valid config with cidrs and custom resolv.conf",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"10.10.0.0/24", "10.128.0.0/12"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"192.168.0.0/16", "10.10.0.0/16"}
				s.Spec.ClusterNetwork.DNS.ResolvConf = &v1alpha1.ResolvConf{Path: "/etc/my-custom-resolv.conf"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_custom_cidrs_cp_expected.yaml",
			wantMDFile: "testdata/valid_deployment_custom_cidrs_md_expected.yaml",
		},
		{
			testName: "with minimal oidc",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}

				s.OIDCConfig = &v1alpha1.OIDCConfig{
					Spec: v1alpha1.OIDCConfigSpec{
						ClientId:  "my-client-id",
						IssuerUrl: "https://mydomain.com/issuer",
					},
				}
			}),
			wantCPFile: "testdata/capd_valid_minimal_oidc_cp_expected.yaml",
			wantMDFile: "testdata/capd_valid_minimal_oidc_md_expected.yaml",
		},
		{
			testName: "with full oidc",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundle = versionsBundle
				s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
				s.OIDCConfig = &v1alpha1.OIDCConfig{
					Spec: v1alpha1.OIDCConfigSpec{
						ClientId:     "my-client-id",
						IssuerUrl:    "https://mydomain.com/issuer",
						GroupsClaim:  "claim1",
						GroupsPrefix: "prefix-for-groups",
						RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{
							{
								Claim: "sub",
								Value: "test",
							},
						},
						UsernameClaim:  "username-claim",
						UsernamePrefix: "username-prefix",
					},
				}
			}),
			wantCPFile: "testdata/capd_valid_full_oidc_cp_expected.yaml",
			wantMDFile: "testdata/capd_valid_full_oidc_md_expected.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
			cluster := &types.Cluster{
				Name: "test",
			}
			currentSpec := tt.clusterSpec.DeepCopy()
			tt.clusterSpec.Bundles.Spec.Number = 2
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", tt.clusterSpec.Name),
				map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.Any(), gomock.Any())
			cpContent, mdContent, err := p.GenerateCAPISpecForUpgrade(ctx, bootstrapCluster, cluster, currentSpec, tt.clusterSpec)
			if err != nil {
				t.Fatalf("provider.GenerateCAPISpecForUpgrade() error = %v, wantErr nil", err)
			}
			test.AssertContentToFile(t, string(cpContent), tt.wantCPFile)
			test.AssertContentToFile(t, string(mdContent), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateDeploymentFileSuccessNotUpdateMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := test.NewClusterSpec()
	clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
	clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
	clusterSpec.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 0, MachineGroupRef: &v1alpha1.Ref{Name: "fluxAddonTestCluster"}, Name: "md-0"}}
	p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	cluster := &types.Cluster{
		Name: "test",
	}
	currentSpec := clusterSpec.DeepCopy()
	bootstrapCluster := &types.Cluster{
		Name: "bootstrap-test",
	}

	cp := &controlplanev1.KubeadmControlPlane{
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					Name: "test-control-plane-template-original",
				},
			},
		},
	}
	md := &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Name: "test-md-0-original",
					},
				},
			},
		},
	}
	machineDeploymentName := fmt.Sprintf("%s-%s", clusterSpec.Name, clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Name)

	kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(cp, nil)
	kubectl.EXPECT().GetMachineDeployment(ctx, machineDeploymentName, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(md, nil)

	cpContent, mdContent, err := p.GenerateCAPISpecForUpgrade(ctx, bootstrapCluster, cluster, currentSpec, clusterSpec)
	if err != nil {
		t.Fatalf("provider.GenerateCAPISpecForUpgrade() error = %v, wantErr nil", err)
	}

	test.AssertContentToFile(t, string(cpContent), "testdata/no_machinetemplate_update_cp_expected.yaml")
	test.AssertContentToFile(t, string(mdContent), "testdata/no_machinetemplate_update_md_expected.yaml")
}

func TestSetupAndValidateClusterWithEndpoint(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = "test-cluster"
		s.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "test-ip"}
	})
	mockCtrl := gomock.NewController(t)
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	ctx := context.Background()
	err := p.SetupAndValidateCreateCluster(ctx, clusterSpec)
	wantErr := fmt.Errorf("specifying endpoint host configuration in Cluster is not supported")

	if !reflect.DeepEqual(wantErr, err) {
		t.Errorf("got = <%v>, want = <%v>", err, wantErr)
	}
}

func TestGetInfrastructureBundleSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		testName    string
		clusterSpec *cluster.Spec
	}{
		{
			testName: "create overrides layer",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.VersionsBundle = versionsBundle
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)

			infraBundle := p.GetInfrastructureBundle(tt.clusterSpec)
			if infraBundle == nil {
				t.Fatalf("provider.GetInfrastructureBundle() should have an infrastructure bundle")
			}
			assert.Equal(t, "infrastructure-docker/v0.3.19/", infraBundle.FolderName, "Incorrect folder name")
			assert.Equal(t, len(infraBundle.Manifests), 3, "Wrong number of files in the infrastructure bundle")

			wantManifests := []releasev1alpha1.Manifest{
				versionsBundle.Docker.Components,
				versionsBundle.Docker.Metadata,
				versionsBundle.Docker.ClusterTemplate,
			}
			assert.ElementsMatch(t, infraBundle.Manifests, wantManifests, "Incorrect manifests")
		})
	}
}

var versionsBundle = &cluster.VersionsBundle{
	KubeDistro: &cluster.KubeDistro{
		Kubernetes: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/kubernetes",
			Tag:        "v1.19.6-eks-1-19-2",
		},
		CoreDNS: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/coredns",
			Tag:        "v1.8.0-eks-1-19-2",
		},
		Etcd: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/etcd-io",
			Tag:        "v3.4.14-eks-1-19-2",
		},
		EtcdVersion: "3.4.14",
	},
	VersionsBundle: &releasev1alpha1.VersionsBundle{
		EksD: releasev1alpha1.EksDRelease{
			KindNode: releasev1alpha1.Image{
				Description: "kind/node container image",
				Name:        "kind/node",
				URI:         "public.ecr.aws/eks-distro/kubernetes-sigs/kind/node:v1.18.16-eks-1-18-4-216edda697a37f8bf16651af6c23b7e2bb7ef42f-62681885fe3a97ee4f2b110cc277e084e71230fa",
			},
		},
		Docker: releasev1alpha1.DockerBundle{
			Version: "v0.3.19",
			Manager: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/capd-manager:v0.3.15-6bdb9fc78bb926135843c58ec8b77b54d8f2c82c",
			},
			KubeProxy: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/brancz/kube-rbac-proxy:v0.8.0-25df7d96779e2a305a22c6e3f9425c3465a77244",
			},
			Components: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/infrastructure-components-development.yaml",
			},
			ClusterTemplate: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/cluster-template-development.yaml",
			},
			Metadata: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/metadata.yaml",
			},
		},
		Haproxy: releasev1alpha1.HaproxyBundle{
			Image: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/haproxy:v0.11.1-eks-a-v0.0.0-dev-build.1464",
			},
		},
	},
}

func TestChangeDiffNoChange(t *testing.T) {
	tt := newTest(t)
	clusterSpec := test.NewClusterSpec()
	assert.Nil(t, tt.provider.ChangeDiff(clusterSpec, clusterSpec))
}

func TestChangeDiffWithChange(t *testing.T) {
	tt := newTest(t)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.Docker.Version = "v0.3.18"
	})
	newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.Docker.Version = "v0.3.19"
	})

	wantDiff := &types.ComponentChangeDiff{
		ComponentName: "docker",
		NewVersion:    "v0.3.19",
		OldVersion:    "v0.3.18",
	}

	tt.Expect(tt.provider.ChangeDiff(clusterSpec, newClusterSpec)).To(Equal(wantDiff))
}

func TestProviderGenerateCAPISpecForCreateWithPodIAMConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{
		Name: "test-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = "test-cluster"
		s.Spec.KubernetesVersion = "1.19"
		s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Spec.ControlPlaneConfiguration.Count = 1
		s.VersionsBundle = versionsBundle
		s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}}}
	})
	clusterSpec.Spec.PodIAMConfig = &v1alpha1.PodIAMConfig{ServiceAccountIssuer: "https://test"}

	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/valid_deployment_cp_pod_iam_expected.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithStackedEtcd(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{
		Name: "test-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = "test-cluster"
		s.Spec.KubernetesVersion = "1.19"
		s.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Spec.ControlPlaneConfiguration.Count = 1
		s.VersionsBundle = versionsBundle
		s.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}}}
	})

	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/valid_deployment_cp_stacked_etcd_expected.yaml")
}
