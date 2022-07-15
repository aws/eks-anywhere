package dependencies_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/dependencies"
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
)

func newTest(t *testing.T, p provider) *factoryTest {
	var clusterConfigFile string
	switch p {
	case vsphere:
		clusterConfigFile = "testdata/cluster_vsphere.yaml"
	case tinkerbell:
		clusterConfigFile = "testdata/cluster_tinkerbell.yaml"
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
	tt.Expect(deps.HelmSecure).NotTo(BeNil())
	tt.Expect(deps.DockerClient).NotTo(BeNil())
	tt.Expect(deps.HelmInsecure).To(BeNil(), "should not build HelmInsecure")
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
		WithFluxAddonClient(tt.clusterSpec.Cluster, tt.clusterSpec.FluxConfig, nil).
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
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Bootstrapper).NotTo(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.FluxAddonClient).NotTo(BeNil())
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
}

func TestFactoryBuildWithRegistryMirror(t *testing.T) {
	tt := newTest(t, vsphere)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithRegistryMirror("1.2.3.4:443").
		WithHelmInsecure().
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.HelmInsecure).NotTo(BeNil())
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
		WithHelmInsecure().
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
		WithHelmInsecure().
		WithCuratedPackagesRegistry("test_host:8080", "1.22", version.Info{GitVersion: "1.19"}).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.BundleRegistry).NotTo(BeNil())
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
		WithHelmSecure().
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Govc).NotTo(BeNil())
	tt.Expect(deps.HelmSecure).NotTo(BeNil())
}

type dummyDockerClient struct{}

func (b dummyDockerClient) PullImage(ctx context.Context, image string) error {
	return nil
}

func (b dummyDockerClient) Execute(ctx context.Context, args ...string) (stdout bytes.Buffer, err error) {
	return bytes.Buffer{}, nil
}
