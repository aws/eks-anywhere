package dependencies_test

import (
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

func newTest(t *testing.T) *factoryTest {
	clusterConfigFile := "testdata/cluster_vsphere.yaml"
	// Disable tools image executable for the tests
	if err := os.Setenv("MR_TOOLS_DISABLE", "true"); err != nil {
		t.Fatal(err)
	}

	return &factoryTest{
		WithT:             NewGomegaWithT(t),
		clusterConfigFile: clusterConfigFile,
		clusterSpec:       test.NewFullClusterSpec(t, clusterConfigFile),
		ctx:               context.Background(),
	}
}

func TestFactoryBuildWithProvider(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false, tt.hardwareConfigFile, false, tt.tinkerbellBootstrapIP).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.DockerClient).To(BeNil(), "it only builds deps for vsphere")
}

func TestFactoryBuildWithClusterManager(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithCliConfig(&tt.cliConfig).
		WithClusterManager(tt.clusterSpec.Cluster).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
}

func TestFactoryBuildWithClusterManagerWithoutCliConfig(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithClusterManager(tt.clusterSpec.Cluster).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
}

func TestFactoryBuildWithMultipleDependencies(t *testing.T) {
	configString := test.ReadFile(t, "testdata/cloudstack_config_multiple_profiles.ini")
	encodedConfig := base64.StdEncoding.EncodeToString([]byte(configString))
	t.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, encodedConfig)

	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
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
}

func TestFactoryBuildWithRegistryMirror(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithRegistryMirror("1.2.3.4:443").
		WithHelm().
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
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithHelm().
		WithKubectl().
		WithPackageInstaller(spec, "/test/packages.yaml").
		Build(context.Background())
	tt.Expect(err).To(BeNil())
	tt.Expect(deps.PackageInstaller).NotTo(BeNil())
}

func TestFactoryBuildWithCuratedPackagesCustomRegistry(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithHelm().
		WithCuratedPackagesRegistry("test_host:8080", "1.22", version.Info{GitVersion: "1.19"}).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.BundleRegistry).NotTo(BeNil())
}

func TestFactoryBuildWithCuratedPackagesDefaultRegistry(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithManifestReader().
		WithCuratedPackagesRegistry("", "1.22", version.Info{GitVersion: "1.19"}).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.BundleRegistry).NotTo(BeNil())
}
