package supportbundle

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
)

const (
	troulbeshootApiVersion    = "troubleshoot.sh/v1beta2"
	generatedBundleNameFormat = "%s-%s-bundle.yaml"
)

type EksaDiagnosticBundleOpts struct {
	AnalyzerFactory  AnalyzerFactory
	Client           BundleClient
	CollectorFactory CollectorFactory
	Writer           filewriter.FileWriter
}

type EksaDiagnosticBundle struct {
	bundle           *supportBundle
	bundlePath       string
	client           BundleClient
	collectorFactory CollectorFactory
	clusterSpec      *cluster.Spec
	analyzerFactory  AnalyzerFactory
	kubeconfig       string
	writer           filewriter.FileWriter
}

func NewDiagnosticBundle(spec *cluster.Spec, provider providers.Provider, kubeconfig string, bundlePath string, opts EksaDiagnosticBundleOpts) (*EksaDiagnosticBundle, error) {
	if bundlePath == "" && spec != nil {
		return NewDiagnosticBundleFromSpec(spec, provider, kubeconfig, opts)
	}
	return NewDiagnosticBundleCustom(kubeconfig, bundlePath, opts), nil
}

func NewDiagnosticBundleFromSpec(spec *cluster.Spec, provider providers.Provider, kubeconfig string, opts EksaDiagnosticBundleOpts) (*EksaDiagnosticBundle, error) {
	b := &EksaDiagnosticBundle{
		bundle: &supportBundle{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SupportBundle",
				APIVersion: troulbeshootApiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%sBundle", spec.Name),
			},
			Spec: supportBundleSpec{},
		},
		analyzerFactory:  opts.AnalyzerFactory,
		collectorFactory: opts.CollectorFactory,
		client:           opts.Client,
		clusterSpec:      spec,
		kubeconfig:       kubeconfig,
		writer:           opts.Writer,
	}

	osFamiliesSet := map[v1alpha1.OSFamily]bool{}
	for _, config := range provider.MachineConfigs() {
		osFamiliesSet[config.OSFamily()] = true
	}

	b = b.
		WithGitOpsConfig(spec.GitOpsConfig).
		WithOidcConfig(spec.OIDCConfig).
		WithExternalEtcd(spec.Spec.ExternalEtcdConfiguration).
		WithDatacenterConfig(spec.Spec.DatacenterRef).
		WithOSFamilies(osFamiliesSet).
		WithDefaultAnalyzers().
		WithDefaultCollectors()

	err := b.WriteBundleConfig()
	if err != nil {
		return nil, fmt.Errorf("error writing bundle config: %v", err)
	}

	return b, nil
}

func NewDiagnosticBundleDefault(af AnalyzerFactory, cf CollectorFactory) *EksaDiagnosticBundle {
	b := &EksaDiagnosticBundle{
		bundle: &supportBundle{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SupportBundle",
				APIVersion: troulbeshootApiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "defaultBundle",
			},
			Spec: supportBundleSpec{},
		},
		analyzerFactory:  af,
		collectorFactory: cf,
	}
	return b.WithDefaultAnalyzers().WithDefaultCollectors()
}

func NewDiagnosticBundleCustom(kubeconfig string, bundlePath string, opts EksaDiagnosticBundleOpts) *EksaDiagnosticBundle {
	return &EksaDiagnosticBundle{
		bundlePath:       bundlePath,
		analyzerFactory:  opts.AnalyzerFactory,
		collectorFactory: opts.CollectorFactory,
		client:           opts.Client,
		kubeconfig:       kubeconfig,
	}
}

func (e *EksaDiagnosticBundle) CollectAndAnalyze(ctx context.Context, sinceTimeValue *time.Time) error {
	archivePath, err := e.client.Collect(ctx, e.bundlePath, sinceTimeValue, e.kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to Collect and Analyze support bundle: %v", err)
	}

	analysis, err := e.client.Analyze(ctx, e.bundlePath, archivePath)
	if err != nil {
		return fmt.Errorf("error when analyzing bundle: %v", err)
	}

	yamlAnalysis, err := yaml.Marshal(analysis)
	if err != nil {
		return fmt.Errorf("error while analyzing bundle: %v", err)
	}

	logger.Info("Support bundle archive created", "archivePath", archivePath)
	fmt.Println(string(yamlAnalysis))
	return nil
}

