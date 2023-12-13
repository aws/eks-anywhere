package dependencies

import (
	"context"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	configcli "github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

// ExecutableBuilder builds the Helm executeble and returns it.
type ExecutableBuilder interface {
	BuildHelmExecutable(...executables.HelmOpt) *executables.Helm
}

// HelmFactoryOpt configures a HelmFactory.
type HelmFactoryOpt func(*HelmFactory)

// HelmFactory is responsible for creating and owning instances of Helm client.
type HelmFactory struct {
	client         cluster.Client
	helmClient     *cluster.HelmClient
	mu             sync.Mutex
	builder        ExecutableBuilder
	registryMirror *registrymirror.RegistryMirror
	env            map[string]string
	insecure       bool
}

// WithRegistryMirror configures the factory to use registry mirror wherever applicable.
func WithRegistryMirror(registryMirror *registrymirror.RegistryMirror) HelmFactoryOpt {
	return func(hf *HelmFactory) {
		hf.registryMirror = registryMirror
	}
}

// WithEnv configures the factory to use proxy configurations wherever applicable.
func WithEnv(env map[string]string) HelmFactoryOpt {
	return func(hf *HelmFactory) {
		hf.env = env
	}
}

// WithInsecure configures the factory to configure helm to use to allow connections to TLS registry without certs or with self-signed certs.
func WithInsecure() HelmFactoryOpt {
	return func(hf *HelmFactory) {
		hf.insecure = true
	}
}

// NewHelmFactory returns a new HelmFactory.
func NewHelmFactory(client cluster.Client, builder ExecutableBuilder, opts ...HelmFactoryOpt) *HelmFactory {
	hf := &HelmFactory{
		client:  client,
		builder: builder,
		mu:      sync.Mutex{},
	}

	for _, o := range opts {
		o(hf)
	}

	return hf
}

// GetClient returns a new Helm executeble.
func (f *HelmFactory) GetClient(opts ...executables.HelmOpt) cluster.Helm {
	defaultOpts := []executables.HelmOpt{}

	if f.registryMirror != nil {
		defaultOpts = append(defaultOpts, executables.WithRegistryMirror(f.registryMirror))
	}
	if f.env != nil {
		defaultOpts = append(defaultOpts, executables.WithEnv(f.env))
	}

	if f.insecure {
		defaultOpts = append(defaultOpts, executables.WithInsecure())
	}

	opts = append(defaultOpts, opts...)
	return f.builder.BuildHelmExecutable(opts...)
}

// GetClientForCluster returns a new Helm client configured using information from the provided cluster's management cluster.
func (f *HelmFactory) GetClientForCluster(ctx context.Context, clus *anywherev1.Cluster) (*cluster.HelmClient, error) {
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
	helm := f.GetClient(executables.WithRegistryMirror(r))

	f.helmClient = cluster.NewHelmClient(helm, managmentCluster, rUsername, rPassword)
	return f.helmClient, nil
}

// getClusterRegistryCredentrials retrieves the regitry mirror credentials for the management cluster.
// Registry credentials may not be found by retrieving on the cluster, this can happen on Cluster creation with the CLI.
// For now, to handle this, we fallback to reading the credentials from the environment variables.
func (f *HelmFactory) getClusterRegistryCredentrials(ctx context.Context, cluster *anywherev1.Cluster) (string, string, error) {
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
