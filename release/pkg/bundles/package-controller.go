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
	"github.com/aws/eks-anywhere/release/pkg/constants"
	"github.com/aws/eks-anywhere/release/pkg/helm"
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
	var helmdir string
	var URI string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	artifactHashes := []string{}

	// Find the source of the Helm chart prior to the initial loop.
	for _, componentName := range sortedComponentNames {
		for _, artifact := range artifacts[componentName] {
			if strings.HasSuffix(artifact.Image.AssetName, "helm") {
				URI = artifact.Image.SourceImageURI
			}
		}
	}
	driver, err := helm.NewHelm()
	if err != nil {
		return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "Error CreateReleaseHelmDriver")
	}

	if !r.DevRelease {
		helmdir, err = helm.GetHelmDest(driver, URI, "eks-anywhere-packages")
		if err != nil {
			return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "Error GetHelmDest")
		}
	}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range artifacts[componentName] {
			imageArtifact := artifact.Image
			sourceBranch = imageArtifact.SourcedFromBranch
			bundleImageArtifact := anywherev1alpha1.Image{}
			if strings.HasSuffix(imageArtifact.AssetName, "helm") {
				assetName := strings.TrimSuffix(imageArtifact.AssetName, "-helm")
				if !r.DevRelease {
					err := helm.ModifyChartYaml(*imageArtifact, r, driver, helmdir)
					if err != nil {
						return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "Error modifying and pushing helm Chart.yaml")
					}
				}
				bundleImageArtifact = anywherev1alpha1.Image{
					Name:        assetName,
					Description: fmt.Sprintf("Helm chart for %s", assetName),
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
				}
			} else {
				// Set the default digest, in case we're doing a dev release it won't fail.
				var digest string
				digest = imageDigests[imageArtifact.ReleaseImageURI]

				if !r.DevRelease {
					requires, err := helm.GetChartImageTags(driver, helmdir)
					if err != nil {
						return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "Error retrieving requires.yaml")
					}
					for _, images := range requires.Spec.Images {
						if images.Repository == imageArtifact.AssetName {
							digest = images.Digest
						}
					}
				}

				bundleImageArtifact = anywherev1alpha1.Image{
					Name:        imageArtifact.AssetName,
					Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
					OS:          imageArtifact.OS,
					Arch:        imageArtifact.Arch,
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: digest,
				}
			}
			bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
			artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
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
