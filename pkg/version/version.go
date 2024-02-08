package version

import (
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
)

var gitVersion string

type Info struct {
	GitVersion         string `json:"version"`
	BundleManifestURL  string `json:"bundleManifestURL,omitempty"`
	ReleaseManifestURL string `json:"releaseManifestURL,omitempty"`
}

func Get() Info {
	return Info{
		GitVersion: gitVersion,
	}
}

// GetFullVersionInfo returns the complete version information for the
// EKS Anywhere, including Git version and bundle manifest URL
// associated with this release.
func GetFullVersionInfo() (Info, error) {
	reader := files.NewReader(files.WithEKSAUserAgent("cli", gitVersion))
	bundleManifestURL, err := releases.GetBundleManifestURL(reader, gitVersion)
	if err != nil {
		return Info{GitVersion: gitVersion}, err
	}

	return Info{
		GitVersion:         gitVersion,
		BundleManifestURL:  bundleManifestURL,
		ReleaseManifestURL: releases.ManifestURL(),
	}, nil
}
