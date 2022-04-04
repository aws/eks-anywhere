package manifests

import (
	"fmt"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/bundles"
	"github.com/aws/eks-anywhere/pkg/releases"
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

func (r *Reader) ReadBundlesForVersion(version string) (*releasev1.Bundles, error) {
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

func (r *Reader) ReadCharts(eksaVersion string) ([]releasev1.Image, error) {
	bundle, err := r.ReadBundlesForVersion(eksaVersion)
	if err != nil {
		return nil, err
	}

	return bundles.Charts(bundle), nil
}
