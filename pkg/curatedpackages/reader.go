package curatedpackages

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/registry"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// Temporary: Curated packages dev and prod accounts are currently hard coded
// This is because there is no mechanism to extract these values as of now.
const (
	prodAccount       = "eks-anywhere"
	devAccount        = "l0g8r8j6"
	stagingAccount    = "w9m0f3l5"
	publicProdECR     = "public.ecr.aws/" + prodAccount
	publicDevECR      = "public.ecr.aws/" + devAccount
	publicStagingECR  = "public.ecr.aws/" + stagingAccount
	packageProdDomain = "783794618700.dkr.ecr.us-west-2.amazonaws.com"
	packageDevDomain  = "857151390494.dkr.ecr.us-west-2.amazonaws.com"
)

type PackageReader struct {
	cache           *registry.Cache
	credentialStore *registry.CredentialStore
	awsRegion       string
}

// NewPackageReader create a new package reader with storage client.
func NewPackageReader(cache *registry.Cache, credentialStore *registry.CredentialStore, awsRegion string) *PackageReader {
	if len(awsRegion) <= 0 {
		awsRegion = eksaDefaultRegion
	}
	return &PackageReader{
		cache:           cache,
		credentialStore: credentialStore,
		awsRegion:       awsRegion,
	}
}

// ReadImagesFromBundles and return a list of image artifacts.
func (r *PackageReader) ReadImagesFromBundles(ctx context.Context, b *releasev1.Bundles) ([]registry.Artifact, error) {
	var err error
	var images []registry.Artifact
	for _, vb := range b.Spec.VersionsBundles {
		bundleURI, bundle, err := r.getBundle(ctx, vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		packageImages := r.fetchImagesFromBundle(bundleURI, bundle)
		images = append(images, packageImages...)
	}

	return removeDuplicateImages(images), err
}

// ReadChartsFromBundles and return a list of chart artifacts.
func (r *PackageReader) ReadChartsFromBundles(ctx context.Context, b *releasev1.Bundles) []registry.Artifact {
	var images []registry.Artifact
	for _, vb := range b.Spec.VersionsBundles {
		bundleURI, bundle, err := r.getBundle(ctx, vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		bundleArtifact := registry.NewArtifactFromURI(bundleURI)
		images = append(images, bundleArtifact)
		packagesHelmChart := r.fetchPackagesHelmChart(bundleURI, bundle)
		images = append(images, packagesHelmChart...)
	}
	return removeDuplicateImages(images)
}

func (r *PackageReader) getBundle(ctx context.Context, vb releasev1.VersionsBundle) (string, *packagesv1.PackageBundle, error) {
	bundleURI, err := GetPackageBundleRef(vb)
	if err != nil {
		return "", nil, err
	}

	artifact := registry.NewArtifactFromURI(bundleURI)
	sc, err := r.cache.Get(registry.NewStorageContext(artifact.Registry, r.credentialStore, nil, false))
	if err != nil {
		return "", nil, err
	}

	data, err := registry.PullBytes(ctx, sc, artifact)
	if err != nil {
		return "", nil, err
	}
	bundle := packagesv1.PackageBundle{}
	err = yaml.Unmarshal(data, &bundle)
	if err != nil {
		return "", nil, err
	}
	return artifact.VersionedImage(), &bundle, nil
}

func (r *PackageReader) fetchPackagesHelmChart(bundleURI string, bundle *packagesv1.PackageBundle) []registry.Artifact {
	images := make([]registry.Artifact, 0, len(bundle.Spec.Packages))
	bundleRegistry := getChartRegistry(bundleURI)
	for _, p := range bundle.Spec.Packages {
		chartURI := fmt.Sprintf("%s/%s@%s", bundleRegistry, p.Source.Repository, p.Source.Versions[0].Digest)
		pHC := registry.NewArtifactFromURI(chartURI)
		pHC.Tag = p.Source.Versions[0].Name
		images = append(images, pHC)
	}
	return images
}

func (r *PackageReader) fetchImagesFromBundle(bundleURI string, bundle *packagesv1.PackageBundle) []registry.Artifact {
	images := make([]registry.Artifact, 0, len(bundle.Spec.Packages))
	bundleRegistry := getImageRegistry(bundleURI, r.awsRegion)
	for _, p := range bundle.Spec.Packages {
		// each package will have at least one version
		for _, version := range p.Source.Versions[0].Images {
			imageURI := fmt.Sprintf("%s/%s@%s", bundleRegistry, version.Repository, version.Digest)
			image := registry.NewArtifactFromURI(imageURI)
			image.Tag = "" // We do not have the tag right now
			images = append(images, image)
		}
	}
	return images
}

func removeDuplicateImages(images []registry.Artifact) []registry.Artifact {
	uniqueImages := make(map[string]struct{})
	var list []registry.Artifact
	for _, item := range images {
		if _, value := uniqueImages[item.VersionedImage()]; !value {
			uniqueImages[item.VersionedImage()] = struct{}{}
			list = append(list, item)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].VersionedImage() < list[j].VersionedImage()
	})
	return list
}

func getChartRegistry(uri string) string {
	if strings.Contains(uri, publicProdECR) {
		return publicProdECR
	}
	return publicDevECR
}

func getImageRegistry(uri, awsRegion string) string {
	if strings.Contains(uri, publicProdECR) {
		return strings.ReplaceAll(packageProdDomain, eksaDefaultRegion, awsRegion)
	}
	return strings.ReplaceAll(packageDevDomain, eksaDefaultRegion, awsRegion)
}
