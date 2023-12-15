package helm

import "context"

// Client represents a Helm client.
type Client interface {
	PushChart(ctx context.Context, chart, registry string) error
	PullChart(ctx context.Context, ociURI, version string) error
	ListCharts(ctx context.Context, kubeconfigFilePath string) ([]string, error)
	SaveChart(ctx context.Context, ociURI, version, folder string) error
	Delete(ctx context.Context, kubeconfigFilePath, installName, namespace string) error
	UpgradeChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, valuesFilePath string, opts ...Opt) error
	InstallChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, valuesFilePath string) error
	InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, namespace, valueFilePath string, skipCRDs bool, values []string) error
	Template(ctx context.Context, ociURI, version, namespace string, values interface{}, kubeVersion string) ([]byte, error)
	RegistryLogin(ctx context.Context, registry, username, password string) error
}
