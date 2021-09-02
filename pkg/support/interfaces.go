package supportbundle

import (
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type DiagnosticBundle interface {
	PrintBundleConfig() error
	WithDefaultAnalyzers() *EksaDiagnosticBundle
	WithDefaultCollectors() *EksaDiagnosticBundle
	WithDatacenterConfig(config v1alpha1.Ref) *EksaDiagnosticBundle
	WithOidcConfig(config *v1alpha1.OIDCConfig) *EksaDiagnosticBundle
	WithExternalEtcd(config *v1alpha1.ExternalEtcdConfiguration) *EksaDiagnosticBundle
	WithGitOpsConfig(config *v1alpha1.GitOpsConfig) *EksaDiagnosticBundle
}

type AnalyzerFactory interface {
	DefaultAnalyzers() []*v1beta2.Analyze
	EksaGitopsAnalyzers() []*v1beta2.Analyze
	EksaOidcAnalyzers() []*v1beta2.Analyze
	EksaExternalEtcdAnalyzers() []*v1beta2.Analyze
	DataCenterConfigAnalyzers(datacenter v1alpha1.Ref) []*v1beta2.Analyze
}

type CollectorFactory interface {
	DefaultCollectors() []*v1beta2.Collect
}
