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
)

const (
	troulbeshootApiVersion    = "troubleshoot.sh/v1beta2"
	generatedBundleNameFormat = "%s-%s-bundle.yaml"
)

type EksaDiagnosticBundleOpts struct {
	AnalyzerFactory  AnalyzerFactory
	BundlePath       string
	CollectorFactory CollectorFactory
	Client           BundleClient
	ClusterSpec      *cluster.Spec
	Kubeconfig       string
	Writer           filewriter.FileWriter
}

type EksaDiagnosticBundle struct {
	bundle           *supportBundle
	bundlePath       string
	clusterSpec      *cluster.Spec
	analyzerFactory  AnalyzerFactory
	collectorFactory CollectorFactory
	client           BundleClient
	kubeconfig       string
	writer           filewriter.FileWriter
}

func NewDiagnosticBundle(opts EksaDiagnosticBundleOpts) (*EksaDiagnosticBundle, error) {
	if opts.BundlePath == "" && opts.ClusterSpec != nil {
		// user did not provide any bundle-config to the support-bundle command, generate one using the default collectors & analyzers
		bundle := NewDiagnosticBundleFromSpec(opts)
		err := bundle.WriteBundleConfig()
		if err != nil {
			return nil, err
		}
		return bundle, nil
	}
	return NewDiagnosticBundleCustom(opts), nil
}

func NewDiagnosticBundleFromSpec(opts EksaDiagnosticBundleOpts) *EksaDiagnosticBundle {
	b := &EksaDiagnosticBundle{
		bundle: &supportBundle{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SupportBundle",
				APIVersion: troulbeshootApiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%sBundle", opts.ClusterSpec.Name),
			},
			Spec: supportBundleSpec{},
		},
		analyzerFactory:  opts.AnalyzerFactory,
		collectorFactory: opts.CollectorFactory,
		client:           opts.Client,
		clusterSpec:      opts.ClusterSpec,
		kubeconfig:       opts.Kubeconfig,
		writer:           opts.Writer,
	}
	return b.
		WithGitOpsConfig(opts.ClusterSpec.GitOpsConfig).
		WithOidcConfig(opts.ClusterSpec.OIDCConfig).
		WithExternalEtcd(opts.ClusterSpec.Spec.ExternalEtcdConfiguration).
		WithDatacenterConfig(opts.ClusterSpec.Spec.DatacenterRef).
		WithDefaultAnalyzers().
		WithDefaultCollectors()
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

func NewDiagnosticBundleCustom(opts EksaDiagnosticBundleOpts) *EksaDiagnosticBundle {
	return &EksaDiagnosticBundle{
		bundlePath:       opts.BundlePath,
		analyzerFactory:  opts.AnalyzerFactory,
		collectorFactory: opts.CollectorFactory,
		client:           opts.Client,
		kubeconfig:       opts.Kubeconfig,
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
