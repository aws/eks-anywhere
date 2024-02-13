package executables_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	mockproviders "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type clusterctlTest struct {
	*WithT
	ctx                  context.Context
	managementComponents *cluster.ManagementComponents
	cluster              *types.Cluster
	clusterctl           *executables.Clusterctl
	e                    *mockexecutables.MockExecutable
	provider             *mockproviders.MockProvider
	writer               filewriter.FileWriter
	providerEnvMap       map[string]string
}

func newClusterctlTest(t *testing.T) *clusterctlTest {
	ctrl := gomock.NewController(t)
	_, writer := test.NewWriter(t)
	reader := files.NewReader()
	e := mockexecutables.NewMockExecutable(ctrl)

	return &clusterctlTest{
		WithT:                NewWithT(t),
		ctx:                  context.Background(),
		managementComponents: cluster.ManagementComponentsFromBundles(clusterSpec.Bundles),
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "config/c.kubeconfig",
		},
		e:              e,
		provider:       mockproviders.NewMockProvider(ctrl),
		clusterctl:     executables.NewClusterctl(e, writer, reader),
		writer:         writer,
		providerEnvMap: map[string]string{"var": "value"},
	}
}

func (ct *clusterctlTest) expectBuildOverrideLayer() {
	ct.provider.EXPECT().GetInfrastructureBundle(ct.managementComponents).Return(&types.InfrastructureBundle{})
}

func (ct *clusterctlTest) expectGetProviderEnvMap() {
	ct.provider.EXPECT().EnvMap(ct.managementComponents, clusterSpec).Return(ct.providerEnvMap, nil)
}

func TestClusterctlInitInfrastructure(t *testing.T) {
	_, writer := test.NewWriter(t)

	core := "cluster-api:v0.3.19"
	bootstrap := "kubeadm:v0.3.19"
	controlPlane := "kubeadm:v0.3.19"
	etcdadmBootstrap := "etcdadm-bootstrap:v0.1.0"
	etcdadmController := "etcdadm-controller:v0.1.0"

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
			providerName:    "vsphere",
			providerVersion: versionBundle.VSphere.Version,
			env:             map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"},
			wantExecArgs: []interface{}{
				"init", "--core", core, "--bootstrap", bootstrap, "--control-plane", controlPlane, "--infrastructure", "vsphere:v0.7.8", "--config", test.OfType("string"),
				"--bootstrap", etcdadmBootstrap, "--bootstrap", etcdadmController,
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
				"--bootstrap", etcdadmBootstrap, "--bootstrap", etcdadmController,
				"--kubeconfig", "tmp/k.kubeconfig",
			},
			wantConfig: "testdata/clusterctl_expected.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			defer func() {
				if !t.Failed() {
					os.RemoveAll(tt.cluster.Name)
				}
			}()
			tc := newClusterctlTest(t)

			gotConfig := ""

			tc.provider.EXPECT().Name().Return(tt.providerName)
			tc.provider.EXPECT().Version(tc.managementComponents).Return(tt.providerVersion)
			tc.provider.EXPECT().EnvMap(tc.managementComponents, clusterSpec).Return(tt.env, nil)
			tc.provider.EXPECT().GetInfrastructureBundle(tc.managementComponents).Return(&types.InfrastructureBundle{})

			tc.e.EXPECT().ExecuteWithEnv(tc.ctx, tt.env, tt.wantExecArgs...).Return(bytes.Buffer{}, nil).Times(1).Do(
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

			if err := tc.clusterctl.InitInfrastructure(tc.ctx, tc.managementComponents, clusterSpec, tt.cluster, tc.provider); err != nil {
				t.Fatalf("Clusterctl.InitInfrastructure() error = %v, want nil", err)
			}
		})
	}
}

