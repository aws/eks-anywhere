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
	publicProdECR     = "public.ecr.aws/eks-anywhere"
	packageProdDomain = "783794618700.dkr.ecr.us-west-2.amazonaws.com"
	packageDevDomain  = "857151390494.dkr.ecr.us-west-2.amazonaws.com"
)

type ManifestReader interface {
	ReadBundlesForVersion(eksaVersion string) (*releasev1.Bundles, error)
	ReadImagesFromBundles(ctx context.Context, bundles *releasev1.Bundles) ([]releasev1.Image, error)
	ReadChartsFromBundles(ctx context.Context, bundles *releasev1.Bundles) []releasev1.Image
}

type PackageReader struct {
	ManifestReader
	cache *registry.Cache
}


// NewPackageReader create a new package reader with storage client.
func NewPackageReader(mr ManifestReader, cache *registry.Cache) *PackageReader {
	return &PackageReader{
		ManifestReader: mr,
		cache:          cache,
	}
}

func (r *PackageReader) ReadImagesFromBundles(ctx context.Context, b *releasev1.Bundles) ([]releasev1.Image, error) {
	var err error
	var images []releasev1.Image
	for _, vb := range b.Spec.VersionsBundles {
		bundleURI, bundle, err := r.getBundle(ctx, vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		images = append(images, releasev1.Image{URI: bundleURI})
		packageImages := r.fetchImagesFromBundle(bundleURI, bundle)
		images = append(images, packageImages...)
	}

	return removeDuplicateImages(images), err
}

func (r *PackageReader) ReadChartsFromBundles(ctx context.Context, b *releasev1.Bundles) []releasev1.Image {
	var images []releasev1.Image
	for _, vb := range b.Spec.VersionsBundles {
		bundleRegistry, bundle, err := r.getBundle(ctx, vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		images = append(images, releasev1.Image{URI: bundleRegistry})
		packagesHelmChart := r.fetchPackagesHelmChart(bundleRegistry, bundle)
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
	sc, err := r.cache.Get(registry.NewDefaultStorageContext(artifact.Registry))
	if err != nil {
		return "", nil, err
	}

	data, err := sc.PullBytes(ctx, artifact)
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

func (r *PackageReader) fetchPackagesHelmChart(bundleURI string, bundle *packagesv1.PackageBundle) []releasev1.Image {
	images := make([]releasev1.Image, 0, len(bundle.Spec.Packages))
	for _, p := range bundle.Spec.Packages {
		pHC := releasev1.Image{
			Name:        p.Name,
			Description: p.Name,
			URI:         fmt.Sprintf("%s/%s@%s", registry.NewArtifactFromURI(bundleURI).Registry, p.Source.Repository, p.Source.Versions[0].Digest),
			ImageDigest: p.Source.Versions[0].Digest,
		}
		images = append(images, pHC)
	}
	return images
}

func (r *PackageReader) fetchImagesFromBundle(bundleURI string, bundle *packagesv1.PackageBundle) []releasev1.Image {
	images := make([]releasev1.Image, 0, len(bundle.Spec.Packages))
	for _, p := range bundle.Spec.Packages {
		// each package will have at least one version
		for _, version := range p.Source.Versions[0].Images {
			image := releasev1.Image{
				Name:        version.Repository,
				Description: version.Repository,
				URI:         fmt.Sprintf("%s/%s@%s", getRegistry(bundleURI), version.Repository, version.Digest),
				ImageDigest: version.Digest,
			}
			images = append(images, image)
		}
	}
	return images
}

func removeDuplicateImages(images []releasev1.Image) []releasev1.Image {
	uniqueImages := make(map[string]struct{})
	var list []releasev1.Image
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

func getRegistry(uri string) string {
	if strings.Contains(uri, publicProdECR) {
		return packageProdDomain
	}
	return packageDevDomain
}
