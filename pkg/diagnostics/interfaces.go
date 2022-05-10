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
	DiagnosticBundle(spec *cluster.Spec, provider providers.Provider, kubeconfig string, bundlePath string) (DiagnosticBundle, error)
	DiagnosticBundleFromSpec(spec *cluster.Spec, provider providers.Provider, kubeconfig string) (DiagnosticBundle, error)
	DiagnosticBundleManagementCluster(kubeconfig string) (DiagnosticBundle, error)
	DiagnosticBundleDefault() DiagnosticBundle
	DiagnosticBundleCustom(kubeconfig string, bundlePath string) DiagnosticBundle
}

type DiagnosticBundle interface {
	PrintBundleConfig() error
	WriteBundleConfig() error
	PrintAnalysis() error
	WriteAnalysisToFile() (path string, err error)
	CollectAndAnalyze(ctx context.Context, sinceTimeValue *time.Time) error
	WithDefaultAnalyzers() *EksaDiagnosticBundle
	WithDefaultCollectors() *EksaDiagnosticBundle
	WithDatacenterConfig(config v1alpha1.Ref) *EksaDiagnosticBundle
	WithOidcConfig(config *v1alpha1.OIDCConfig) *EksaDiagnosticBundle
	WithExternalEtcd(config *v1alpha1.ExternalEtcdConfiguration) *EksaDiagnosticBundle
	WithGitOpsConfig(config *v1alpha1.GitOpsConfig) *EksaDiagnosticBundle
	WithMachineConfigs(configs []providers.MachineConfig) *EksaDiagnosticBundle
	WithLogTextAnalyzers() *EksaDiagnosticBundle
}

type AnalyzerFactory interface {
	DefaultAnalyzers() []*Analyze
	EksaGitopsAnalyzers() []*Analyze
	EksaLogTextAnalyzers(collectors []*Collect) []*Analyze
	EksaOidcAnalyzers() []*Analyze
	EksaExternalEtcdAnalyzers() []*Analyze
	DataCenterConfigAnalyzers(datacenter v1alpha1.Ref) []*Analyze
	ManagementClusterAnalyzers() []*Analyze
	PackageAnalyzers() []*Analyze
}

type CollectorFactory interface {
	PackagesCollectors() []*Collect
	DefaultCollectors() []*Collect
	ManagementClusterCollectors() []*Collect
	EksaHostCollectors(configs []providers.MachineConfig) []*Collect
	DataCenterConfigCollectors(datacenter v1alpha1.Ref) []*Collect
}
