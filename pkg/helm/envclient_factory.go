package helm

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	configcli "github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

// EnvClientFactory provides a Helm client for a cluster.
type EnvClientFactory struct {
	helmClient Client
	builder    ClientBuilder
}

// NewEnvClientFactory returns a new EnvClientFactory.
func NewEnvClientFactory(builder ClientBuilder) *EnvClientFactory {
	return &EnvClientFactory{
		builder: builder,
	}
}

// buildClient returns a new helm executable client.
func (f *EnvClientFactory) buildClient(opts ...Opt) Client {
	return f.builder.BuildHelm(opts...)
}

// Get returns the helm registry client.
// The parameters here are not used and it only returns the client that is initialized using Init.
func (f *EnvClientFactory) Get(_ context.Context, _ *anywherev1.Cluster) (Client, error) {
	return f.helmClient, nil
}

// Init builds the helm registry client once using the registry mirror information from the cluster information.
// It should be called at least once first, before trying to retrieving and using the client using Get.
// It only builds the helm registry client once.
// This is not thread safe and the caller should guarantee that it does not get called from multiple threads.
func (f *EnvClientFactory) Init(ctx context.Context, r *registrymirror.RegistryMirror, opts ...Opt) error {
	if f.helmClient != nil {
		return nil
	}

	opts = append(opts, WithRegistryMirror(r))

	helmClient := f.buildClient(opts...)

	if r == nil || r != nil && !r.Auth {
		f.helmClient = helmClient
		return nil
	}

	// TODO (cxbrowne): The registry credentials should be injected on construction through environment variables REGISTRY_USERNAME
	// and REGISTRY_PASSWORD, or passed to this method as arguments.
	// Issue: https://github.com/aws/eks-anywhere-internal/issues/2115
	rUsername, rPassword, err := configcli.ReadCredentials()
	if err != nil {
		return err
	}

	if err := helmClient.RegistryLogin(ctx, r.BaseRegistry, rUsername, rPassword); err != nil {
		return err
	}

	f.helmClient = helmClient
	return nil
}
