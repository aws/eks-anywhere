package executables_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	specv1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	mockproviders "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestClusterctlInitInfrastructure(t *testing.T) {
	_, writer := test.NewWriter(t)

	core := "cluster-api:v0.3.19"
	bootstrap := "kubeadm:v0.3.19"
	controlPlane := "kubeadm:v0.3.19"

	tests := []struct {
		cluster         *types.Cluster
		env             map[string]string
		testName        string
		providerName    string
		providerVersion string
		infrastructure  string
		wantConfig      string
		wantExecArgs    []interface{}
	}{
		{
			testName: "without kubconfig",
			cluster: &types.Cluster{
				Name:           "cluster-name",
				KubeconfigFile: "",
			},
			providerName:    "aws",
			providerVersion: versionBundle.Aws.Version,
			env:             map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"},
			wantExecArgs: []interface{}{
				"init", "--core", core, "--bootstrap", bootstrap, "--control-plane", controlPlane, "--infrastructure", "aws:v0.6.4", "--config", test.OfType("string"),
				"--watching-namespace", constants.EksaSystemNamespace,
			},
			wantConfig: "testdata/clusterctl_expected.yaml",
		},
		{
			testName: "with kubconfig",
			cluster: &types.Cluster{
				Name:           "cluster-name",
				KubeconfigFile: "tmp/k.kubeconfig",
			},
			providerName:    "vsphere",
			providerVersion: versionBundle.VSphere.Version,
			env:             map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"},
			wantExecArgs: []interface{}{
				"init", "--core", core, "--bootstrap", bootstrap, "--control-plane", controlPlane, "--infrastructure", "vsphere:v0.7.8", "--config", test.OfType("string"),
				"--watching-namespace", constants.EksaSystemNamespace, "--kubeconfig", "tmp/k.kubeconfig",
			},
			wantConfig: "testdata/clusterctl_expected.yaml",
		},
	}

	mockCtrl := gomock.NewController(t)

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			defer func() {
				if !t.Failed() {
					os.RemoveAll(tt.cluster.Name)
				}
			}()
			gotConfig := ""
			ctx := context.Background()

			provider := mockproviders.NewMockProvider(mockCtrl)
			provider.EXPECT().Name().Return(tt.providerName)
			provider.EXPECT().Version(clusterSpec).Return(tt.providerVersion)
			provider.EXPECT().EnvMap().Return(tt.env, nil)
			provider.EXPECT().GetInfrastructureBundle(clusterSpec).Return(&types.InfrastructureBundle{})

			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().ExecuteWithEnv(ctx, tt.env, tt.wantExecArgs...).Return(bytes.Buffer{}, nil).Times(1).Do(
				func(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) {
					gotConfig = args[10]
					tw := templater.New(writer)
					path, err := os.Getwd()
					if err != nil {
						t.Fatalf("Error getting local folder: %v", err)
					}
					data := map[string]string{
						"dir": path,
					}

					template, err := os.ReadFile(tt.wantConfig)
					if err != nil {
						t.Fatalf("Error reading local file %s: %v", tt.wantConfig, err)
					}
					filePath, err := tw.WriteToFile(string(template), data, "file.tmp")
					if err != nil {
						t.Fatalf("Error writing local file %s: %v", "file.tmp", err)
					}

					test.AssertFilesEquals(t, gotConfig, filePath)

					return bytes.Buffer{}, nil
				},
			)

			c := executables.NewClusterctl(executable, writer)

			if err := c.InitInfrastructure(ctx, clusterSpec, tt.cluster, provider); err != nil {
				t.Fatalf("Clusterctl.InitInfrastructure() error = %v, want nil", err)
			}
		})
	}
}

