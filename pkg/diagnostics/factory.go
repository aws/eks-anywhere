package diagnostics

import (
	_ "embed"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

//go:embed config/diagnostic-collector-rbac.yaml
var diagnosticCollectorRbac []byte

const (
	troubleshootApiVersion    = "troubleshoot.sh/v1beta2"
	generatedBundleNameFormat = "%s-%s-bundle.yaml"
	maxRetries                = 5
	backOffPeriod             = 5 * time.Second
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

func (f *eksaDiagnosticBundleFactory) NewDiagnosticBundle(spec *cluster.Spec, provider providers.Provider, kubeconfig string, bundlePath string) (*EksaDiagnosticBundle, error) {
	if bundlePath == "" && spec != nil {
		return f.NewDiagnosticBundleFromSpec(spec, provider, kubeconfig)
	}
	return f.NewDiagnosticBundleCustom(kubeconfig, bundlePath), nil
}

func (f *eksaDiagnosticBundleFactory) NewDiagnosticBundleFromSpec(spec *cluster.Spec, provider providers.Provider, kubeconfig string) (*EksaDiagnosticBundle, error) {
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

	err := b.WriteBundleConfig()
	if err != nil {
		return nil, fmt.Errorf("error writing bundle config: %v", err)
	}

	return b, nil
}

func (f *eksaDiagnosticBundleFactory) NewDiagnosticBundleDefault() *EksaDiagnosticBundle {
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

func (f *eksaDiagnosticBundleFactory) NewDiagnosticBundleCustom(kubeconfig string, bundlePath string) *EksaDiagnosticBundle {
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
