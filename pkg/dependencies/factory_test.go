package dependencies_test

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
)

type factoryTest struct {
	*WithT
	clusterConfigFile  string
	clusterSpec        *cluster.Spec
	ctx                context.Context
	hardwareConfigFile string
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
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false, tt.hardwareConfigFile, false, false).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.DockerClient).To(BeNil(), "it only builds deps for vsphere")
}

func TestFactoryBuildWithClusterManager(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithClusterManager(tt.clusterSpec.Cluster).
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
}

func TestFactoryBuildWithMultipleDependencies(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		UseExecutableImage("image:1").
		WithBootstrapper().
		WithClusterManager(tt.clusterSpec.Cluster).
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false, tt.hardwareConfigFile, false, false).
		WithFluxAddonClient(tt.ctx, tt.clusterSpec.Cluster, tt.clusterSpec.GitOpsConfig).
		WithWriter().
		WithDiagnosticCollectorImage("public.ecr.aws/collector").
		WithAnalyzerFactory().
		WithCollectorFactory().
		WithTroubleshoot().
		WithCAPIManager().
		WithManifestReader().
		Build(context.Background())

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Bootstrapper).NotTo(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.FluxAddonClient).NotTo(BeNil())
	tt.Expect(deps.Writer).NotTo(BeNil())
	tt.Expect(deps.AnalyzerFactory).NotTo(BeNil())
	tt.Expect(deps.CollectorFactory).NotTo(BeNil())
	tt.Expect(deps.Troubleshoot).NotTo(BeNil())
	tt.Expect(deps.CAPIManager).NotTo(BeNil())
	tt.Expect(deps.ManifestReader).NotTo(BeNil())
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