func TestClusterctlInitInfrastructureInstallEtcdadmControllers(t *testing.T) {
	_, writer := test.NewWriter(t)

	clusterWithUnstackedEtcd := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle = versionBundle
	})
	clusterWithUnstackedEtcd.Spec.ExternalEtcdConfiguration = &specv1.ExternalEtcdConfiguration{}

	core := "cluster-api:v0.3.19"
	bootstrapKubeadm := "kubeadm:v0.3.19"
	controlPlane := "kubeadm:v0.3.19"
	bootstrapEtcdamBootstrap := "etcdadm-bootstrap:v0.1.0"
	bootstrapEtcdamController := "etcdadm-controller:v0.1.0"

	tests := []struct {
		cluster         *types.Cluster
		env             map[string]string
		testName        string
		providerName    string
		providerVersion string
		infrastructure  string
		wantConfig      string
		wantExecArgs    []interface{}
	}{
		{
			testName: "unstacked etcd",
			cluster: &types.Cluster{
				Name:           "cluster-name",
				KubeconfigFile: "tmp/k.kubeconfig",
			},
			providerName:    "vsphere",
			providerVersion: versionBundle.VSphere.Version,
			env:             map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"},
			wantExecArgs: []interface{}{
				"init", "--core", core, "--bootstrap", bootstrapKubeadm, "--control-plane", controlPlane, "--infrastructure", "vsphere:v0.7.8", "--config", test.OfType("string"),
				"--watching-namespace", constants.EksaSystemNamespace, "--bootstrap", bootstrapEtcdamBootstrap, "--bootstrap", bootstrapEtcdamController, "--kubeconfig", "tmp/k.kubeconfig",
			},
			wantConfig: "testdata/clusterctl_expected.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer func() {
				if !t.Failed() {
					os.RemoveAll(tt.cluster.Name)
				}
			}()
			gotConfig := ""
			ctx := context.Background()

			provider := mockproviders.NewMockProvider(mockCtrl)
			provider.EXPECT().Name().Return(tt.providerName)
			provider.EXPECT().Version(clusterWithUnstackedEtcd).Return(tt.providerVersion)
			provider.EXPECT().EnvMap().Return(tt.env, nil)
			provider.EXPECT().GetInfrastructureBundle(clusterWithUnstackedEtcd).Return(&types.InfrastructureBundle{})

			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().ExecuteWithEnv(ctx, tt.env, tt.wantExecArgs...).Return(bytes.Buffer{}, nil).Times(1).Do(
				func(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) {
					gotConfig = args[10]
					tw := templater.New(writer)
					path, err := os.Getwd()
					if err != nil {
						t.Fatalf("Error getting local folder: %v", err)
					}
					data := map[string]string{
						"dir": path,
					}

					template, err := os.ReadFile(tt.wantConfig)
					if err != nil {
						t.Fatalf("Error reading local file %s: %v", tt.wantConfig, err)
					}
					filePath, err := tw.WriteToFile(string(template), data, "file.tmp")
					if err != nil {
						t.Fatalf("Error writing local file %s: %v", "file.tmp", err)
					}

					test.AssertFilesEquals(t, gotConfig, filePath)

					return bytes.Buffer{}, nil
				},
			)

			c := executables.NewClusterctl(executable, writer)

			if err := c.InitInfrastructure(ctx, clusterWithUnstackedEtcd, tt.cluster, provider); err != nil {
				t.Fatalf("Clusterctl.InitInfrastructure() error = %v, want nil", err)
			}
		})
	}
}

func TestClusterctlInitInfrastructureEnvMapError(t *testing.T) {
	ctx := context.Background()

	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	provider := mockproviders.NewMockProvider(mockCtrl)
	provider.EXPECT().Name()
	provider.EXPECT().Version(clusterSpec)
	provider.EXPECT().EnvMap().Return(nil, errors.New("error with env map"))
	provider.EXPECT().GetInfrastructureBundle(clusterSpec).Return(&types.InfrastructureBundle{})

	executable := mockexecutables.NewMockExecutable(mockCtrl)

	c := executables.NewClusterctl(executable, writer)

	if err := c.InitInfrastructure(ctx, clusterSpec, &types.Cluster{Name: "cluster-name"}, provider); err == nil {
		t.Fatal("Clusterctl.InitInfrastructure() error = nil")
	}
}

func TestClusterctlInitInfrastructureExecutableError(t *testing.T) {
	cluster := &types.Cluster{Name: "cluster-name"}
	defer func() {
		if !t.Failed() {
			os.RemoveAll(cluster.Name)
		}
	}()
	ctx := context.Background()

	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	provider := mockproviders.NewMockProvider(mockCtrl)
	provider.EXPECT().Name()
	provider.EXPECT().Version(clusterSpec)
	provider.EXPECT().EnvMap()
	provider.EXPECT().GetInfrastructureBundle(clusterSpec).Return(&types.InfrastructureBundle{})

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, nil, gomock.Any()).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	c := executables.NewClusterctl(executable, writer)

	if err := c.InitInfrastructure(ctx, clusterSpec, cluster, provider); err == nil {
		t.Fatal("Clusterctl.InitInfrastructure() error = nil")
	}
}

func TestClusterctlInitInfrastructureInvalidClusterNameError(t *testing.T) {
	ctx := context.Background()

	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	provider := mockproviders.NewMockProvider(mockCtrl)
	executable := mockexecutables.NewMockExecutable(mockCtrl)

	c := executables.NewClusterctl(executable, writer)

	if err := c.InitInfrastructure(ctx, clusterSpec, &types.Cluster{Name: ""}, provider); err == nil {
		t.Fatal("Clusterctl.InitInfrastructure() error != nil")
	}
}

