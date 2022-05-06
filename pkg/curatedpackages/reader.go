package curatedpackages

import (
	"bytes"
	"context"
	"fmt"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere/pkg/manifests"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type PackageReader struct {
	ManifestReader *manifests.Reader
}

func NewPackageReader(mr *manifests.Reader) *PackageReader {
	return &PackageReader{
		ManifestReader: mr,
	}
}

func (r *PackageReader) ReadBundlesForVersion(version string) (*releasev1.Bundles, error) {
	return r.ManifestReader.ReadBundlesForVersion(version)
}

func (r *PackageReader) ReadEKSD(eksaVersion, kubeVersion string) (*eksdv1.Release, error) {
	return r.ManifestReader.ReadEKSD(eksaVersion, kubeVersion)
}

func (r *PackageReader) ReadImages(eksaVersion string) ([]releasev1.Image, error) {
	return r.ManifestReader.ReadImages(eksaVersion)
}

func (r *PackageReader) ReadImagesFromBundles(b *releasev1.Bundles) ([]releasev1.Image, error) {
	images, err := r.ManifestReader.ReadImagesFromBundles(b)
	for _, v := range b.Spec.VersionsBundles {
		images = append(images, v.PackageControllerImage()...)
	}
	return images, err
}

func (r *PackageReader) ReadCharts(eksaVersion string) ([]releasev1.Image, error) {
	return r.ManifestReader.ReadCharts(eksaVersion)
}

func (r *PackageReader) ReadChartsFromBundles(ctx context.Context, b *releasev1.Bundles) []releasev1.Image {
	images := r.ManifestReader.ReadChartsFromBundles(ctx, b)
	for _, vb := range b.Spec.VersionsBundles {
		artifact := GetPackageBundleRef(vb)
		packages, err := fetchPackages(ctx, vb, artifact)
		if err != nil {
			fmt.Printf("error finding packages: %v", err)
			continue
		}
		images = append(images, packages...)
	}
	return images
}

func fetchPackages(ctx context.Context, versionsBundle releasev1.VersionsBundle, art string) ([]releasev1.Image, error) {
	data, err := Pull(ctx, art)
	if err != nil {
		return nil, err
	}
	ctrl := versionsBundle.PackageController.Controller
	bundle := &packagesv1.PackageBundle{}
	err = yaml.Unmarshal(data, bundle)
	if err != nil {
		return nil, err
	}
	var images []releasev1.Image
	for _, p := range bundle.Spec.Packages {
		pI := releasev1.Image{
			Name:        p.Name,
			Description: p.Name,
			OS:          ctrl.OS,
			OSName:      ctrl.OSName,
			URI:         fmt.Sprintf("%s/%s:%s", p.Source.Registry, p.Source.Repository, p.Source.Versions[0].Name),
		}
		images = append(images, pI)
	}
	return images, nil
}

func Pull(ctx context.Context, art string) ([]byte, error) {
	puller := artifacts.NewRegistryPuller()

	data, err := puller.Pull(ctx, art)
	if err != nil {
		return nil, fmt.Errorf("unable to pull artifacts %v", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("latest package bundle artifact is empty")
	}

	return data, nil
}
