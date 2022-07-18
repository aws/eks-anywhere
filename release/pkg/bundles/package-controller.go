// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bundles

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	"github.com/aws/eks-anywhere/release/pkg/version"
)

const (
	packagesProjectPath = "projects/aws/eks-anywhere-packages"
)

func GetPackagesBundle(r *releasetypes.ReleaseConfig, imageDigests map[string]string) (anywherev1alpha1.PackageBundle, error) {
	artifacts := r.BundleArtifactsTable["eks-anywhere-packages"]

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	artifactHashes := []string{}

	for _, artifact := range artifacts {
		imageArtifact := artifact.Image
		sourceBranch = imageArtifact.SourcedFromBranch
		bundleImageArtifact := anywherev1alpha1.Image{}
		if strings.HasSuffix(imageArtifact.AssetName, "helm") {
			assetName := strings.TrimSuffix(imageArtifact.AssetName, "-helm")
			bundleImageArtifact = anywherev1alpha1.Image{
				Name:        assetName,
				Description: fmt.Sprintf("Helm chart for %s", assetName),
				URI:         imageArtifact.ReleaseImageURI,
				ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
			}
		} else {
			bundleImageArtifact = anywherev1alpha1.Image{
				Name:        imageArtifact.AssetName,
				Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
				OS:          imageArtifact.OS,
				Arch:        imageArtifact.Arch,
				URI:         imageArtifact.ReleaseImageURI,
				ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
			}
		}

		bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
		artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
	}

	if r.DryRun {
		componentChecksum = version.FakeComponentChecksum
	} else {
		componentChecksum = version.GenerateComponentHash(artifactHashes, r.DryRun)
	}
	version, err := version.BuildComponentVersion(
		version.NewVersionerWithGITTAG(r.BuildRepoSource, packagesProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "Error getting version for EKS Anywhere package controller")
	}

	bundle := anywherev1alpha1.PackageBundle{
		Version:    version,
		Controller: bundleImageArtifacts["eks-anywhere-packages"],
		HelmChart:  bundleImageArtifacts["eks-anywhere-packages-helm"],
	}
	return bundle, nil
}