func TestClusterctlInitInfrastructureEnvMapError(t *testing.T) {
	cluster := &types.Cluster{Name: "cluster-name"}
	defer func() {
		if !t.Failed() {
			os.RemoveAll(cluster.Name)
		}
	}()
	tt := newClusterctlTest(t)

	tt.provider.EXPECT().Name()
	tt.provider.EXPECT().Version(tt.managementComponents)
	tt.provider.EXPECT().EnvMap(tt.managementComponents, clusterSpec).Return(nil, errors.New("error with env map"))
	tt.provider.EXPECT().GetInfrastructureBundle(tt.managementComponents).Return(&types.InfrastructureBundle{})

	if err := tt.clusterctl.InitInfrastructure(tt.ctx, tt.managementComponents, clusterSpec, cluster, tt.provider); err == nil {
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
	tt := newClusterctlTest(t)

	tt.provider.EXPECT().Name()
	tt.provider.EXPECT().Version(tt.managementComponents)
	tt.provider.EXPECT().EnvMap(tt.managementComponents, clusterSpec)
	tt.provider.EXPECT().GetInfrastructureBundle(tt.managementComponents).Return(&types.InfrastructureBundle{})

	tt.e.EXPECT().ExecuteWithEnv(tt.ctx, nil, gomock.Any()).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	if err := tt.clusterctl.InitInfrastructure(tt.ctx, tt.managementComponents, clusterSpec, cluster, tt.provider); err == nil {
		t.Fatal("Clusterctl.InitInfrastructure() error = nil")
	}
}

func TestClusterctlInitInfrastructureInvalidClusterNameError(t *testing.T) {
	tt := newClusterctlTest(t)

	if err := tt.clusterctl.InitInfrastructure(tt.ctx, tt.managementComponents, clusterSpec, &types.Cluster{Name: ""}, tt.provider); err == nil {
		t.Fatal("Clusterctl.InitInfrastructure() error != nil")
	}
}

func TestClusterctlBackupManagement(t *testing.T) {
	managementClusterState := fmt.Sprintf("cluster-state-backup-%s", time.Now().Format("2006-01-02T15_04_05"))
	clusterName := "cluster"

	tests := []struct {
		testName     string
		cluster      *types.Cluster
		wantMoveArgs []interface{}
	}{
		{
			testName: "backup success",
			cluster: &types.Cluster{
				Name:           clusterName,
				KubeconfigFile: "cluster.kubeconfig",
			},
			wantMoveArgs: []interface{}{"move", "--to-directory", fmt.Sprintf("%s/%s", clusterName, managementClusterState), "--kubeconfig", "cluster.kubeconfig", "--namespace", constants.EksaSystemNamespace, "--filter-cluster", clusterName},
		},
		{
			testName: "no kubeconfig file",
			cluster: &types.Cluster{
				Name: clusterName,
			},
			wantMoveArgs: []interface{}{"move", "--to-directory", fmt.Sprintf("%s/%s", clusterName, managementClusterState), "--kubeconfig", "", "--namespace", constants.EksaSystemNamespace, "--filter-cluster", clusterName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tc := newClusterctlTest(t)
			tc.e.EXPECT().Execute(tc.ctx, tt.wantMoveArgs...)

			if err := tc.clusterctl.BackupManagement(tc.ctx, tt.cluster, managementClusterState, clusterName); err != nil {
				t.Fatalf("Clusterctl.BackupManagement() error = %v, want nil", err)
			}
		})
	}
}

func TestClusterctlBackupManagementFailed(t *testing.T) {
	managementClusterState := fmt.Sprintf("cluster-state-backup-%s", time.Now().Format("2006-01-02T15_04_05"))
	tt := newClusterctlTest(t)

	cluster := &types.Cluster{
		Name:           "cluster",
		KubeconfigFile: "cluster.kubeconfig",
	}

	wantMoveArgs := []interface{}{"move", "--to-directory", fmt.Sprintf("%s/%s", cluster.Name, managementClusterState), "--kubeconfig", "cluster.kubeconfig", "--namespace", constants.EksaSystemNamespace, "--filter-cluster", cluster.Name}

	tt.e.EXPECT().Execute(tt.ctx, wantMoveArgs...).Return(bytes.Buffer{}, fmt.Errorf("error backing up management cluster resources"))
	if err := tt.clusterctl.BackupManagement(tt.ctx, cluster, managementClusterState, cluster.Name); err == nil {
		t.Fatalf("Clusterctl.BackupManagement() error = %v, want nil", err)
	}
}

func TestClusterctlMoveManagement(t *testing.T) {
	tests := []struct {
		testName     string
		from         *types.Cluster
		to           *types.Cluster
		clusterName  string
		wantMoveArgs []interface{}
	}{
		{
			testName:     "no kubeconfig",
			from:         &types.Cluster{},
			to:           &types.Cluster{},
			clusterName:  "",
			wantMoveArgs: []interface{}{"move", "--to-kubeconfig", "", "--namespace", constants.EksaSystemNamespace, "--filter-cluster", ""},
		},
		{
			testName: "no kubeconfig in 'from' cluster",
			from:     &types.Cluster{},
			to: &types.Cluster{
				KubeconfigFile: "to.kubeconfig",
			},
			clusterName:  "",
			wantMoveArgs: []interface{}{"move", "--to-kubeconfig", "to.kubeconfig", "--namespace", constants.EksaSystemNamespace, "--filter-cluster", ""},
		},
		{
			testName: "with both kubeconfigs",
			from: &types.Cluster{
				KubeconfigFile: "from.kubeconfig",
			},
			to: &types.Cluster{
				KubeconfigFile: "to.kubeconfig",
			},
			clusterName:  "",
			wantMoveArgs: []interface{}{"move", "--to-kubeconfig", "to.kubeconfig", "--namespace", constants.EksaSystemNamespace, "--filter-cluster", "", "--kubeconfig", "from.kubeconfig"},
		},
		{
			testName: "with filter cluster",
			from: &types.Cluster{
				KubeconfigFile: "from.kubeconfig",
			},
			to: &types.Cluster{
				KubeconfigFile: "to.kubeconfig",
			},
			clusterName:  "test-cluster",
			wantMoveArgs: []interface{}{"move", "--to-kubeconfig", "to.kubeconfig", "--namespace", constants.EksaSystemNamespace, "--filter-cluster", "test-cluster", "--kubeconfig", "from.kubeconfig"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tc := newClusterctlTest(t)
			tc.e.EXPECT().Execute(tc.ctx, tt.wantMoveArgs...)

			if err := tc.clusterctl.MoveManagement(tc.ctx, tt.from, tt.to, tt.clusterName); err != nil {
				t.Fatalf("Clusterctl.MoveManagement() error = %v, want nil", err)
			}
		})
	}
}

func TestClusterctlUpgradeAllProvidersSucess(t *testing.T) {
	tt := newClusterctlTest(t)

	changeDiff := &clusterapi.CAPIChangeDiff{
		Core: &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    "v0.3.19",
		},
		ControlPlane: &types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    "v0.3.19",
		},
		InfrastructureProvider: &types.ComponentChangeDiff{
			ComponentName: "vsphere",
			NewVersion:    "v0.4.1",
		},
		BootstrapProviders: []types.ComponentChangeDiff{
			{
				ComponentName: "kubeadm",
				NewVersion:    "v0.3.19",
			},
			{
				ComponentName: "etcdadm-bootstrap",
				NewVersion:    "v0.1.0",
			},
			{
				ComponentName: "etcdadm-controller",
				NewVersion:    "v0.1.0",
			},
		},
	}

	tt.expectBuildOverrideLayer()
	tt.expectGetProviderEnvMap()
	tt.e.EXPECT().ExecuteWithEnv(tt.ctx, tt.providerEnvMap,
		"upgrade", "apply",
		"--config", test.OfType("string"),
		"--kubeconfig", tt.cluster.KubeconfigFile,
		"--control-plane", "capi-kubeadm-control-plane-system/kubeadm:v0.3.19",
		"--core", "capi-system/cluster-api:v0.3.19",
		"--infrastructure", "capv-system/vsphere:v0.4.1",
		"--bootstrap", "capi-kubeadm-bootstrap-system/kubeadm:v0.3.19",
		"--bootstrap", "etcdadm-bootstrap-provider-system/etcdadm-bootstrap:v0.1.0",
		"--bootstrap", "etcdadm-controller-system/etcdadm-controller:v0.1.0",
	)

	tt.Expect(tt.clusterctl.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.managementComponents, clusterSpec, changeDiff)).To(Succeed())
}

