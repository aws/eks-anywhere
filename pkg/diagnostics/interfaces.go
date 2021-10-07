package diagnostics

import (
	"context"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers"
)

type BundleClient interface {
	Collect(ctx context.Context, bundlePath string, sinceTime *time.Time, kubeconfig string) (archivePath string, err error)
	Analyze(ctx context.Context, bundleSpecPath string, archivePath string) ([]*executables.SupportBundleAnalysis, error)
}

type DiagnosticBundleFactory interface {
	NewDiagnosticBundle(spec *cluster.Spec, provider providers.Provider, kubeconfig string, bundlePath string) (*EksaDiagnosticBundle, error)
	NewDiagnosticBundleFromSpec(spec *cluster.Spec, provider providers.Provider, kubeconfig string) (*EksaDiagnosticBundle, error)
	NewDiagnosticBundleDefault() *EksaDiagnosticBundle
	NewDiagnosticBundleCustom(kubeconfig string, bundlePath string) *EksaDiagnosticBundle
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
	EksaLogTextAnalyzers(collectors []*Collect) []*Analyze
	EksaOidcAnalyzers() []*Analyze
	EksaExternalEtcdAnalyzers() []*Analyze
	DataCenterConfigAnalyzers(datacenter v1alpha1.Ref) []*Analyze
}

type CollectorFactory interface {
	DefaultCollectors() []*Collect
	EksaHostCollectors(configs []providers.MachineConfig) []*Collect
}
