package dependencies_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
)

type factoryTest struct {
	*WithT
	clusterConfigFile string
	clusterSpec       *cluster.Spec
	ctx               context.Context
}

func newTest(t *testing.T) *factoryTest {
	clusterConfigFile := "testdata/cluster_vsphere.yaml"
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
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false).
		Build()

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
}

func TestFactoryBuildWithClusterManager(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		WithClusterManager().
		Build()

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
}

func TestFactoryBuildWithMultipleDependencies(t *testing.T) {
	tt := newTest(t)
	deps, err := dependencies.NewFactory().
		WithBootstrapper().
		WithClusterManager().
		WithProvider(tt.clusterConfigFile, tt.clusterSpec.Cluster, false).
		WithFluxAddonClient(tt.ctx, tt.clusterSpec.Cluster, tt.clusterSpec.GitOpsConfig).
		WithWriter().
		WithDiagnosticCollectorImage("public.ecr.aws/collector").
		WithAnalyzerFactory().
		WithCollectorFactory().
		WithTroubleshoot().
		WithCapiUpgrader().
		Build()

	tt.Expect(err).To(BeNil())
	tt.Expect(deps.Bootstrapper).NotTo(BeNil())
	tt.Expect(deps.ClusterManager).NotTo(BeNil())
	tt.Expect(deps.Provider).NotTo(BeNil())
	tt.Expect(deps.FluxAddonClient).NotTo(BeNil())
	tt.Expect(deps.Writer).NotTo(BeNil())
	tt.Expect(deps.AnalyzerFactory).NotTo(BeNil())
	tt.Expect(deps.CollectorFactory).NotTo(BeNil())
	tt.Expect(deps.Troubleshoot).NotTo(BeNil())
	tt.Expect(deps.CapiUpgrader).NotTo(BeNil())
}
