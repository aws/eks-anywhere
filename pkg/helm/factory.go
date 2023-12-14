package helm

import (
	"context"
	"reflect"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	configcli "github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

// Config contains configuration options for Helm.
type Config struct {
	RegistryMirror *registrymirror.RegistryMirror
	Env            map[string]string
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

// WithEnv joins the default and the provided maps together.
func WithEnv(env map[string]string) Opt {
	return func(c *Config) {
		for k, v := range env {
			c.Env[k] = v
		}
	}
}

// ExecutableBuilder builds the Helm executable and returns it.
type ExecutableBuilder interface {
	BuildHelmExecutable(...Opt) Client
}

// ClientFactory provides a helm client for a cluster.
type ClientFactory struct {
	client         client.Client
	helmClient     Client
	mu             sync.Mutex
	builder        ExecutableBuilder
	registryMirror *registrymirror.RegistryMirror
}

// NewClientForClusterFactory returns a new helm ClientFactory.
func NewClientForClusterFactory(client client.Client, builder ExecutableBuilder) *ClientFactory {
	hf := &ClientFactory{
		client:  client,
		builder: builder,
		mu:      sync.Mutex{},
	}
	return hf
}

// Get returns a new Helm client configured using information from the provided cluster's management cluster.
func (f *ClientFactory) Get(ctx context.Context, clus *anywherev1.Cluster) (Client, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

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

	// There is no need to rebuild the client everytime,
	// here we check if the registry has changed and rebuild the helm client in that case
	// Rebuild the Helm client only if the registry has changed
	if f.helmClient == nil || !reflect.DeepEqual(f.registryMirror, r) {
		f.helmClient = f.builder.BuildHelmExecutable(WithRegistryMirror(r), WithInsecure())
		f.registryMirror = r
	}

	if r != nil && managmentCluster.RegistryAuth() {
		if err := f.helmClient.RegistryLogin(ctx, r.BaseRegistry, rUsername, rPassword); err != nil {
			return nil, err
		}
	}

	return f.helmClient, nil
}
