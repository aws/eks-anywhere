package bootstrap_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type testKindOption func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption

func TestKindCreateBootstrapClusterSuccess(t *testing.T) {
	_, writer := test.NewWriter(t)

	clusterName := "test_cluster"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.VersionsBundle = versionBundle
	})
	eksClusterName := "test_cluster-eks-a-cluster"
	kubeConfigFile := "test_cluster.kind.kubeconfig"
	kindImage := "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.2"
	registryMirror := "registry-mirror.test"
	registryMirrorWithPort := net.JoinHostPort(registryMirror, constants.DefaultHttpsPort)
	kindImageMirror := fmt.Sprintf("%s/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.2", registryMirrorWithPort)
	clusterSpecWithMirror := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.VersionsBundle = versionBundle
		s.Cluster.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Endpoint: registryMirror,
			Port:     constants.DefaultHttpsPort,
		}
	})

	// Initialize gomock
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		name               string
		wantKubeconfig     string
		env                map[string]string
		options            []testKindOption
		wantKindConfig     string
		registryMirrorTest bool
	}{
		{
			name:               "No options",
			wantKubeconfig:     kubeConfigFile,
			options:            nil,
			env:                map[string]string{},
			wantKindConfig:     "testdata/kind_config.yaml",
			registryMirrorTest: false,
		},
		{
			name:           "With env option",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithEnv(map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"})
				},
			},
			env:                map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"},
			wantKindConfig:     "testdata/kind_config.yaml",
			registryMirrorTest: false,
		},
		{
			name:           "With docker option",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithExtraDockerMounts()
				},
			},
			env:                map[string]string{},
			wantKindConfig:     "testdata/kind_config_docker_mount_networking.yaml",
			registryMirrorTest: false,
		},
		{
			name:           "With extra port mappings option",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithExtraPortMappings([]int{80, 443})
				},
			},
			env:                map[string]string{},
			wantKindConfig:     "testdata/kind_config_extra_port_mappings.yaml",
			registryMirrorTest: false,
		},
		{
			name:           "With docker option and env option",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithEnv(map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"})
				},
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithExtraDockerMounts()
				},
			},
			env:                map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"},
			wantKindConfig:     "testdata/kind_config_docker_mount_networking.yaml",
			registryMirrorTest: false,
		},
		{
			name:           "With registry mirror option, no CA cert provided",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithRegistryMirror(registryMirrorWithPort, "")
				},
			},
			env:                map[string]string{},
			wantKindConfig:     "testdata/kind_config_registry_mirror_insecure.yaml",
			registryMirrorTest: true,
		},
		{
			name:           "With registry mirror option, with CA cert",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithRegistryMirror(registryMirrorWithPort, "ca.crt")
				},
			},
			env:                map[string]string{},
			wantKindConfig:     "testdata/kind_config_registry_mirror_with_ca.yaml",
			registryMirrorTest: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)

			var (
				spec  *cluster.Spec
				image string
			)
			if tt.registryMirrorTest {
				spec = clusterSpecWithMirror
				image = kindImageMirror
			} else {
				spec = clusterSpec
				image = kindImage
			}

			executable.EXPECT().ExecuteWithEnv(
				ctx,
				tt.env,
				"create", "cluster", "--name", eksClusterName, "--kubeconfig", test.OfType("string"), "--image", image, "--config", test.OfType("string"),
			).Return(bytes.Buffer{}, nil).Times(1).Do(
				func(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) {
					gotKindConfig := args[9]
					test.AssertFilesEquals(t, gotKindConfig, tt.wantKindConfig)

					return bytes.Buffer{}, nil
				},
			)

			k := executables.NewKind(executable, writer)
			gotKubeconfig, err := k.CreateBootstrapCluster(ctx, spec, testOptionsToBootstrapOptions(k, tt.options)...)
			if err != nil {
				t.Fatalf("CreateBootstrapCluster() error = %v, wantErr %v", err, nil)
			}

			if !strings.HasSuffix(gotKubeconfig, tt.wantKubeconfig) {
				t.Errorf("CreateBootstrapCluster() gotKubeconfig = %v, want to end with %v", gotKubeconfig, tt.wantKubeconfig)
			}
		})
	}
}

