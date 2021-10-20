package diagnostics

import (
	_ "embed"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
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

func (f *eksaDiagnosticBundleFactory) DiagnosticBundle(spec *cluster.Spec, provider providers.Provider, kubeconfig string, bundlePath string) (DiagnosticBundle, error) {
	if bundlePath == "" && spec != nil {
		b, err := f.DiagnosticBundleFromSpec(spec, provider, kubeconfig)
		return b, err
	}
	return f.DiagnosticBundleCustom(kubeconfig, bundlePath), nil
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundleBootstrapCluster(kubeconfig string) (DiagnosticBundle, error) {
	return newDiagnosticBundleBootstrapCluster(f.analyzerFactory, f.collectorFactory, f.client, f.kubectl, kubeconfig, f.writer)
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundleFromSpec(spec *cluster.Spec, provider providers.Provider, kubeconfig string) (DiagnosticBundle, error) {
	b := &EksaDiagnosticBundle{
		bundle: &supportBundle{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SupportBundle",
				APIVersion: troubleshootApiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%sBundle", spec.Name),
			},
			Spec: supportBundleSpec{},
		},
		analyzerFactory:  f.analyzerFactory,
		collectorFactory: f.collectorFactory,
		client:           f.client,
		clusterSpec:      spec,
		kubeconfig:       kubeconfig,
		kubectl:          f.kubectl,
		retrier:          retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
		writer:           f.writer,
	}

	b = b.
		WithGitOpsConfig(spec.GitOpsConfig).
		WithOidcConfig(spec.OIDCConfig).
		WithExternalEtcd(spec.Spec.ExternalEtcdConfiguration).
		WithDatacenterConfig(spec.Spec.DatacenterRef).
		WithMachineConfigs(provider.MachineConfigs()).
		WithDefaultAnalyzers().
		WithDefaultCollectors().
		WithLogTextAnalyzers()

	err := b.WriteBundleConfig(b.clusterSpec.Name)
	if err != nil {
		return nil, fmt.Errorf("error writing bundle config: %v", err)
	}

	return b, nil
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundleDefault() DiagnosticBundle {
	b := &EksaDiagnosticBundle{
		bundle: &supportBundle{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SupportBundle",
				APIVersion: troubleshootApiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "defaultBundle",
			},
			Spec: supportBundleSpec{},
		},
		analyzerFactory:  f.analyzerFactory,
		collectorFactory: f.collectorFactory,
	}
	return b.WithDefaultAnalyzers().WithDefaultCollectors()
}

func (f *eksaDiagnosticBundleFactory) DiagnosticBundleCustom(kubeconfig string, bundlePath string) DiagnosticBundle {
	return &EksaDiagnosticBundle{
		bundlePath:       bundlePath,
		analyzerFactory:  f.analyzerFactory,
		collectorFactory: f.collectorFactory,
		client:           f.client,
		kubeconfig:       kubeconfig,
		kubectl:          f.kubectl,
		retrier:          retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
	}
}
