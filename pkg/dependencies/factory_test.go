package dependencies_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type factoryTest struct {
	*WithT
	clusterConfigFile     string
	clusterSpec           *cluster.Spec
	ctx                   context.Context
	hardwareConfigFile    string
	tinkerbellBootstrapIP string
	cliConfig             config.CliConfig
}

type provider string

const (
	vsphere    provider = "vsphere"
	tinkerbell provider = "tinkerbell"
	nutanix    provider = "nutanix"
)

func newTest(t *testing.T, p provider) *factoryTest {
	var clusterConfigFile string
	switch p {
	case vsphere:
		clusterConfigFile = "testdata/cluster_vsphere.yaml"
	case tinkerbell:
		clusterConfigFile = "testdata/cluster_tinkerbell.yaml"
	case nutanix:
		clusterConfigFile = "testdata/cluster_nutanix.yaml"
	default:
		t.Fatalf("Not a valid provider: %v", p)
	}

	return &factoryTest{
		WithT:             NewGomegaWithT(t),
		clusterConfigFile: clusterConfigFile,
		clusterSpec:       test.NewFullClusterSpec(t, clusterConfigFile),
		ctx:               context.Background(),
	}
}

func TestFactoryBuildWithProvidervSphere(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false, tt.hardwareConfigFile, false, tt.tinkerbellBootstrapIP).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.DockerClient).To(BeNil(), "it only builds deps for vsphere")
}

func TestFactoryBuildWithProviderTinkerbell(t *testing.T) {
	tt := newTest(t, tinkerbell)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false, tt.hardwareConfigFile, false, tt.tinkerbellBootstrapIP).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.Helm).NotTo(BeNil())
	tt.Expect(deps.DockerClient).NotTo(BeNil())
}

func TestFactoryBuildWithProviderNutanix(t *testing.T) {
	tt := newTest(t, nutanix)
	os.Setenv("NUTANIX_USER", "test")
	defer os.Unsetenv("NUTANIX_USER")
	os.Setenv("NUTANIX_PASSWORD", "test")
	defer os.Unsetenv("NUTANIX_PASSWORD")
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false, tt.hardwareConfigFile, false, tt.tinkerbellBootstrapIP).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.NutanixValidator).NotTo(BeNil())
}

func TestFactoryBuildWithClusterManager(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithCliConfig(&tt.cliConfig).
		WithClusterManager(tt.clusterSpec.Cluster).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
}

func TestFactoryBuildWithClusterManagerWithoutCliConfig(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithClusterManager(tt.clusterSpec.Cluster).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
}

func TestFactoryBuildWithMultipleDependencies(t *testing.T) {
	configString := test.ReadFile(t, "testdata/cloudstack_config_multiple_profiles.ini")
	encodedConfig := base64.StdEncoding.EncodeToString([]byte(configString))
	t.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, encodedConfig)

	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithBootstrapper().
		WithCliConfig(&tt.cliConfig).
		WithClusterManager(tt.clusterSpec.Cluster).
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false, tt.hardwareConfigFile, false, tt.tinkerbellBootstrapIP).
		WithGitOpsFlux(tt.clusterSpec.Cluster, tt.clusterSpec.FluxConfig, nil).
		WithWriter().
		WithEksdInstaller().
		WithEksdUpgrader().
		WithDiagnosticCollectorImage("public.ecr.aws/collector").
		WithAnalyzerFactory().
		WithCollectorFactory().
		WithTroubleshoot().
		WithCAPIManager().
		WithManifestReader().
		WithUnAuthKubeClient().
		WithCmk().
		WithVSphereDefaulter().
		WithVSphereValidator().
		WithCiliumTemplater().
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Bootstrapper).NotTo(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.GitOpsFlux).NotTo(BeNil())
	tt.Expect(deps.Writer).NotTo(BeNil())
	tt.Expect(deps.EksdInstaller).NotTo(BeNil())
	tt.Expect(deps.EksdUpgrader).NotTo(BeNil())
	tt.Expect(deps.AnalyzerFactory).NotTo(BeNil())
	tt.Expect(deps.CollectorFactory).NotTo(BeNil())
	tt.Expect(deps.Troubleshoot).NotTo(BeNil())
	tt.Expect(deps.CAPIManager).NotTo(BeNil())
	tt.Expect(deps.ManifestReader).NotTo(BeNil())
	tt.Expect(deps.UnAuthKubeClient).NotTo(BeNil())
	tt.Expect(deps.VSphereDefaulter).NotTo(BeNil())
	tt.Expect(deps.VSphereValidator).NotTo(BeNil())
	tt.Expect(deps.CiliumTemplater).NotTo(BeNil())
}