func TestKindCreateBootstrapClusterExecutableError(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "clusterName"
		s.VersionsBundle = versionBundle
	})

	ctx := context.Background()
	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, map[string]string{}, gomock.Any()).Return(bytes.Buffer{}, errors.New("error from execute with env"))
	k := executables.NewKind(executable, writer)
	gotKubeconfig, err := k.CreateBootstrapCluster(ctx, clusterSpec)
	if err == nil {
		t.Fatal("Kind.CreateBootstrapCluster() error = nil")
	}

	if gotKubeconfig != "" {
		t.Errorf("CreateBootstrapCluster() gotKubeconfig = %v, want empty string", gotKubeconfig)
	}
}

func testOptionsToBootstrapOptions(k *executables.Kind, testOpts []testKindOption) []bootstrapper.BootstrapClusterClientOption {
	opts := make([]bootstrapper.BootstrapClusterClientOption, 0, len(testOpts))
	for _, opt := range testOpts {
		opts = append(opts, opt(k))
	}

	return opts
}

func TestKindDeleteBootstrapClusterSuccess(t *testing.T) {
	cluster := &types.Cluster{
		Name: "clusterName",
	}
	ctx := context.Background()
	_, writer := test.NewWriter(t)
	internalName := fmt.Sprintf("%s-eks-a-cluster", cluster.Name)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	expectedParam := []string{"delete", "cluster", "--name", internalName}
	executable.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	k := executables.NewKind(executable, writer)
	if err := k.DeleteBootstrapCluster(ctx, cluster); err != nil {
		t.Fatalf("Kind.DeleteBootstrapCluster() error = %v, want nil", err)
	}
}

func TestKindDeleteBootstrapClusterExecutableError(t *testing.T) {
	cluster := &types.Cluster{
		Name: "clusterName",
	}
	ctx := context.Background()
	_, writer := test.NewWriter(t)
	internalName := fmt.Sprintf("%s-eks-a-cluster", cluster.Name)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	expectedParam := []string{"delete", "cluster", "--name", internalName}
	executable.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	k := executables.NewKind(executable, writer)
	if err := k.DeleteBootstrapCluster(ctx, cluster); err == nil {
		t.Fatalf("Kind.DeleteBootstrapCluster() error = %v, want not nil", err)
	}
}

func TestKindClusterExists(t *testing.T) {
	tests := []struct {
		testName     string
		clusterName  string
		internalName string
		kindResponse string
	}{
		{
			testName:     "one cluster",
			clusterName:  "cluster-name",
			internalName: "cluster-name-eks-a-cluster",
			kindResponse: "cluster-name-eks-a-cluster",
		},
		{
			testName:     "3 clusters",
			clusterName:  "cluster-name-2",
			internalName: "cluster-name-2-eks-a-cluster",
			kindResponse: "cluster-name-eks-a-cluster\ncluster-name-eks-a-cluster-6\ncluster-name-2-eks-a-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			_, writer := test.NewWriter(t)

			mockCtrl := gomock.NewController(t)
			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().Execute(ctx, "get", "clusters").Return(*bytes.NewBufferString(tt.kindResponse), nil)
			k := executables.NewKind(executable, writer)
			clusterExists, err := k.ClusterExists(ctx, tt.clusterName)
			if err != nil {
				t.Fatalf("Kind.ClusterExists() error = %v, wantErr nil", err)
			}

			if !clusterExists {
				t.Fatal("ClusterExists() clusterExists = false, want true")
			}
		})
	}
}

