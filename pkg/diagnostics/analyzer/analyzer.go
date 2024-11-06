package analyzer

import (
	"context"

	"github.com/pkg/errors"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// Analyzer checks the health of all clusters in a management cluster.
type Analyzer struct {
	readers readers
}

// ClusterAnalysisResult contains the result of an analysis of a cluster.
type ClusterAnalysisResult struct {
	Cluster  *anywherev1.Cluster
	Findings []Finding
}

// New creates a new Analyzer.
func New(client kubernetes.Reader, podLogs PodLogsReader) *Analyzer {
	return &Analyzer{
		readers: readers{
			client:  client,
			podLogs: podLogs,
		},
	}
}

// LogFilter allows to filter logs from a reader.
type LogFilter func(string) bool

// PodLogsReader accesses logs from pods in a cluster.
type PodLogsReader interface {
	LogsFromDeployment(name, namespace string, filters ...LogFilter) ([]string, error)
}

// AnalyzeAll checks the health and tries to find the cause of issues in all clusters.
func (a *Analyzer) AnalyzeAll(ctx context.Context) ([]ClusterAnalysisResult, error) {
	clusters := &anywherev1.ClusterList{}
	if err := a.readers.client.List(ctx, clusters); err != nil {
		return nil, errors.Wrapf(err, "listing all clusters for analysis")
	}

	var results []ClusterAnalysisResult

	for _, cluster := range clusters.Items {
		result, err := a.AnalyzeCluster(ctx, &cluster)
		if err != nil {
			return nil, errors.Wrapf(err, "analyzing cluster %s", cluster.Name)
		}

		results = append(results, *result)
	}

	return results, nil
}

// AnalyzeCluster checks the health and tries to find the cause of issues in a cluster.
func (a *Analyzer) AnalyzeCluster(ctx context.Context, cluster *anywherev1.Cluster) (*ClusterAnalysisResult, error) {
	r := &ClusterAnalysisResult{
		Cluster: cluster,
	}

	if isTrue(condition(cluster, anywherev1.ReadyCondition)) {
		return r, nil
	}

	if !isTrue(condition(cluster, anywherev1.ControlPlaneReadyCondition)) {
		f, err := a.analyzeControlPlane(ctx, cluster)
		if err != nil {
			return nil, errors.Wrapf(err, "analyzing control plane for cluster %s", cluster.Name)
		}
		r.Findings = append(r.Findings, f...)
	}

	return r, nil
}

type analyzeResults struct {
	Finding       Finding
	nextAnalyzers []analyzer
}

type readers struct {
	client  kubernetes.Reader
	podLogs PodLogsReader
}

type analyzer interface {
	analyze(context.Context, readers) (*analyzeResults, error)
}

func run(ctx context.Context, readers readers, analyzer analyzer) (*Finding, error) {
	results, err := analyzer.analyze(ctx, readers)
	if err != nil {
		return nil, err
	}

	if results == nil {
		return nil, nil
	}

	for _, a := range results.nextAnalyzers {
		f, err := run(ctx, readers, a)
		if err != nil {
			return nil, err
		}
		if f == nil {
			continue
		}

		results.Finding.Findings = append(results.Finding.Findings, *f)
	}

	return &results.Finding, nil
}
