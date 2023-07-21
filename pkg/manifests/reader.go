package manifests

import (
	"context"
	"fmt"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type FileReader interface {
	ReadFile(url string) ([]byte, error)
}

type Reader struct {
	FileReader
	releasesManifestURL string
}

type ReaderOpt func(*Reader)

func WithReleasesManifest(manifestURL string) ReaderOpt {
	return func(r *Reader) {
		r.releasesManifestURL = manifestURL
	}
}

func NewReader(filereader FileReader, opts ...ReaderOpt) *Reader {
	r := &Reader{
		FileReader:          filereader,
		releasesManifestURL: releases.ManifestURL(),
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

// ReadReleaseForVersion returns an EksaRelease based on a version.
func (r *Reader) ReadReleaseForVersion(version string) (*releasev1.EksARelease, error) {
	rls, err := releases.ReadReleasesFromURL(r, r.releasesManifestURL)
	if err != nil {
		return nil, err
	}

	release, err := releases.ReleaseForVersion(rls, version)
	if err != nil {
		return nil, err
	}

	if release == nil {
		return nil, fmt.Errorf("invalid version %s, no matching release found", version)
	}

	return release, nil
}

// ReadBundlesForVersion returns a Bundle based on the version.
func (r *Reader) ReadBundlesForVersion(version string) (*releasev1.Bundles, error) {
	release, err := r.ReadReleaseForVersion(version)
	if err != nil {
		return nil, err
	}

	return releases.ReadBundlesForRelease(r, release)
}

func (r *Reader) ReadEKSD(eksaVersion, kubeVersion string) (*eksdv1.Release, error) {
	b, err := r.ReadBundlesForVersion(eksaVersion)
	if err != nil {
		return nil, err
	}

	versionsBundle := bundles.VersionsBundleForKubernetesVersion(b, kubeVersion)
	if versionsBundle == nil {
		return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", kubeVersion, b.Spec.Number)
	}

	return bundles.ReadEKSD(r, *versionsBundle)
}

func (r *Reader) ReadImages(eksaVersion string) ([]releasev1.Image, error) {
	bundle, err := r.ReadBundlesForVersion(eksaVersion)
	if err != nil {
		return nil, err
	}

	return bundles.ReadImages(r, bundle)
}

func (r *Reader) ReadImagesFromBundles(_ context.Context, b *releasev1.Bundles) ([]releasev1.Image, error) {
	return bundles.ReadImages(r, b)
}

func (r *Reader) ReadCharts(eksaVersion string) ([]releasev1.Image, error) {
	bundle, err := r.ReadBundlesForVersion(eksaVersion)
	if err != nil {
		return nil, err
	}

	return bundles.Charts(bundle), nil
}

func (r *Reader) ReadChartsFromBundles(ctx context.Context, b *releasev1.Bundles) []releasev1.Image {
	return bundles.Charts(b)
}
