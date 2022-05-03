package curatedpackages

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/version"
)

func NewRegistry(deps *dependencies.Dependencies, registryName, kubeVersion, username, password string) (BundleRegistry, error) {
	if registryName != "" {
		registry := NewCustomRegistry(
			deps.Helm,
			registryName,
			username,
			password,
		)
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
	k := NewKubeVersion(major, minor)
	discovery := NewDiscovery(k)
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
