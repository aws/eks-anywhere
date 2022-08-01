package curatedpackages

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/version"
)

type DefaultRegistry struct {
	releaseManifestReader Reader
	kubeVersion           string
	cliVersion            version.Info
}

func NewDefaultRegistry(rmr Reader, kv string, cv version.Info) *DefaultRegistry {
	return &DefaultRegistry{
		releaseManifestReader: rmr,
		kubeVersion:           kv,
		cliVersion:            cv,
	}
}

func (dr *DefaultRegistry) GetRegistryBaseRef(ctx context.Context) (string, error) {
	release, err := dr.releaseManifestReader.ReadBundlesForVersion(dr.cliVersion.GitVersion)
	if err != nil {
		return "", fmt.Errorf("unable to parse the release manifest %v", err)
	}
	versionsBundle := bundles.VersionsBundleForKubernetesVersion(release, dr.kubeVersion)
	if versionsBundle == nil {
		return "", fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", dr.kubeVersion, release.Spec.Number)
	}
	packageController := versionsBundle.PackageController

	// Use package controller registry to fetch packageBundles.
	// Format of controller image is: <uri>/<env_type>/<repository_name>
	registry := GetRegistry(packageController.Controller.Image())
	registryBaseRef := fmt.Sprintf("%s/%s", registry, ImageRepositoryName)
	return registryBaseRef, nil
}
