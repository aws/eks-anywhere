package curatedpackages

import (
	"bytes"
	"context"
	"fmt"

	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere/pkg/manifests"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

type PackagesReader struct {
	ManifestReader *manifests.Reader
}

func NewPackageReader(mr *manifests.Reader) *PackagesReader {
	return &PackagesReader{
		ManifestReader: mr,
	}
}

func (r *PackagesReader) ReadBundlesForVersion(version string) (*releasev1.Bundles, error) {
	return r.ManifestReader.ReadBundlesForVersion(version)
}

func (r *PackagesReader) ReadEKSD(eksaVersion, kubeVersion string) (*eksdv1.Release, error) {
	return r.ManifestReader.ReadEKSD(eksaVersion, kubeVersion)
}

func (r *PackagesReader) ReadImages(eksaVersion string) ([]releasev1.Image, error) {
	images, err := r.ManifestReader.ReadImages(eksaVersion)
	return images, err
}

func (r *PackagesReader) ReadImagesFromBundles(b *releasev1.Bundles) ([]releasev1.Image, error) {
	images, err := r.ManifestReader.ReadImagesFromBundles(b)
	for _, v := range b.Spec.VersionsBundles {
		images = append(images, v.PackagesControllerImage()...)
	}
	return images, err
}

func (r *PackagesReader) ReadCharts(eksaVersion string) ([]releasev1.Image, error) {
	images, err := r.ManifestReader.ReadCharts(eksaVersion)
	return images, err
}

func (r *PackagesReader) ReadChartsFromBundles(ctx context.Context, b *releasev1.Bundles) []releasev1.Image {
	images := r.ManifestReader.ReadChartsFromBundles(ctx, b)
	for _, vb := range b.Spec.VersionsBundles {
		art := GetPackageBundle(vb)
		packages, err := FetchPackages(vb, ctx, art)
		if err != nil {
			fmt.Sprintf("error finding packages: %v", err)
			continue
		}
		images = append(images, packages...)
	}
	return images
}

func FetchPackages(versionsBundle releasev1.VersionsBundle, ctx context.Context, art string) ([]releasev1.Image, error) {
	data, err := Pull(ctx, art)
	ctrl := versionsBundle.PackageController.Controller
	if err != nil {
		return nil, err
	}
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
			URI:         p.Source.Registry + "/" + p.Source.Repository + ":" + p.Source.Versions[0].Name,
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

func Push(ctx context.Context, art, ref, fileName string, fileContent []byte) error {
	registry, err := content.NewRegistry(content.RegistryOptions{})
	if err != nil {
		return fmt.Errorf("creating registry: %w", err)
	}
	memoryStore := content.NewMemory()
	desc, err := memoryStore.Add(fileName, "", fileContent)

	manifest, manifestDesc, config, configDesc, err := content.GenerateManifestAndConfig(nil, nil, desc)
	memoryStore.Set(configDesc, config)
	err = memoryStore.StoreManifest(ref, manifestDesc, manifest)

	fmt.Printf("Pushing %s to %s...\n", fileName, ref)
	desc, err = oras.Copy(ctx, memoryStore, ref, registry, "")
	fmt.Printf("Pushed to %s with digest %s\n", ref, desc.Digest)
	return nil
}
