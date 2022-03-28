package curatedpackages

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/version"
)

func NewRegistry(ctx context.Context, deps *dependencies.Dependencies, registry, kubeVersion string) (BundleRegistry, error) {
	if registry != "" {
		registryUsername := os.Getenv("REGISTRY_USERNAME")
		registryPassword := os.Getenv("REGISTRY_PASSWORD")
		if registryUsername == "" || registryPassword == "" {
			return nil, fmt.Errorf("username or password not set. Provide REGISTRY_USERNAME and REGISTRY_PASSWORD when using custom registry")
		}
		registry := NewCustomRegistry(
			deps.Helm,
			registry,
			registryUsername,
			registryPassword,
		)
		err := registry.Login(ctx)
		if err != nil {
			return nil, err
		}
		return registry, nil
	}
	defaultRegistry := NewDefaultRegistry(
		deps.ManifestReader,
		kubeVersion,
		version.Get(),
	)
	return defaultRegistry, nil
}

func CreateBundleManager(kubeVersion string) bundle.Manager {
	versionSplit := strings.Split(kubeVersion, ".")
	if len(versionSplit) != 2 {
		return nil
	}
	major, minor := versionSplit[0], versionSplit[1]
	log := logr.Discard()
	discovery := NewDiscovery(major, minor)
	puller := artifacts.NewRegistryPuller()
	return bundle.NewBundleManager(log, discovery, puller)
}
