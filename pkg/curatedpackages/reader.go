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
}

func NewPackageReader(mr *manifests.Reader) *PackageReader {
	return &PackageReader{
		Reader: mr,
	}
}

func (r *PackageReader) ReadImagesFromBundles(b *releasev1.Bundles) ([]releasev1.Image, error) {
	return r.Reader.ReadImagesFromBundles(b)
}

func (r *PackageReader) ReadChartsFromBundles(ctx context.Context, b *releasev1.Bundles) []releasev1.Image {
	images := r.Reader.ReadChartsFromBundles(ctx, b)
	for _, vb := range b.Spec.VersionsBundles {
		artifact, err := GetPackageBundleRef(vb)
		if err != nil {
			logger.Info("Warning: Failed getting bundle reference", "error", err)
			continue
		}
		packages, err := fetchPackages(ctx, vb, artifact)
		if err != nil {
			logger.Info("Warning: Failed extracting packages", "error", err)
			continue
		}
		images = append(images, packages...)
	}
	return images
}

func fetchPackages(ctx context.Context, versionsBundle releasev1.VersionsBundle, artifact string) ([]releasev1.Image, error) {
	data, err := Pull(ctx, artifact)
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
		pI := releasev1.Image{
			Name:        p.Name,
			Description: p.Name,
			OS:          ctrl.OS,
			OSName:      ctrl.OSName,
			URI:         fmt.Sprintf("%s/%s:%s", getDefaultRegistry(ctrl), p.Source.Repository, p.Source.Versions[0].Name),
		}
		images = append(images, pI)
	}
	return images, nil
}

func getDefaultRegistry(ctrl releasev1.Image) string {
	registry := GetRegistry(ctrl.URI)
	return registry
}
