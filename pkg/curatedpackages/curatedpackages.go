package curatedpackages

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
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

func GetVersionBundle(reader Reader, eksaVersion, kubeVersion string) (*releasev1.VersionsBundle, error) {
	b, err := reader.ReadBundlesForVersion(eksaVersion)
	if err != nil {
		return nil, err
	}
	versionsBundle := bundles.VersionsBundleForKubernetesVersion(b, kubeVersion)
	if versionsBundle == nil {
		return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", kubeVersion, b.Spec.Number)
	}
	return versionsBundle, nil
}

func NewDependenciesForPackages(ctx context.Context, paths ...string) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		WithExecutableMountDirs(paths...).
		WithExecutableBuilder().
		WithManifestReader().
		WithKubectl().
		WithHelm().
		Build(ctx)
}
