package curatedpackages

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/version"
)

func NewRegistry(ctx context.Context, deps *dependencies.Dependencies, registry, kubeVersion string) (BundleRegistry, error) {
	if registry != "" {
		username, password, err := helm.ReadRegistryCredentials()
		if err != nil {
			return nil, err
		}
		registry := NewCustomRegistry(
			deps.Helm,
			registry,
			username,
			password,
		)
		err = registry.Login(ctx)
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
	major, minor, err := parseKubeVersion(kubeVersion)
	if err != nil {
		return nil
	}
	log := logr.Discard()
	discovery := NewDiscovery(major, minor)
	puller := artifacts.NewRegistryPuller()
	return bundle.NewBundleManager(log, discovery, puller)
}

func parseKubeVersion(kubeVersion string) (string, string, error) {
	versionSplit := strings.Split(kubeVersion, ".")
	if len(versionSplit) != 2 {
		return "", "", fmt.Errorf("invalid kubeversion")
	}
	major, minor := versionSplit[0], versionSplit[1]
	return major, minor, nil
}
