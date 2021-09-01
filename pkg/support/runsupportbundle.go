package supportbundle

import (
	_ "embed"
	"fmt"
	"time"

	analyzerunner "github.com/replicatedhq/troubleshoot/pkg/analyze"
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	"github.com/replicatedhq/troubleshoot/pkg/k8sutil"
	"github.com/replicatedhq/troubleshoot/pkg/supportbundle"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const resultsSeparator = "\n------------\n"

func ParseBundleFromDoc(clusterSpec *cluster.Spec, bundleConfig string) (*EksaDiagnosticBundle, error) {
	af := NewAnalyzerFactory()
	cf := NewCollectorFactory()
	if bundleConfig == "" {
		// user did not provide any bundle-config to the support-bundle command, generate one using the default collectors & analyzers
		return NewBundleConfig(clusterSpec, af, cf), nil
	}

	// parse bundle-config provided by the user
	collectorContent, err := supportbundle.LoadSupportBundleSpec(bundleConfig)
	if err != nil {
		return nil, err
	}
	bundle, err := supportbundle.ParseSupportBundleFromDoc(collectorContent)
	if err != nil {
		return nil, err
	}
	return NewCustomBundleConfig(bundle, af, cf), nil
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
			return nil, fmt.Errorf("unable to parse --since-time option: %v", err)
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

type EksaDiagnosticBundle struct {
	Bundle           *v1beta2.SupportBundle
	AnalyzerFactory  AnalyzerFactory
	CollectorFactory CollectorFactory
}

func NewCustomBundleConfig(customBundle *v1beta2.SupportBundle, af AnalyzerFactory, cf CollectorFactory) *EksaDiagnosticBundle {
	return &EksaDiagnosticBundle{
		Bundle:           customBundle,
		AnalyzerFactory:  af,
		CollectorFactory: cf,
	}
}

func NewDefaultBundleConfig(af AnalyzerFactory, cf CollectorFactory) *EksaDiagnosticBundle {
	b := &EksaDiagnosticBundle{
		Bundle: &v1beta2.SupportBundle{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SupportBundle",
				APIVersion: "troubleshoot.sh/v1beta2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "defaultBundle",
			},
			Spec: v1beta2.SupportBundleSpec{},
		},
		AnalyzerFactory:  af,
		CollectorFactory: cf,
	}
	return b.WithDefaultAnalyzers().WithDefaultCollectors()
}

func NewBundleConfig(spec *cluster.Spec, af AnalyzerFactory, cf CollectorFactory) *EksaDiagnosticBundle {
	b := &EksaDiagnosticBundle{
		Bundle: &v1beta2.SupportBundle{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SupportBundle",
				APIVersion: "troubleshoot.sh/v1beta2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%sBundle", spec.Name),
			},
			Spec: v1beta2.SupportBundleSpec{},
		},
		AnalyzerFactory:  af,
		CollectorFactory: cf,
	}
	return b.
		WithGitOpsConfig(spec.GitOpsConfig).
		WithOidcConfig(spec.OIDCConfig).
		WithExternalEtcd(spec.Spec.ExternalEtcdConfiguration).
		WithDatacenterConfig(spec.Spec.DatacenterRef).
		WithDefaultAnalyzers().
		WithDefaultCollectors()
}

func (e *EksaDiagnosticBundle) CollectBundleFromSpec(sinceTimeValue *time.Time) (string, error) {
	k8sConfig, err := k8sutil.GetRESTConfig()
	if err != nil {
		return "", fmt.Errorf("failed to convert kube flags to rest config: %v", err)
	}

	progressChan := make(chan interface{})
	go func() {
		var lastMsg string
		for {
			msg := <-progressChan
			switch msg := msg.(type) {
			case error:
				logger.Info(fmt.Sprintf("\r * %v", msg))
			case string:
				if lastMsg != msg {
					logger.Info(fmt.Sprintf("\r \033[36mCollecting support bundle\033[m %v", msg))
					lastMsg = msg
				}
			}
		}
	}()

	collectorCB := func(c chan interface{}, msg string) {
		c <- msg
	}
	additionalRedactors := &v1beta2.Redactor{}
	createOpts := supportbundle.SupportBundleCreateOpts{
		CollectorProgressCallback: collectorCB,
		KubernetesRestConfig:      k8sConfig,
		ProgressChan:              progressChan,
		SinceTime:                 sinceTimeValue,
	}

	archivePath, err := supportbundle.CollectSupportBundleFromSpec(&e.Bundle.Spec, additionalRedactors, createOpts)
	if err != nil {
		return "", err
	}
	return archivePath, nil
}

