package diagnostics

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type EksaDiagnosticBundle struct {
	bundle           *supportBundle
	bundlePath       string
	client           BundleClient
	collectorFactory CollectorFactory
	clusterSpec      *cluster.Spec
	analyzerFactory  AnalyzerFactory
	kubeconfig       string
	kubectl          *executables.Kubectl
	retrier          *retrier.Retrier
	writer           filewriter.FileWriter
}

func (e *EksaDiagnosticBundle) CollectAndAnalyze(ctx context.Context, sinceTimeValue *time.Time) error {
	e.createDiagnosticNamespaceAndRoles(ctx)

	logger.Info("collecting support bundle, this can take a while", "bundle", e.bundlePath, "since", sinceTimeValue, "kubeconfig", e.kubeconfig)
	archivePath, err := e.client.Collect(ctx, e.bundlePath, sinceTimeValue, e.kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to Collect support bundle: %v", err)
	}

	logger.Info("analyzing support bundle", "bundle", e.bundlePath, "archive", archivePath)
	analysis, err := e.client.Analyze(ctx, e.bundlePath, archivePath)
	if err != nil {
		return fmt.Errorf("error when analyzing bundle: %v", err)
	}

	yamlAnalysis, err := yaml.Marshal(analysis)
	if err != nil {
		return fmt.Errorf("error while analyzing bundle: %v", err)
	}

	fmt.Println(string(yamlAnalysis))
	logger.Info("Support bundle archive created", "archivePath", archivePath)

	e.deleteDiagnosticNamespaceAndRoles(ctx)
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
	if err != nil {
		return err
	}
	logger.V(3).Info("bundle config written", "path", e.bundlePath)
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

func (e *EksaDiagnosticBundle) WithMachineConfigs(configs []providers.MachineConfig) *EksaDiagnosticBundle {
	e.bundle.Spec.Collectors = append(e.bundle.Spec.Collectors, e.collectorFactory.EksaHostCollectors(configs)...)
	return e
}

func (e *EksaDiagnosticBundle) WithLogTextAnalyzers() *EksaDiagnosticBundle {
	e.bundle.Spec.Analyzers = append(e.bundle.Spec.Analyzers, e.analyzerFactory.EksaLogTextAnalyzers(e.bundle.Spec.Collectors)...)
	return e
}

// createDiagnosticNamespace attempts to create the namespace eksa-diagnostics and associated RBAC objects.
// collector pods, for example host log collectors or run command collectors, will be launched in this namespace with the default service account.
// this method intentionally does not return an error
// a cluster in need of diagnosis may be unable to create new API objects and we should not stop our collection/analysis just because the namespace fails to create
func (e *EksaDiagnosticBundle) createDiagnosticNamespaceAndRoles(ctx context.Context) {
	targetCluster := &types.Cluster{
		KubeconfigFile: e.kubeconfig,
	}

	logger.V(1).Info("creating temporary namespace for diagnostic collector", "namespace", constants.EksaDiagnosticsNamespace)
	err := e.retrier.Retry(
		func() error {
			return e.kubectl.CreateNamespace(ctx, e.kubeconfig, constants.EksaDiagnosticsNamespace)
		},
	)
	if err != nil {
		logger.Info("WARNING: failed to create eksa-diagnostics namespace. Some collectors may fail to run.", "err", err)
	}

	logger.V(1).Info("creating temporary ClusterRole and RoleBinding for diagnostic collector")
	err = e.retrier.Retry(
		func() error {
			return e.kubectl.ApplyKubeSpecFromBytes(ctx, targetCluster, diagnosticCollectorRbac)
		},
	)
	if err != nil {
		logger.Info("WARNING: failed to create roles for eksa-diagnostic-collector. Some collectors may fail to run.", "err", err)
	}
}

func (e *EksaDiagnosticBundle) deleteDiagnosticNamespaceAndRoles(ctx context.Context) {
	targetCluster := &types.Cluster{
		KubeconfigFile: e.kubeconfig,
	}

	logger.V(1).Info("cleaning up temporary roles for diagnostic collectors")
	err := e.retrier.Retry(
		func() error {
			return e.kubectl.DeleteKubeSpecFromBytes(ctx, targetCluster, diagnosticCollectorRbac)
		},
	)
	if err != nil {
		logger.Info("WARNING: failed to clean up roles for eksa-diagnostics.", "err", err)
	}

	logger.V(1).Info("cleaning up temporary namespace  for diagnostic collectors", "namespace", constants.EksaDiagnosticsNamespace)
	err = e.retrier.Retry(
		func() error {
			return e.kubectl.DeleteNamespace(ctx, e.kubeconfig, constants.EksaDiagnosticsNamespace)
		},
	)
	if err != nil {
		logger.Info("WARNING: failed to clean up eksa-diagnostics namespace.", "err", err, "namespace", constants.EksaDiagnosticsNamespace)
	}
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
