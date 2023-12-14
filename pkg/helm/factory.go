package helm

import (
	"context"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	configcli "github.com/aws/eks-anywhere/pkg/config"
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

// ExecuteableClient represents a Helm client.
type ExecuteableClient interface {
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

// RegistryClient represents a Helm client.
type RegistryClient interface {
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
	BuildHelmExecutable(...Opt) ExecuteableClient
}

// ClientFactory provides a helm client for a cluster.
type ClientFactory struct {
	client     kubernetes.Client
	helmClient RegistryClient
	mu         sync.Mutex
	builder    ExecutableBuilder
	helmOpts   []Opt
}

// NewClientFactory returns a new HelmFactory.
func NewClientFactory(client kubernetes.Client, builder ExecutableBuilder, helmOpts ...Opt) *ClientFactory {
	hf := &ClientFactory{
		client:   client,
		builder:  builder,
		mu:       sync.Mutex{},
		helmOpts: helmOpts,
	}
	return hf
}

func (f *ClientFactory) withRegistryMirror(r *registrymirror.RegistryMirror) Opt {
	return func(ho *Config) {
		ho.RegistryMirror = r
	}
}

// buildClient returns a new Helm executeble.
func (f *ClientFactory) buildClient(opts ...Opt) ExecuteableClient {
	opts = append(f.helmOpts, opts...)
	return f.builder.BuildHelmExecutable(opts...)
}

// GetClientForCluster returns a new Helm client configured using information from the provided cluster's management cluster.
func (f *ClientFactory) GetClientForCluster(ctx context.Context, clus *anywherev1.Cluster) (RegistryClient, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	managmentCluster := clus

	var rUsername, rPassword string
	var err error

	if clus.IsManaged() {
		managmentCluster = &anywherev1.Cluster{}
		if err := f.client.Get(ctx, clus.ManagedBy(), clus.Namespace, managmentCluster); err != nil {
			return nil, err
		}
	}

	if managmentCluster.RegistryAuth() {
		rUsername, rPassword, err = f.getClusterRegistryCredentrials(ctx, managmentCluster)
		if err != nil {
			return nil, err
		}
	}

	r := registrymirror.FromCluster(managmentCluster)
	f.helmClient = f.buildClient(f.withRegistryMirror(r))

	if managmentCluster.RegistryAuth() {
		if err := f.helmClient.RegistryLogin(ctx, r.BaseRegistry, rUsername, rPassword); err != nil {
			return nil, err
		}
	}

	return f.helmClient, nil
}

// getClusterRegistryCredentrials retrieves the regitry mirror credentials for the management cluster.
// Registry credentials may not be found by retrieving on the cluster, this can happen on Cluster creation with the CLI.
// For now, to handle this, we fallback to reading the credentials from the environment variables.
func (f *ClientFactory) getClusterRegistryCredentrials(ctx context.Context, cluster *anywherev1.Cluster) (string, string, error) {
	var rUsername, rPassword string
	var err error

	rUsername, rPassword, err = configcli.ReadCredentialsFromSecret(ctx, f.client)
	if err != nil && !apierrors.IsNotFound(err) {
		return "", "", err
	}

	if apierrors.IsNotFound(err) {
		rUsername, rPassword, err = configcli.ReadCredentials()
		if err != nil {
			return "", "", err
		}
	}

	return rUsername, rPassword, nil
}
