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
	"github.com/aws/eks-anywhere/release/pkg/aws/ecr"
	"github.com/aws/eks-anywhere/release/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/pkg/version"
)

func GetPackagesBundle(r *releasetypes.ReleaseConfig, imageDigests map[string]string) (anywherev1alpha1.PackageBundle, error) {
	artifacts := map[string][]releasetypes.Artifact{
		"eks-anywhere-packages": r.BundleArtifactsTable["eks-anywhere-packages"],
		"ecr-token-refresher":   r.BundleArtifactsTable["ecr-token-refresher"],
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(artifacts)

	var sourceBranch string
	var componentChecksum string
	var Helmtag, Imagetag string
	var Helmsha, Imagesha string
	var err error
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	artifactHashes := []string{}

	// Find latest Package Dev build for the Helm chart and Image which will always start with `0.0.0` and is built off of the package Github repo main on every commit.
	// If we can't find the build starting with our substring, we default to the original dev tag.
	if r.DevRelease {
		Helmtag, Helmsha, err = ecr.FilterECRRepoByTagPrefix(r.SourceClients.ECR.EcrClient, "eks-anywhere-packages", "0.0.0")
		if err != nil {
			fmt.Printf("Error getting dev version helm tag EKS Anywhere package controller, using latest version %v", err)
		}
		Imagetag, Imagesha, err = ecr.FilterECRRepoByTagPrefix(r.SourceClients.ECR.EcrClient, "eks-anywhere-packages", "v0.0.0")
		if err != nil {
			fmt.Printf("Error getting dev version Image tag EKS Anywhere package controller, using latest version %v", err)
		}
	}
	for _, componentName := range sortedComponentNames {
		for _, artifact := range artifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				sourceBranch = imageArtifact.SourcedFromBranch
				bundleImageArtifact := anywherev1alpha1.Image{}
				if strings.HasSuffix(imageArtifact.AssetName, "helm") {
					Digest := imageDigests[imageArtifact.ReleaseImageURI]
					if r.DevRelease && Helmsha != "" && Helmtag != "" {
						Digest = Helmsha
						imageArtifact.ReleaseImageURI = replaceTag(imageArtifact.ReleaseImageURI, Helmtag)
					}
					assetName := strings.TrimSuffix(imageArtifact.AssetName, "-helm")
					bundleImageArtifact = anywherev1alpha1.Image{
						Name:        assetName,
						Description: fmt.Sprintf("Helm chart for %s", assetName),
						URI:         imageArtifact.ReleaseImageURI,
						ImageDigest: Digest,
					}
				} else {
					Digest := imageDigests[imageArtifact.ReleaseImageURI]
					if r.DevRelease && Imagesha != "" && Imagetag != "" {
						Digest = Imagesha
						imageArtifact.ReleaseImageURI = replaceTag(imageArtifact.ReleaseImageURI, Imagetag)
					}
					bundleImageArtifact = anywherev1alpha1.Image{
						Name:        imageArtifact.AssetName,
						Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
						OS:          imageArtifact.OS,
						Arch:        imageArtifact.Arch,
						URI:         imageArtifact.ReleaseImageURI,
						ImageDigest: Digest,
					}
				}
				bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
				artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
			}
		}
	}

	if r.DryRun {
		componentChecksum = version.FakeComponentChecksum
	} else {
		componentChecksum = version.GenerateComponentHash(artifactHashes, r.DryRun)
	}
	version, err := version.BuildComponentVersion(
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.PackagesProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "Error getting version for EKS Anywhere package controller")
	}

	bundle := anywherev1alpha1.PackageBundle{
		Version:        version,
		Controller:     bundleImageArtifacts["eks-anywhere-packages"],
		TokenRefresher: bundleImageArtifacts["ecr-token-refresher"],
		HelmChart:      bundleImageArtifacts["eks-anywhere-packages-helm"],
	}
	return bundle, nil
}

// replaceTag is used to replace the tag of an Image URI with a string.
func replaceTag(uri, tag string) string {
	NewURIList := strings.Split(uri, ":")
	if len(NewURIList) < 2 {
		return uri
	}
	NewURIList[len(NewURIList)-1] = tag
	uri = strings.Join(NewURIList[:], ":")
	return uri
}
