package docker_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	dockerMocks "github.com/aws/eks-anywhere/pkg/providers/docker/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type dockerTest struct {
	*WithT
	dockerClient *dockerMocks.MockProviderClient
	kubectl      *dockerMocks.MockProviderKubectlClient
	provider     *docker.Provider
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

func TestVersion(t *testing.T) {
	tt := newTest(t)
	dockerProviderVersion := "v0.3.19"
	managementComponents := givenManagementComponents()
	managementComponents.Docker.Version = dockerProviderVersion

	tt.Expect(tt.provider.Version(managementComponents)).To(Equal(dockerProviderVersion))
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

func TestGetInfrastructureBundleSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		testName             string
		managementComponents *cluster.ManagementComponents
	}{
		{
			testName:             "create overrides layer",
			managementComponents: givenManagementComponents(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)

			infraBundle := p.GetInfrastructureBundle(tt.managementComponents)
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
		EtcdURL:     "https://distro.eks.amazonaws.com/kubernetes-1-21/releases/4/artifacts/etcd/v3.4.16/etcd-linux-amd64-v3.4.16.tar.gz",
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

func givenManagementComponents() *cluster.ManagementComponents {
	return &cluster.ManagementComponents{
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
	}
}

func TestChangeDiffNoChange(t *testing.T) {
	tt := newTest(t)
	assert.Nil(t, tt.provider.ChangeDiff(givenManagementComponents(), givenManagementComponents()))
}

func TestChangeDiffWithChange(t *testing.T) {
	tt := newTest(t)
	currentManagementComponents := givenManagementComponents()
	currentManagementComponents.Docker.Version = "v0.3.18"

	newManagementComponents := givenManagementComponents()
	newManagementComponents.Docker.Version = "v0.3.19"

	wantDiff := &types.ComponentChangeDiff{
		ComponentName: "docker",
		NewVersion:    "v0.3.19",
		OldVersion:    "v0.3.18",
	}

	tt.Expect(tt.provider.ChangeDiff(currentManagementComponents, newManagementComponents)).To(Equal(wantDiff))
}

func TestDockerTemplateBuilderGenerateCAPISpecControlPlane(t *testing.T) {
	type args struct {
		clusterSpec  *cluster.Spec
		buildOptions []providers.BuildMapOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "kube 119 test",
			args: args{
				clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
					s.Cluster.Name = "test-cluster"
				}),
				buildOptions: nil,
			},
			wantErr: nil,
		},
		{
			name: "kubelet config specified",
			args: args{
				clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
					s.Cluster.Name = "test-cluster"
					s.Cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
						KubeletConfiguration: &unstructured.Unstructured{
							Object: map[string]interface{}{
								"maxPods":    20,
								"apiVersion": "kubelet.config.k8s.io/v1beta1",
								"kind":       "KubeletConfiguration",
							},
						},
						Count: 1,
						Endpoint: &v1alpha1.Endpoint{
							Host: "1.1.1.1",
						},
					}
				}),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			builder := docker.NewDockerTemplateBuilder(time.Now)

			gotContent, err := builder.GenerateCAPISpecControlPlane(tt.args.clusterSpec, tt.args.buildOptions...)
			if err != tt.wantErr && !assert.Contains(t, err.Error(), tt.wantErr.Error()) {
				t.Errorf("Got DockerTemplateBuilder.GenerateCAPISpecControlPlane() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				g.Expect(gotContent).NotTo(BeEmpty())
			}
		})
	}
}

func TestDockerTemplateBuilderGenerateCAPISpecWorkers(t *testing.T) {
	type args struct {
		clusterSpec *cluster.Spec
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "kubelet config specified",
			args: args{
				clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
					s.Cluster.Name = "test-cluster"
					s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{
						{
							KubeletConfiguration: &unstructured.Unstructured{
								Object: map[string]interface{}{
									"maxPods":    20,
									"apiVersion": "kubelet.config.k8s.io/v1beta1",
									"kind":       "KubeletConfiguration",
								},
							},
							Count: ptr.Int(1),
							Name:  "test",
						},
					}
				}),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			builder := docker.NewDockerTemplateBuilder(time.Now)

			gotContent, err := builder.GenerateCAPISpecWorkers(tt.args.clusterSpec, nil, nil)
			if err != tt.wantErr && !assert.Contains(t, err.Error(), tt.wantErr.Error()) {
				t.Errorf("Got DockerTemplateBuilder.GenerateCAPISpecWorkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				g.Expect(gotContent).NotTo(BeEmpty())
			}
		})
	}
}

func TestInvalidDockerTemplateWithControlplaneEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.KubernetesVersion = "1.19"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "test-ip"}
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
	})
	wantErr := fmt.Errorf("specifying endpoint host configuration in Cluster is not supported")
	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	if err == nil || err.Error() != wantErr.Error() {
		t.Fatalf("err %v, wantErr %v", err, wantErr)
	}
}

func TestTemplateBuilder_CertSANs(t *testing.T) {
	for _, tc := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "testdata/cluster_api_server_cert_san_domain_name.yaml",
			Output: "testdata/expected_cluster_api_server_cert_san_domain_name.yaml",
		},
		{
			Input:  "testdata/cluster_api_server_cert_san_ip.yaml",
			Output: "testdata/expected_cluster_api_server_cert_san_ip.yaml",
		},
	} {
		g := NewWithT(t)
		clusterSpec := test.NewFullClusterSpec(t, tc.Input)

		bldr := docker.NewDockerTemplateBuilder(time.Now)

		data, err := bldr.GenerateCAPISpecControlPlane(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())

		test.AssertContentToFile(t, string(data), tc.Output)
	}
}

func TestDockerWriteKubeconfig(t *testing.T) {
	for _, tc := range []struct {
		clusterName    string
		kubeconfigPath string
		providerErr    error
		writerErr      error
	}{
		{
			clusterName:    "test",
			kubeconfigPath: "test",
		},
		{
			clusterName:    "test",
			kubeconfigPath: "test",
			writerErr:      fmt.Errorf("failed to write kubeconfig"),
		},
		{
			clusterName:    "test",
			kubeconfigPath: "test",
			providerErr:    fmt.Errorf("failed to get LB port"),
		},
	} {
		g := NewWithT(t)
		ctx := context.Background()
		buf := bytes.NewBuffer([]byte{})
		mockCtrl := gomock.NewController(t)
		dockerClient := dockerMocks.NewMockProviderClient(mockCtrl)
		mocksWriter := dockerMocks.NewMockKubeconfigReader(mockCtrl)
		mocksWriter.EXPECT().GetClusterKubeconfig(ctx, tc.clusterName, tc.kubeconfigPath).Return([]byte{}, tc.writerErr)
		if tc.writerErr == nil {
			dockerClient.EXPECT().GetDockerLBPort(ctx, tc.clusterName).Return("", tc.providerErr)
		}
		writer := docker.NewKubeconfigWriter(dockerClient, mocksWriter)
		err := writer.WriteKubeconfig(ctx, tc.clusterName, tc.kubeconfigPath, buf)
		if tc.writerErr == nil && tc.providerErr == nil {
			g.Expect(err).ToNot(HaveOccurred())
		} else {
			g.Expect(err).To(HaveOccurred())
		}
	}
}
