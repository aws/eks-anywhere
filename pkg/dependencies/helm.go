package dependencies

import (
	"context"
	"sync"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

type ExecutableBuilder interface {
	BuildHelmExecutable(...executables.HelmOpt) *executables.Helm
}

// HelmFactory is responsible for creating and owning instances of Helm client.
type HelmFactory struct {
	mu                 sync.Mutex
	builder            ExecutableBuilder
	helm               *executables.Helm
	registryMirror     *registrymirror.RegistryMirror
	proxyConfiguration map[string]string
	insecure           bool
}

// WithRegistryMirror configures the factory to use registry mirror wherever applicable.
func (f *HelmFactory) WithRegistryMirror(registryMirror *registrymirror.RegistryMirror) *HelmFactory {
	f.registryMirror = registryMirror

	return f
}

// WithProxyConfigurations configures the factory to use proxy configurations wherever applicable.
func (f *HelmFactory) WithProxyConfigurations(proxyConfiguration map[string]string) *HelmFactory {
	f.proxyConfiguration = proxyConfiguration

	return f
}

// WithInsecure configures the factory to configure helm to use to allow connections to TLS registry without certs or with self-signed certs
func (f *HelmFactory) WithInsecure() *HelmFactory {
	f.insecure = true

	return f
}

func NewHelmFactory(builder ExecutableBuilder) *HelmFactory {
	return &HelmFactory{
		builder: builder,
		mu:      sync.Mutex{},
	}

}

// GetClient returns a new Helm executeble client.
func (f *HelmFactory) GetClient(opts ...executables.HelmOpt) *executables.Helm {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.registryMirror != nil {
		opts = append(opts, executables.WithRegistryMirror(f.registryMirror))
	}

	if f.proxyConfiguration != nil {
		opts = append(opts, executables.WithEnv(f.proxyConfiguration))
	}

	if f.insecure {
		opts = append(opts, executables.WithInsecure())
	}

	f.helm = f.builder.BuildHelmExecutable(opts...)
	return f.helm
}

// GetClientForCluster returns a new helm client.
// There is no cluster information that needs to be passed, but this method was needed to satisfy
// an interface together with the controller helm factory.
func (f *HelmFactory) GetClientForCluster(_ context.Context, _ string) (*executables.Helm, error) {
	return f.GetClient(), nil
}
