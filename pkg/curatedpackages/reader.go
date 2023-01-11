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
	AllImages     bool
	storageClient registry.StorageClient
}

var _ ManifestReader = (*PackageReader)(nil)

func NewPackageReader(mr ManifestReader, sc registry.StorageClient) *PackageReader {
	return &PackageReader{
		ManifestReader: mr,
		storageClient:  sc,
		AllImages:      true,
	}
}

func (r *PackageReader) ReadImagesFromBundles(ctx context.Context, b *releasev1.Bundles) ([]releasev1.Image, error) {
	var err error
	var images []releasev1.Image
	if r.AllImages {
		images, err = r.ManifestReader.ReadImagesFromBundles(ctx, b)
	}

	for _, vb := range b.Spec.VersionsBundles {
		artifact, err := GetPackageBundleRef(vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		packageImages, err := r.fetchImagesFromBundle(ctx, vb, artifact)
		if err != nil {
			logger.Info("Warning: Failed extracting packages", "error", err)
			continue
		}
		images = append(images, packageImages...)
	}

	return removeDuplicateImages(images), err
}

func (r *PackageReader) ReadChartsFromBundles(ctx context.Context, b *releasev1.Bundles) []releasev1.Image {
	var images []releasev1.Image
	eksaCharts := append(images, r.ManifestReader.ReadChartsFromBundles(ctx, b)...)
	if r.AllImages {
		images = append(images, eksaCharts...)
	}

	// Generate bundle repository
	for _, chart := range eksaCharts {
		if strings.Contains(chart.Repository(), "eks-anywhere-packages") {
			var bundleRepository releasev1.Image
			chart.DeepCopyInto(&bundleRepository)
			bundleRepository.URI = strings.Replace(bundleRepository.URI, "eks-anywhere-packages", "eks-anywhere-packages-bundles", 1)
			bundleRepository.URI = strings.Replace(bundleRepository.URI, bundleRepository.Version(), "v1-24-latest", 1)
			bundleRepository.ImageDigest = ""
			images = append(images, bundleRepository)
		}
	}

	for _, vb := range b.Spec.VersionsBundles {
		artifact, err := GetPackageBundleRef(vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		packagesHelmChart, err := r.fetchPackagesHelmChart(ctx, vb, artifact)
		if err != nil {
			logger.Info("Warning: Failed extracting packages", "error", err)
			continue
		}
		images = append(images, packagesHelmChart...)
	}
	return removeDuplicateImages(images)
}

func (r *PackageReader) fetchPackagesHelmChart(ctx context.Context, versionsBundle releasev1.VersionsBundle, artifact string) ([]releasev1.Image, error) {
	data, err := r.storageClient.PullBytes(ctx, registry.NewArtifactFromURI(artifact))
	if err != nil {
		return nil, err
	}
	ctrl := versionsBundle.PackageController.Controller
	bundle := &packagesv1.PackageBundle{}
	err = yaml.Unmarshal(data, bundle)
	if err != nil {
		return nil, err
	}
	images := make([]releasev1.Image, 0, len(bundle.Spec.Packages))
	for _, p := range bundle.Spec.Packages {
		pHC := releasev1.Image{
			Name:        p.Name,
			Description: p.Name,
			OS:          ctrl.OS,
			OSName:      ctrl.OSName,
			URI:         fmt.Sprintf("%s/%s@%s", GetRegistry(ctrl.URI), p.Source.Repository, p.Source.Versions[0].Digest),
			ImageDigest: p.Source.Versions[0].Digest,
		}
		images = append(images, pHC)
	}
	return images, nil
}

func (r *PackageReader) fetchImagesFromBundle(ctx context.Context, versionsBundle releasev1.VersionsBundle, artifact string) ([]releasev1.Image, error) {
	data, err := r.storageClient.PullBytes(ctx, registry.NewArtifactFromURI(artifact))
	if err != nil {
		return nil, err
	}
	bundle := &packagesv1.PackageBundle{}
	err = yaml.Unmarshal(data, bundle)
	if err != nil {
		return nil, err
	}

	ctrl := versionsBundle.PackageController.Controller
	images := make([]releasev1.Image, 0, len(bundle.Spec.Packages))

	for _, p := range bundle.Spec.Packages {
		// each package will have at least one version
		for _, version := range p.Source.Versions[0].Images {
			image := releasev1.Image{
				Name:        version.Repository,
				Description: version.Repository,
				OS:          ctrl.OS,
				OSName:      ctrl.OSName,
				URI:         fmt.Sprintf("%s/%s@%s", getRegistry(ctrl.URI), version.Repository, version.Digest),
				ImageDigest: version.Digest,
			}
			images = append(images, image)
		}
	}
	return images, nil
}

func removeDuplicateImages(images []releasev1.Image) []releasev1.Image {
	uniqueImages := make(map[string]releasev1.Image)
	var list []releasev1.Image
	for _, item := range images {
		if _, value := uniqueImages[item.VersionedImage()]; !value {
			uniqueImages[item.VersionedImage()] = item
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