func TestFactoryBuildWithProxyConfiguration(t *testing.T) {
	tt := newTest(t, vsphere)
	wantHttpsProxy := "FOO"
	wantHttpProxy := "BAR"
	wantNoProxy := "localhost,anotherhost"
	env := map[string]string{
		config.HttpsProxyKey: wantHttpsProxy,
		config.HttpProxyKey:  wantHttpProxy,
		config.NoProxyKey:    wantNoProxy,
	}
	for k, v := range env {
		t.Setenv(k, v)
	}

	f := dependencies.NewFactory().WithProxyConfiguration()

	tt.Expect(f.GetProxyConfiguration()).To(BeNil())

	_, err := f.Build(context.Background())

	pc := f.GetProxyConfiguration()
	tt.Expect(err).To(BeNil())

	tt.Expect(pc[config.HttpsProxyKey]).To(Equal(wantHttpsProxy))
	tt.Expect(pc[config.HttpProxyKey]).To(Equal(wantHttpProxy))
	tt.Expect(pc[config.NoProxyKey]).To(Equal(wantNoProxy))
}

func TestFactoryBuildWithRegistryMirror(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithRegistryMirror("1.2.3.4:443").
		WithHelm(executables.WithInsecure()).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Helm).NotTo(BeNil())
}

func TestFactoryBuildWithPackageInstaller(t *testing.T) {
	spec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-cluster",
				},
			},
		},
		VersionsBundle: &cluster.VersionsBundle{
			VersionsBundle: &v1alpha1.VersionsBundle{
				PackageController: v1alpha1.PackageBundle{
					HelmChart: v1alpha1.Image{
						URI:  "test_registry/test/eks-anywhere-packages:v1",
						Name: "test_chart",
					},
				},
			},
		},
	}
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithHelm(executables.WithInsecure()).
		WithKubectl().
		WithPackageInstaller(spec, "/test/packages.yaml").
		Build(context.Background())
	tt.Expect(err).To(BeNil())
	tt.Expect(deps.PackageInstaller).NotTo(BeNil())
}

func TestFactoryBuildWithCuratedPackagesCustomRegistry(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithHelm(executables.WithInsecure()).
		WithCuratedPackagesRegistry("test_host:8080", "1.22", version.Info{GitVersion: "1.19"}).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.BundleRegistry).NotTo(BeNil())
}

func TestFactoryBuildWithPackageClient(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithKubectl().
		WithPackageClient().
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.PackageClient).NotTo(BeNil())
}

func TestFactoryBuildWithPackageControllerClientNoProxy(t *testing.T) {
	spec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: anywherev1.ClusterSpec{
					ManagementCluster: anywherev1.ManagementCluster{
						Name: "mgmt-1",
					},
				},
			},
		},
		VersionsBundle: &cluster.VersionsBundle{
			VersionsBundle: &v1alpha1.VersionsBundle{
				PackageController: v1alpha1.PackageBundle{
					HelmChart: v1alpha1.Image{
						URI:  "test_registry/test/eks-anywhere-packages:v1",
						Name: "test_chart",
					},
				},
			},
		},
	}

	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithHelm(executables.WithInsecure()).
		WithKubectl().
		WithPackageControllerClient(spec).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.PackageControllerClient).NotTo(BeNil())
}

func TestFactoryBuildWithPackageControllerClientProxy(t *testing.T) {
	spec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: anywherev1.ClusterSpec{
					ProxyConfiguration: &anywherev1.ProxyConfiguration{
						HttpProxy:  "1.1.1.1",
						HttpsProxy: "1.1.1.1",
						NoProxy:    []string{"1.1.1.1"},
					},
				},
			},
		},
		VersionsBundle: &cluster.VersionsBundle{
			VersionsBundle: &v1alpha1.VersionsBundle{
				PackageController: v1alpha1.PackageBundle{
					HelmChart: v1alpha1.Image{
						URI:  "test_registry/test/eks-anywhere-packages:v1",
						Name: "test_chart",
					},
				},
			},
		},
	}

	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithHelm(executables.WithInsecure()).
		WithKubectl().
		WithPackageControllerClient(spec).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.PackageControllerClient).NotTo(BeNil())
}

func TestFactoryBuildWithCuratedPackagesDefaultRegistry(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithManifestReader().
		WithCuratedPackagesRegistry("", "1.22", version.Info{GitVersion: "1.19"}).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.BundleRegistry).NotTo(BeNil())
}

func TestFactoryBuildWithExecutablesUsingDocker(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		UseExecutablesDockerClient(dummyDockerClient{}).
		UseExecutableImage("myimage").
		WithGovc().
		WithHelm().
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Govc).NotTo(BeNil())
	tt.Expect(deps.Helm).NotTo(BeNil())
}

type dummyDockerClient struct{}

func (b dummyDockerClient) PullImage(ctx context.Context, image string) error {
	return nil
}

func (b dummyDockerClient) Execute(ctx context.Context, args ...string) (stdout bytes.Buffer, err error) {
	return bytes.Buffer{}, nil
}
