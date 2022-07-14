package diagnostics_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/diagnostics"
)

func TestManagementClusterAnalyzers(t *testing.T) {
	factory := diagnostics.NewAnalyzerFactory()
	analyzers := factory.ManagementClusterAnalyzers()
	assert.Equal(t, len(analyzers), 12, "DataCenterConfigCollectors() mismatch between desired collectors and actual")
	assert.NotNilf(t, getDeploymentStatusAnalyzer(analyzers, "capc-controller-manager"), "capc controller manager analyzer should be present")
	assert.NotNilf(t, getDeploymentStatusAnalyzer(analyzers, "capv-controller-manager"), "capv controller manager analyzer should be present")
	assert.NotNilf(t, getDeploymentStatusAnalyzer(analyzers, "capt-controller-manager"), "capt controller manager analyzer should be present")
}

func getDeploymentStatusAnalyzer(analyzers []*diagnostics.Analyze, name string) *diagnostics.Analyze {
	for _, analyzer := range analyzers {
		if analyzer.DeploymentStatus != nil && analyzer.DeploymentStatus.Name == name {
			return analyzer
		}
	}

	return nil
}

func TestEksaLogTextAnalyzers(t *testing.T) {
	collectorFactory := diagnostics.NewDefaultCollectorFactory()
	collectors := collectorFactory.ManagementClusterCollectors()
	analyzerFactory := diagnostics.NewAnalyzerFactory()
	expectAnalzyers := analyzerFactory.EksaLogTextAnalyzers(collectors)
	for _, analyzer := range expectAnalzyers {
		if analyzer == nil {
			t.Errorf("EksaLogTextAnalyzers failed: return a nil analyzer")
		}
	}
}