func (e *EksaDiagnosticBundle) PrintBundleConfig() error {
	bundleYaml, err := yaml.Marshal(e.bundle)
	if err != nil {
		return fmt.Errorf("error outputting yaml: %v", err)
	}
	fmt.Println(string(bundleYaml))
	return nil
}

func (e *EksaDiagnosticBundle) WriteBundleConfig() error {
	bundleYaml, err := yaml.Marshal(e.bundle)
	if err != nil {
		return fmt.Errorf("error outputing yaml: %v", err)
	}
	timestamp := time.Now().Format(time.RFC3339)
	filename := fmt.Sprintf(generatedBundleNameFormat, e.clusterSpec.Name, timestamp)
	e.bundlePath, err = e.writer.Write(filename, bundleYaml)
	logger.V(3).Info("bundle config written", "path", e.bundlePath)
	if err != nil {
		return err
	}
	return nil
}

func (e *EksaDiagnosticBundle) WithDefaultCollectors() *EksaDiagnosticBundle {
	e.bundle.Spec.Collectors = append(e.bundle.Spec.Collectors, e.collectorFactory.DefaultCollectors()...)
	return e
}

func (e *EksaDiagnosticBundle) WithDefaultAnalyzers() *EksaDiagnosticBundle {
	e.bundle.Spec.Analyzers = append(e.bundle.Spec.Analyzers, e.analyzerFactory.DefaultAnalyzers()...)
	return e
}

func (e *EksaDiagnosticBundle) WithDatacenterConfig(config v1alpha1.Ref) *EksaDiagnosticBundle {
	e.bundle.Spec.Analyzers = append(e.bundle.Spec.Analyzers, e.analyzerFactory.DataCenterConfigAnalyzers(config)...)
	return e
}

func (e *EksaDiagnosticBundle) WithOidcConfig(config *v1alpha1.OIDCConfig) *EksaDiagnosticBundle {
	if config != nil {
		e.bundle.Spec.Analyzers = append(e.bundle.Spec.Analyzers, e.analyzerFactory.EksaOidcAnalyzers()...)
	}
	return e
}

func (e *EksaDiagnosticBundle) WithExternalEtcd(config *v1alpha1.ExternalEtcdConfiguration) *EksaDiagnosticBundle {
	if config != nil {
		e.bundle.Spec.Analyzers = append(e.bundle.Spec.Analyzers, e.analyzerFactory.EksaExternalEtcdAnalyzers()...)
	}
	return e
}

func (e *EksaDiagnosticBundle) WithGitOpsConfig(config *v1alpha1.GitOpsConfig) *EksaDiagnosticBundle {
	if config != nil {
		e.bundle.Spec.Analyzers = append(e.bundle.Spec.Analyzers, e.analyzerFactory.EksaGitopsAnalyzers()...)
	}
	return e
}

func (e *EksaDiagnosticBundle) WithOSFamilies(osFamilies map[v1alpha1.OSFamily]bool) *EksaDiagnosticBundle {
	e.bundle.Spec.Collectors = append(e.bundle.Spec.Collectors, e.collectorFactory.EksaHostCollectors(osFamilies)...)
	return e
}

func ParseTimeOptions(since string, sinceTime string) (*time.Time, error) {
	var sinceTimeValue time.Time
	var err error
	if sinceTime == "" && since == "" {
		return &sinceTimeValue, nil
	} else if sinceTime != "" && since != "" {
		return nil, fmt.Errorf("at most one of `sinceTime` or `since` could be specified")
	} else if sinceTime != "" {
		sinceTimeValue, err = time.Parse(time.RFC3339, sinceTime)
		if err != nil {
			return nil, fmt.Errorf("unable to parse --since-time option, ensure since-time is RFC3339 formatted. error: %v", err)
		}
	} else if since != "" {
		duration, err := time.ParseDuration(since)
		if err != nil {
			return nil, fmt.Errorf("unable to parse --since option: %v", err)
		}
		now := time.Now()
		sinceTimeValue = now.Add(0 - duration)
	}
	return &sinceTimeValue, nil
}
