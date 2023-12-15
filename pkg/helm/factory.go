package helm

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	configcli "github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

// Config contains configuration options for Helm.
type Config struct {
	RegistryMirror *registrymirror.RegistryMirror
	ProxyConfig    map[string]string
	Insecure       bool
}

// Opt is a functional option for configuring Helm behavior.
type Opt func(*Config)

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

// WithRegistryMirror sets up registry mirror for helm.
func WithRegistryMirror(mirror *registrymirror.RegistryMirror) Opt {
	return func(c *Config) {
		c.RegistryMirror = mirror
	}
}

// WithInsecure configures helm to skip validating TLS certificates when
// communicating with the Kubernetes API.
func WithInsecure() Opt {
	return func(c *Config) {
		c.Insecure = true
	}
}

// WithProxyConfig sets the proxy configurations on the Config.
// proxyConfig contains configurations for the proxy settings used by the Helm client.
// The valid keys are as follows:
// "HTTPS_PROXY": Specifies the HTTPS proxy URL.
// "HTTP_PROXY": Specifies the HTTP proxy URL.
// "NO_PROXY": Specifies a comma-separated a list of destination domain names, domains, IP addresses, or other network CIDRs to exclude proxying.
func WithProxyConfig(proxyConfig map[string]string) Opt {
	return func(c *Config) {
		c.ProxyConfig = proxyConfig
	}
}

// ExecutableBuilder builds the Helm executable and returns it.
type ExecutableBuilder interface {
	BuildHelmExecutable(...Opt) Client
}

// ClientFactory provides a helm client for a cluster.
type ClientFactory struct {
	client     client.Client
	helmClient Client
	builder    ExecutableBuilder
}

// NewClientForClusterFactory returns a new helm ClientFactory.
func NewClientForClusterFactory(client client.Client, builder ExecutableBuilder) *ClientFactory {
	hf := &ClientFactory{
		client:  client,
		builder: builder,
	}
	return hf
}

// Get returns a new Helm client configured using information from the provided cluster's management cluster.
func (f *ClientFactory) Get(ctx context.Context, clus *anywherev1.Cluster) (Client, error) {
	managmentCluster := clus

	var err error
	if clus.IsManaged() {
		managmentCluster, err = clusters.FetchManagementEksaCluster(ctx, f.client, clus)
		if err != nil {
			return nil, err
		}
	}

	var rUsername, rPassword string
	if managmentCluster.RegistryAuth() {
		rUsername, rPassword, err = configcli.ReadCredentialsFromSecret(ctx, f.client)
		if err != nil {
			return nil, err
		}
	}

	r := registrymirror.FromCluster(managmentCluster)
	f.helmClient = f.builder.BuildHelmExecutable(WithRegistryMirror(r), WithInsecure())

	if r != nil && managmentCluster.RegistryAuth() {
		if err := f.helmClient.RegistryLogin(ctx, r.BaseRegistry, rUsername, rPassword); err != nil {
			return nil, err
		}
	}

	return f.helmClient, nil
}
