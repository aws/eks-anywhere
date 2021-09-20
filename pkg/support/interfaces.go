package supportbundle

import (
	"context"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
)

type BundleClient interface {
	Collect(ctx context.Context, bundlePath string, sinceTime *time.Time, kubeconfig string) (archivePath string, err error)
	Analyze(ctx context.Context, bundleSpecPath string, archivePath string) ([]*executables.SupportBundleAnalysis, error)
}

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
	DefaultAnalyzers() []*Analyze
	EksaGitopsAnalyzers() []*Analyze
	EksaOidcAnalyzers() []*Analyze
	EksaExternalEtcdAnalyzers() []*Analyze
	DataCenterConfigAnalyzers(datacenter v1alpha1.Ref) []*Analyze
}

type CollectorFactory interface {
	DefaultCollectors() []*Collect
}