func TestClusterctlMoveManagement(t *testing.T) {
	tests := []struct {
		testName     string
		from         *types.Cluster
		to           *types.Cluster
		wantMoveArgs []interface{}
	}{
		{
			testName:     "no kubeconfig",
			from:         &types.Cluster{},
			to:           &types.Cluster{},
			wantMoveArgs: []interface{}{"move", "--to-kubeconfig", "", "--namespace", constants.EksaSystemNamespace},
		},
		{
			testName: "no kubeconfig in 'from' cluster",
			from:     &types.Cluster{},
			to: &types.Cluster{
				KubeconfigFile: "to.kubeconfig",
			},
			wantMoveArgs: []interface{}{"move", "--to-kubeconfig", "to.kubeconfig", "--namespace", constants.EksaSystemNamespace},
		},
		{
			testName: "with both kubeconfigs",
			from: &types.Cluster{
				KubeconfigFile: "from.kubeconfig",
			},
			to: &types.Cluster{
				KubeconfigFile: "to.kubeconfig",
			},
			wantMoveArgs: []interface{}{"move", "--to-kubeconfig", "to.kubeconfig", "--namespace", constants.EksaSystemNamespace, "--kubeconfig", "from.kubeconfig"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			writer := mockswriter.NewMockFileWriter(mockCtrl)
			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().Execute(ctx, tt.wantMoveArgs...)

			c := executables.NewClusterctl(executable, writer)
			if err := c.MoveManagement(ctx, tt.from, tt.to); err != nil {
				t.Fatalf("Clusterctl.MoveManagement() error = %v, want nil", err)
			}
		})
	}
}

var clusterSpec = test.NewClusterSpec(func(s *cluster.Spec) {
	s.VersionsBundle = versionBundle
})

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
	VersionsBundle: &v1alpha1.VersionsBundle{
		KubeVersion: "1.19",
		EksD: v1alpha1.EksDRelease{
			KindNode: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.2",
			},
		},
		CertManager: v1alpha1.CertManagerBundle{
			Acmesolver: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/jetstack/cert-manager-acmesolver:v1.1.0",
			},
			Cainjector: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/jetstack/cert-manager-cainjector:v1.1.0",
			},
			Controller: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/jetstack/cert-manager-controller:v1.1.0",
			},
			Webhook: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/jetstack/cert-manager-webhook:v1.1.0",
			},
		},
		ClusterAPI: v1alpha1.CoreClusterAPI{
			Version: "v0.3.19",
			Controller: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/cluster-api-controller:v0.3.19-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
			Components: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
		},
		Bootstrap: v1alpha1.KubeadmBootstrapBundle{
			Version: "v0.3.19",
			Controller: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/kubeadm-bootstrap-controller:v0.3.19-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
			Components: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
		},
		ControlPlane: v1alpha1.KubeadmControlPlaneBundle{
			Version: "v0.3.19",
			Controller: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/kubeadm-control-plane-controller:v0.3.19-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
			Components: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
		},
		Aws: v1alpha1.AwsBundle{
			Version: "v0.6.4",
			Controller: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-aws/cluster-api-aws-controller:v0.6.4-25df7d96779e2a305a22c6e3f9425c3465a77244",
			},
			KubeProxy: kubeProxyVersion08,
		},
		VSphere: v1alpha1.VSphereBundle{
			Version: "v0.7.8",
			ClusterAPIController: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-vsphere/release/manager:v0.7.8-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
		},
		Docker: v1alpha1.DockerBundle{
			Version: "v0.3.19",
			Manager: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/capd-manager:v0.3.19-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
		},
		Eksa: v1alpha1.EksaBundle{
			CliTools: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/eks-anywhere-cli-tools:v1-19-1-75ac0bf61974d7ea5d83c17a1c629f26c142cca7",
			},
		},
		ExternalEtcdBootstrap: v1alpha1.EtcdadmBootstrapBundle{
			Version: "v0.1.0",
			Components: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Controller: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/mrajashree/etcdadm-bootstrap-provider:v0.1.0",
			},
			KubeProxy: kubeProxyVersion08,
		},
		ExternalEtcdController: v1alpha1.EtcdadmControllerBundle{
			Version: "v0.1.0",
			Components: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Metadata: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Controller: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/mrajashree/etcdadm-controller:v0.1.0",
			},
			KubeProxy: kubeProxyVersion08,
		},
	},
}

var kubeProxyVersion08 = v1alpha1.Image{
	URI: "public.ecr.aws/l0g8r8j6/brancz/kube-rbac-proxy:v0.8.0-25df7d96779e2a305a22c6e3f9425c3465a77244",
}
