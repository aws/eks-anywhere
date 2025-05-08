package diagnostics

import (
	_ "embed"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
)

type EksaDiagnosticBundleFactoryOpts struct {
	AnalyzerFactory  AnalyzerFactory
	Client           BundleClient
	CollectorFactory CollectorFactory
	Kubectl          *executables.Kubectl
	Writer           filewriter.FileWriter
}

type eksaDiagnosticBundleFactory struct {
	analyzerFactory  AnalyzerFactory
	client           BundleClient
	collectorFactory CollectorFactory
	kubectl          *executables.Kubectl
	writer           filewriter.FileWriter
}

func NewFactory(opts EksaDiagnosticBundleFactoryOpts) *eksaDiagnosticBundleFactory {
	return &eksaDiagnosticBundleFactory{
		analyzerFactory:  opts.AnalyzerFactory,
		client:           opts.Client,
		collectorFactory: opts.CollectorFactory,
		kubectl:          opts.Kubectl,
		writer:           opts.Writer,
	}
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundle(spec *cluster.Spec, provider providers.Provider, kubeconfig string, bundlePath string, auditLogs bool) (DiagnosticBundle, error) {
	if bundlePath == "" && spec != nil {
		b, err := f.DiagnosticBundleWorkloadCluster(spec, provider, kubeconfig, auditLogs)
		return b, err
	}
	return f.DiagnosticBundleCustom(kubeconfig, bundlePath), nil
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundleManagementCluster(spec *cluster.Spec, kubeconfig string) (DiagnosticBundle, error) {
	return newDiagnosticBundleManagementCluster(f.analyzerFactory, f.collectorFactory, spec, f.client, f.kubectl, kubeconfig, f.writer)
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundleWorkloadCluster(spec *cluster.Spec, provider providers.Provider, kubeconfig string, auditLogs bool) (DiagnosticBundle, error) {
	return newDiagnosticBundleFromSpec(f.analyzerFactory, f.collectorFactory, spec, provider, f.client, f.kubectl, kubeconfig, f.writer, auditLogs)
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundleDefault() DiagnosticBundle {
	return newDiagnosticBundleDefault(f.analyzerFactory, f.collectorFactory)
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundleCustom(kubeconfig string, bundlePath string) DiagnosticBundle {
	return newDiagnosticBundleCustom(f.analyzerFactory, f.collectorFactory, f.client, f.kubectl, bundlePath, kubeconfig, f.writer)
}
