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

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/cli/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/cli/pkg/version"
)

func GetSnowBundle(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.SnowBundle, error) {
	projectsInBundle := []string{"cluster-api-provider-aws-snow", "kube-rbac-proxy", "kube-vip"}
	snowBundleArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, err := r.BundleArtifactsTable.Load(project)
		if err != nil {
			return anywherev1alpha1.SnowBundle{}, fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		snowBundleArtifacts[project] = projectArtifacts
	}
	bottlerocketBootstrapSnowArtifacts, err := r.BundleArtifactsTable.Load(fmt.Sprintf("bottlerocket-bootstrap-%s", eksDReleaseChannel))
	if err != nil {
		return anywherev1alpha1.SnowBundle{}, fmt.Errorf("artifacts for project bottlerocket-bootstrap-%s not found in bundle artifacts table", eksDReleaseChannel)
	}
	snowBundleArtifacts["bottlerocket-bootstrap-snow"] = bottlerocketBootstrapSnowArtifacts

	sortedComponentNames := bundleutils.SortArtifactsMap(snowBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range snowBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
				if err != nil {
					return anywherev1alpha1.SnowBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
				}
				bundleImageArtifact := anywherev1alpha1.Image{
					Name:        imageArtifact.AssetName,
					Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
					OS:          imageArtifact.OS,
					Arch:        imageArtifact.Arch,
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: imageDigest,
				}
				bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
				artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
			}

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				if componentName == "cluster-api-provider-aws-snow" {
					sourceBranch = manifestArtifact.SourcedFromBranch
				}
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
				if err != nil {
					return anywherev1alpha1.SnowBundle{}, err
				}
				artifactHashes = append(artifactHashes, manifestHash)
			}
		}
	}

	if r.DryRun {
		componentChecksum = version.FakeComponentChecksum
	} else {
		componentChecksum = version.GenerateComponentHash(artifactHashes, r.DryRun)
	}
	version, err := version.BuildComponentVersion(
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CapasProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.SnowBundle{}, errors.Wrapf(err, "Error getting version for CAPAS")
	}

	bundle := anywherev1alpha1.SnowBundle{
		Version:                   version,
		Manager:                   bundleImageArtifacts["cluster-api-snow-controller"],
		KubeVip:                   bundleImageArtifacts["kube-vip"],
		BottlerocketBootstrapSnow: bundleImageArtifacts["bottlerocket-bootstrap-snow"],
		Components:                bundleManifestArtifacts["infrastructure-components.yaml"],
		Metadata:                  bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
