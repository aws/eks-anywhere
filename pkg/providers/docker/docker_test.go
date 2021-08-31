package docker_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	kubeadmnv1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	dockerMocks "github.com/aws/eks-anywhere/pkg/providers/docker/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

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
			_, writer := test.NewWriter(t)
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			client.EXPECT().GetDockerLBPort(gomock.Any(), tt.args.clusterName).Return("4332", nil)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, writer, test.FakeNow)

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

	tests := []struct {
		testName     string
		clusterSpec  *cluster.Spec
		wantFileName string
	}{
		{
			testName: "valid config",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
				s.VersionsBundle = versionsBundle
			}),
			wantFileName: "testdata/valid_deployment_expected.yaml",
		},
		{
			testName: "with minimal oidc",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
				s.VersionsBundle = versionsBundle

				s.OIDCConfig = &v1alpha1.OIDCConfig{
					Spec: v1alpha1.OIDCConfigSpec{
						ClientId:  "my-client-id",
						IssuerUrl: "https://mydomain.com/issuer",
					},
				}
			}),
			wantFileName: "testdata/capd_valid_minimal_oidc_expected.yaml",
		},
		{
			testName: "with full oidc",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
				s.VersionsBundle = versionsBundle

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
			wantFileName: "testdata/capd_valid_full_oidc_expected.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, writer := test.NewWriter(t)
			ctx := context.Background()
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, writer, test.FakeNow)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			oriCluster := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: v1alpha1.Kube118,
				},
			}
			kubectl.EXPECT().GetEksaCluster(ctx, cluster).Return(oriCluster, nil)
			got, err := p.GenerateDeploymentFileForUpgrade(ctx, bootstrapCluster, cluster, tt.clusterSpec, "cluster.yaml")
			if err != nil {
				t.Fatalf("provider.GenerateDeploymentFileForUpgrade() error = %v, wantErr nil", err)
			}

			test.AssertFilesEquals(t, got, tt.wantFileName)
		})
	}
}

func TestProviderGenerateDeploymentFileSuccessNotUpdateMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	_, writer := test.NewWriter(t)
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := test.NewClusterSpec()
	p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, writer, test.FakeNow)
	cluster := &types.Cluster{
		Name: "test",
	}
	bootstrapCluster := &types.Cluster{
		Name: "bootstrap-test",
	}

	cp := &kubeadmnv1alpha3.KubeadmControlPlane{
		Spec: kubeadmnv1alpha3.KubeadmControlPlaneSpec{
			InfrastructureTemplate: v1.ObjectReference{
				Name: "test-control-plane-template-original",
			},
		},
	}
	md := &v1alpha3.MachineDeployment{
		Spec: v1alpha3.MachineDeploymentSpec{
			Template: v1alpha3.MachineTemplateSpec{
				Spec: v1alpha3.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Name: "test-worker-node-template-original",
					},
				},
			},
		},
	}

	kubectl.EXPECT().GetEksaCluster(ctx, cluster).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(cp, nil)
	kubectl.EXPECT().GetMachineDeployment(ctx, cluster, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(md, nil)

	got, err := p.GenerateDeploymentFileForUpgrade(ctx, bootstrapCluster, cluster, clusterSpec, "cluster.yaml")
	if err != nil {
		t.Fatalf("provider.GenerateDeploymentFileForUpgrade() error = %v, wantErr nil", err)
	}

	test.AssertFilesEquals(t, got, "testdata/no_machinetemplate_update_expected.yaml")
}

func TestSetupAndValidateClusterWithEndpoint(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = "test-cluster"
		s.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "test-ip"}
	})
	mockCtrl := gomock.NewController(t)
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	_, writer := test.NewWriter(t)
	p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, writer, test.FakeNow)
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
			_, writer := test.NewWriter(t)
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, writer, test.FakeNow)

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
	},
}