func TestClusterctlUpgradeInfrastructureProvidersSucess(t *testing.T) {
	tt := newClusterctlTest(t)

	changeDiff := &clusterapi.CAPIChangeDiff{
		InfrastructureProvider: &types.ComponentChangeDiff{
			ComponentName: "vsphere",
			NewVersion:    "v0.4.1",
		},
	}

	tt.expectBuildOverrideLayer()
	tt.expectGetProviderEnvMap()
	tt.e.EXPECT().ExecuteWithEnv(tt.ctx, tt.providerEnvMap,
		"upgrade", "apply",
		"--config", test.OfType("string"),
		"--kubeconfig", tt.cluster.KubeconfigFile,
		"--infrastructure", "capv-system/vsphere:v0.4.1",
	)

	tt.Expect(tt.clusterctl.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.managementComponents, clusterSpec, changeDiff)).To(Succeed())
}

func TestClusterctlUpgradeInfrastructureProvidersError(t *testing.T) {
	tt := newClusterctlTest(t)

	changeDiff := &clusterapi.CAPIChangeDiff{
		InfrastructureProvider: &types.ComponentChangeDiff{
			ComponentName: "vsphere",
			NewVersion:    "v0.4.1",
		},
	}

	tt.expectBuildOverrideLayer()
	tt.expectGetProviderEnvMap()
	tt.e.EXPECT().ExecuteWithEnv(tt.ctx, tt.providerEnvMap,
		"upgrade", "apply",
		"--config", test.OfType("string"),
		"--kubeconfig", tt.cluster.KubeconfigFile,
		"--infrastructure", "capv-system/vsphere:v0.4.1",
	).Return(bytes.Buffer{}, errors.New("error in exec"))

	tt.Expect(tt.clusterctl.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.managementComponents, clusterSpec, changeDiff)).NotTo(Succeed())
}

