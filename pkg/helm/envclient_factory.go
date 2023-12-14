package helm

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	configcli "github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

// EnvClientFactory provides a Helm client for a cluster.
type EnvClientFactory struct {
	helmClient RegistryClient
	builder    ExecutableBuilder
}

// NewEnvClientFactory returns a new EnvClientFactory.
func NewEnvClientFactory(builder ExecutableBuilder) *EnvClientFactory {
	return &EnvClientFactory{
		builder: builder,
	}
}

// buildClient returns a new helm executable client.
func (f *EnvClientFactory) buildClient(opts ...Opt) ExecuteableClient {
	return f.builder.BuildHelmExecutable(opts...)
}

// Get returns the helm registry client.
// The parameters here are not used and it only returns the client that is initialized using Init.
func (f *EnvClientFactory) Get(_ context.Context, _ *anywherev1.Cluster) (RegistryClient, error) {
	return f.helmClient, nil
}

// Init builds the helm registry client once using the registry mirror information from the cluster information.
// It should be called at least once first, before trying to retrieving and using the client using Get.
// It only builds the helm registry client once.
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
