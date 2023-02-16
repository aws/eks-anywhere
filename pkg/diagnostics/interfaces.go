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
	DiagnosticBundleWorkloadCluster(spec *cluster.Spec, provider providers.Provider, kubeconfig string) (DiagnosticBundle, error)
	DiagnosticBundleManagementCluster(spec *cluster.Spec, kubeconfig string) (DiagnosticBundle, error)
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
	WithFileCollectors(paths []string) *EksaDiagnosticBundle
	WithDatacenterConfig(config v1alpha1.Ref, spec *cluster.Spec) *EksaDiagnosticBundle
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

// CollectorFactory generates support-bundle collectors.
type CollectorFactory interface {
	PackagesCollectors() []*Collect
	DefaultCollectors() []*Collect
	FileCollectors(paths []string) []*Collect
	ManagementClusterCollectors() []*Collect
	EksaHostCollectors(configs []providers.MachineConfig) []*Collect
	DataCenterConfigCollectors(datacenter v1alpha1.Ref, spec *cluster.Spec) []*Collect
}