var clusterSpec = test.NewClusterSpec(func(s *cluster.Spec) {
	s.VersionsBundles["1.19"] = versionBundle
	s.Bundles.Spec.VersionsBundles[0] = *versionBundle.VersionsBundle
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
				URI: "public.ecr.aws/l0g8r8j6/cert-manager/cert-manager-acmesolver:v1.1.0",
			},
			Cainjector: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/cert-manager/cert-manager-cainjector:v1.1.0",
			},
			Controller: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/cert-manager/cert-manager-controller:v1.1.0",
			},
			Webhook: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/cert-manager/cert-manager-webhook:v1.1.0",
			},
			Manifest: v1alpha1.Manifest{
				URI: "testdata/fake_manifest.yaml",
			},
			Version: "v1.5.3",
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
		Snow: v1alpha1.SnowBundle{
			Version: "v0.0.0",
		},
		VSphere: v1alpha1.VSphereBundle{
			Version: "v0.7.8",
			ClusterAPIController: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-vsphere/release/manager:v0.7.8-eks-a-0.0.1.build.38",
			},
			KubeProxy: kubeProxyVersion08,
		},
		Nutanix: v1alpha1.NutanixBundle{
			Version: "v1.0.1",
			ClusterAPIController: v1alpha1.Image{
				URI: "public.ecr.aws/release-container-registry/nutanix-cloud-native/cluster-api-provider-nutanix/release/manager:v1.0.1-eks-a-v0.0.0-dev-build.1",
			},
		},
		Tinkerbell: v1alpha1.TinkerbellBundle{
			Version: "v0.1.0",
			ClusterAPIController: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/tinkerbell/cluster-api-provider-tinkerbell:v0.1.0-eks-a-0.0.1.build.38",
			},
		},
		CloudStack: v1alpha1.CloudStackBundle{
			Version: "v0.7.8",
			ClusterAPIController: v1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-cloudstack/release/manager:v0.7.8-eks-a-0.0.1.build.38",
			},
			KubeRbacProxy: kubeProxyVersion08,
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
