package diagnostics_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
)

func TestManagementClusterAnalyzers(t *testing.T) {
	g := NewGomegaWithT(t)
	factory := diagnostics.NewAnalyzerFactory()
	analyzers := factory.ManagementClusterAnalyzers()
	g.Expect(analyzers).To(HaveLen(13), "DataCenterConfigCollectors() mismatch between desired collectors and actual")
	g.Expect(getDeploymentStatusAnalyzer(analyzers, "capc-controller-manager")).ToNot(BeNil(), "capc controller manager analyzer should be present")
	g.Expect(getDeploymentStatusAnalyzer(analyzers, "capv-controller-manager")).ToNot(BeNil(), "capv controller manager analyzer should be present")
	g.Expect(getDeploymentStatusAnalyzer(analyzers, "capt-controller-manager")).ToNot(BeNil(), "capt controller manager analyzer should be present")
	g.Expect(getDeploymentStatusAnalyzer(analyzers, "capx-controller-manager")).ToNot(BeNil(), "capx controller manager analyzer should be present")
	g.Expect(analyzers[11].CustomResourceDefinition.CheckName).To(Equal("clusters.anywhere.eks.amazonaws.com"))
	g.Expect(analyzers[12].CustomResourceDefinition.CheckName).To(Equal("bundles.anywhere.eks.amazonaws.com"))
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
	collectorFactory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := collectorFactory.DefaultCollectors()
	collectors = append(collectors, collectorFactory.ManagementClusterCollectors()...)
	analyzerFactory := diagnostics.NewAnalyzerFactory()
	expectAnalzyers := analyzerFactory.EksaLogTextAnalyzers(collectors)
	for _, analyzer := range expectAnalzyers {
		if analyzer == nil {
			t.Errorf("EksaLogTextAnalyzers failed: return a nil analyzer")
		}
	}
}

func TestVsphereDataCenterConfigAnalyzers(t *testing.T) {
	g := NewGomegaWithT(t)
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.VSphereDatacenterKind}
	analyzerFactory := diagnostics.NewAnalyzerFactory()
	analyzers := analyzerFactory.DataCenterConfigAnalyzers(datacenter)
	g.Expect(analyzers).To(HaveLen(4), "DataCenterConfigAnalyzers() mismatch between desired analyzers and actual")
	g.Expect(analyzers[0].CustomResourceDefinition.CustomResourceDefinitionName).To(Equal("vspheredatacenterconfigs.anywhere.eks.amazonaws.com"),
		"vSphere generateCrdAnalyzers() mismatch between desired datacenter config group version and actual")
	g.Expect(analyzers[1].CustomResourceDefinition.CustomResourceDefinitionName).To(Equal("vspheremachineconfigs.anywhere.eks.amazonaws.com"),
		"vSphere generateCrdAnalyzers() mismatch between desired machine config group version and actual")
	g.Expect(analyzers[2].TextAnalyze.RegexPattern).To(Equal("exit code: 0"),
		"validControlPlaneIPAnalyzer() mismatch between desired regexPattern and actual")
	g.Expect(analyzers[3].TextAnalyze.RegexPattern).To(Equal("session \"msg\"=\"error checking if session is active\" \"error\"=\"ServerFaultCode: Permission to perform this operation was denied.\""),
		"vcenterSessionValidatePermissionAnalyzer() mismatch between desired regexPattern and actual")
}

func TestDockerDataCenterConfigAnalyzers(t *testing.T) {
	g := NewGomegaWithT(t)
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.DockerDatacenterKind}
	analyzerFactory := diagnostics.NewAnalyzerFactory()
	analyzers := analyzerFactory.DataCenterConfigAnalyzers(datacenter)
	g.Expect(analyzers).To(HaveLen(2), "DataCenterConfigAnalyzers() mismatch between desired analyzers and actual")
}

func TestCloudStackDataCenterConfigAnalyzers(t *testing.T) {
	g := NewGomegaWithT(t)
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.CloudStackDatacenterKind}
	analyzerFactory := diagnostics.NewAnalyzerFactory()
	analyzers := analyzerFactory.DataCenterConfigAnalyzers(datacenter)
	g.Expect(analyzers).To(HaveLen(3), "DataCenterConfigAnalyzers() mismatch between desired analyzers and actual")
}

func TestNutanixDataCenterConfigAnalyzers(t *testing.T) {
	g := NewGomegaWithT(t)
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.NutanixDatacenterKind}
	analyzerFactory := diagnostics.NewAnalyzerFactory()
	analyzers := analyzerFactory.DataCenterConfigAnalyzers(datacenter)
	g.Expect(analyzers).To(HaveLen(3), "DataCenterConfigAnalyzers() mismatch between desired analyzers and actual")
	g.Expect(analyzers[0].CustomResourceDefinition.CustomResourceDefinitionName).To(Equal("nutanixdatacenterconfigs.anywhere.eks.amazonaws.com"),
		"Nutanix generateCrdAnalyzers() mismatch between desired datacenter config group version and actual")
	g.Expect(analyzers[1].CustomResourceDefinition.CustomResourceDefinitionName).To(Equal("nutanixmachineconfigs.anywhere.eks.amazonaws.com"),
		"Nutanix generateCrdAnalyzers() mismatch between desired machine config group version and actual")
	g.Expect(analyzers[2].TextAnalyze.RegexPattern).To(Equal("exit code: 0"),
		"validControlPlaneIPAnalyzer() mismatch between desired regexPattern and actual")
}

func TestSnowAnalyzers(t *testing.T) {
	g := NewGomegaWithT(t)
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.SnowDatacenterKind}
	analyzerFactory := diagnostics.NewAnalyzerFactory()
	analyzers := analyzerFactory.DataCenterConfigAnalyzers(datacenter)
	g.Expect(analyzers).To(HaveLen(3), "DataCenterConfigAnalyzers() mismatch between desired analyzers and actual")
}
