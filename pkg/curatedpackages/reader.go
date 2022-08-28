package curatedpackages

import (
	"context"
	"fmt"

	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type PackageReader struct {
	*manifests.Reader
	readerType *ReaderType
}

func NewPackageReader(mr *manifests.Reader, readerType *ReaderType) *PackageReader {
	return &PackageReader{
		Reader:     mr,
		readerType: readerType,
	}
}

func (r *PackageReader) ReadImagesFromBundles(ctx context.Context, b *releasev1.Bundles) ([]releasev1.Image, error) {
	images, err := r.Reader.ReadImagesFromBundles(ctx, b)

	for _, vb := range b.Spec.VersionsBundles {
		artifact, err := GetPackageBundleRef(vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		packagesHelmChart, err := r.fetchPackagesImage(ctx, vb, artifact)
		if err != nil {
			logger.Info("Warning: Failed extracting packages", "error", err)
			continue
		}
		images = append(images, packagesHelmChart...)
	}

	return images, err
}

func (r *PackageReader) ReadChartsFromBundles(ctx context.Context, b *releasev1.Bundles) []releasev1.Image {
	images := r.Reader.ReadChartsFromBundles(ctx, b)
	for _, vb := range b.Spec.VersionsBundles {
		artifact, err := GetPackageBundleRef(vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		packagesHelmChart, err := fetchPackagesHelmChart(ctx, vb, artifact)
		if err != nil {
			logger.Info("Warning: Failed extracting packages", "error", err)
			continue
		}
		images = append(images, packagesHelmChart...)
	}
	return images
}

func fetchPackagesHelmChart(ctx context.Context, versionsBundle releasev1.VersionsBundle, artifact string) ([]releasev1.Image, error) {
	data, err := PullLatestBundle(ctx, artifact)
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
			URI:         fmt.Sprintf("%s/%s:%s", GetRegistry(ctrl.URI), p.Source.Repository, p.Source.Versions[0].Name),
		}
		images = append(images, pHC)
	}
	return images, nil
}

func (r *PackageReader) fetchPackagesImage(ctx context.Context, versionsBundle releasev1.VersionsBundle, artifact string) ([]releasev1.Image, error) {
	data, err := PullLatestBundle(ctx, artifact)
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
		for _, version := range p.Source.Versions[0].Images {
			image := releasev1.Image{
				Name:        version.Repository,
				Description: version.Repository,
				OS:          ctrl.OS,
				OSName:      ctrl.OSName,
				URI:         fmt.Sprintf("%s/%s@%s", r.readerType.GetRegistry(ctrl.URI), version.Repository, version.Digest),
			}
			logger.Info("Package Reader is: " + image.URI)
			images = append(images, image)
		}
	}
	return images, nil
}