func (e *EksaDiagnosticBundle) AnalyzeBundle(archivePath string) error {
	analyzeResults, err := supportbundle.AnalyzeAndExtractSupportBundle(&e.Bundle.Spec, archivePath)
	if err != nil {
		return err
	}
	if len(analyzeResults) > 0 {
		showAnalyzeResults(analyzeResults)
	}
	return nil
}

func (e *EksaDiagnosticBundle) PrintBundleConfig() error {
	bundleYaml, err := yaml.Marshal(e.Bundle)
	if err != nil {
		return fmt.Errorf("error outputting yaml: %v", err)
	}
	fmt.Println(string(bundleYaml))
	return nil
}

func (e *EksaDiagnosticBundle) WithDefaultCollectors() *EksaDiagnosticBundle {
	e.Bundle.Spec.Collectors = append(e.Bundle.Spec.Collectors, e.CollectorFactory.DefaultCollectors()...)
	return e
}

func (e *EksaDiagnosticBundle) WithDefaultAnalyzers() *EksaDiagnosticBundle {
	e.Bundle.Spec.Analyzers = append(e.Bundle.Spec.Analyzers, e.AnalyzerFactory.DefaultAnalyzers()...)
	return e
}

func (e *EksaDiagnosticBundle) WithDatacenterConfig(config v1alpha1.Ref) *EksaDiagnosticBundle {
	e.Bundle.Spec.Analyzers = append(e.Bundle.Spec.Analyzers, e.AnalyzerFactory.DataCenterConfigAnalyzers(config)...)
	return e
}

func (e *EksaDiagnosticBundle) WithOidcConfig(config *v1alpha1.OIDCConfig) *EksaDiagnosticBundle {
	if config != nil {
		e.Bundle.Spec.Analyzers = append(e.Bundle.Spec.Analyzers, e.AnalyzerFactory.EksaOidcAnalyzers()...)
	}
	return e
}

func (e *EksaDiagnosticBundle) WithExternalEtcd(config *v1alpha1.ExternalEtcdConfiguration) *EksaDiagnosticBundle {
	if config != nil {
		e.Bundle.Spec.Analyzers = append(e.Bundle.Spec.Analyzers, e.AnalyzerFactory.EksaExternalEtcdAnalyzers()...)
	}
	return e
}

func (e *EksaDiagnosticBundle) WithGitOpsConfig(config *v1alpha1.GitOpsConfig) *EksaDiagnosticBundle {
	if config != nil {
		e.Bundle.Spec.Analyzers = append(e.Bundle.Spec.Analyzers, e.AnalyzerFactory.EksaGitopsAnalyzers()...)
	}
	return e
}

func showAnalyzeResults(analyzeResults []*analyzerunner.AnalyzeResult) {
	results := "\n Analyze Results" + resultsSeparator
	for _, analyzeResult := range analyzeResults {
		result := ""
		if analyzeResult.IsPass {
			result = "Check PASS\n"
		} else if analyzeResult.IsFail {
			result = "Check FAIL\n"
		}
		result = result + fmt.Sprintf("Title: %s\n", analyzeResult.Title)
		result = result + fmt.Sprintf("Message: %s\n", analyzeResult.Message)
		if analyzeResult.URI != "" {
			result = result + fmt.Sprintf("URI: %s\n", analyzeResult.URI)
		}
		result = result + resultsSeparator
		results = results + result
	}
	logger.Info(results)
}
