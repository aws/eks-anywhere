package dependencies

import (
	"sync"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

type ExecutableBuilder interface {
	BuildHelmExecutable(...executables.HelmOpt) *executables.Helm
}

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
	}
}

func (f *HelmFactory) GetInstance(opts ...executables.HelmOpt) *executables.Helm {
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