func TestKindGetKubeconfig(t *testing.T) {
	clusterName := "cluster-name"
	ctx := context.Background()
	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "get", "kubeconfig", "--name", fmt.Sprintf("%s-eks-a-cluster", clusterName)).Return(bytes.Buffer{}, nil)
	k := executables.NewKind(executable, writer)
	_, err := k.GetKubeconfig(ctx, clusterName)
	if err != nil {
		t.Fatalf("Kind.GetKubeconfig() error = %v, wantErr nil", err)
	}
}

var versionBundle = &cluster.VersionsBundle{
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
	VersionsBundle: &releasev1.VersionsBundle{
		KubeVersion: "1.19",
		EksD: releasev1.EksDRelease{
			KindNode: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.2",
			},
		},
		CertManager: releasev1.CertManagerBundle{
			Acmesolver: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/cert-manager/cert-manager-acmesolver:v1.1.0",
			},
			Cainjector: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/cert-manager/cert-manager-cainjector:v1.1.0",
			},
			Controller: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/cert-manager/cert-manager-controller:v1.1.0",
			},
			Webhook: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/cert-manager/cert-manager-webhook:v1.1.0",
			},
			Manifest: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Version: "v1.5.3",
		},
		ClusterAPI: releasev1.CoreClusterAPI{
			Version: "v0.3.19",
			Controller: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/cluster-api-controller:v0.3.19-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
			Components: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
		},
		Bootstrap: releasev1.KubeadmBootstrapBundle{
			Version: "v0.3.19",
			Controller: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/kubeadm-bootstrap-controller:v0.3.19-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
			Components: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
		},
		ControlPlane: releasev1.KubeadmControlPlaneBundle{
			Version: "v0.3.19",
			Controller: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/kubeadm-control-plane-controller:v0.3.19-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
			Components: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
		},
		Snow: releasev1.SnowBundle{
			Version: "v0.0.0",
		},
		VSphere: releasev1.VSphereBundle{
			Version: "v0.7.8",
			ClusterAPIController: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-vsphere/release/manager:v0.7.8-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
		},
		Nutanix: releasev1.NutanixBundle{
			Version: "v0.5.1",
			ClusterAPIController: releasev1.Image{
				URI: "public.ecr.aws/release-container-registry/nutanix-cloud-native/cluster-api-provider-nutanix/release/manager:v0.5.1-eks-a-v0.0.0-dev-build.1",
			},
		},
		Tinkerbell: releasev1.TinkerbellBundle{
			Version: "v0.1.0",
			ClusterAPIController: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/tinkerbell/cluster-api-provider-tinkerbell:v0.1.0-eks-a-0.0.1.build.38",
			},
		},
		CloudStack: releasev1.CloudStackBundle{
			Version: "v0.7.8",
			ClusterAPIController: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-cloudstack/release/manager:v0.7.8-eks-a-0.0.1.build.38",
			},
			KubeRbacProxy: kubeProxyVersion08,
		},
		Docker: releasev1.DockerBundle{
			Version: "v0.3.19",
			Manager: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/capd-manager:v0.3.19-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
		},
		Eksa: releasev1.EksaBundle{
			CliTools: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/eks-anywhere-cli-tools:v1-19-1-75ac0bf61974d7ea5d83c17a1c629f26c142cca7",
			},
		},
		ExternalEtcdBootstrap: releasev1.EtcdadmBootstrapBundle{
			Version: "v0.1.0",
			Components: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Controller: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/mrajashree/etcdadm-bootstrap-provider:v0.1.0",
			},
			KubeProxy: kubeProxyVersion08,
		},
		ExternalEtcdController: releasev1.EtcdadmControllerBundle{
			Version: "v0.1.0",
			Components: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: releasev1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Controller: releasev1.Image{
				URI: "public.ecr.aws/l0g8r8j6/mrajashree/etcdadm-controller:v0.1.0",
			},
			KubeProxy: kubeProxyVersion08,
		},
	},
}

var kubeProxyVersion08 = releasev1.Image{
	URI: "public.ecr.aws/l0g8r8j6/brancz/kube-rbac-proxy:v0.8.0-25df7d96779e2a305a22c6e3f9425c3465a77244",
}
