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
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/pkg/version"
)

const capasProjectPath = "projects/aws/cluster-api-provider-aws-snow"

func GetSnowBundle(r *releasetypes.ReleaseConfig, imageDigests map[string]string) (anywherev1alpha1.SnowBundle, error) {
	capasBundleArtifacts := map[string][]releasetypes.Artifact{
		"cluster-api-provider-aws-snow": r.BundleArtifactsTable["cluster-api-provider-aws-snow"],
		"kube-rbac-proxy":               r.BundleArtifactsTable["kube-rbac-proxy"],
		"kube-vip":                      r.BundleArtifactsTable["kube-vip"],
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(capasBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range capasBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image

				bundleImageArtifact := anywherev1alpha1.Image{
					Name:        imageArtifact.AssetName,
					Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
					OS:          imageArtifact.OS,
					Arch:        imageArtifact.Arch,
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
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
		version.NewVersionerWithGITTAG(r.BuildRepoSource, capasProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.SnowBundle{}, errors.Wrapf(err, "Error getting version for CAPAS")
	}

	bundle := anywherev1alpha1.SnowBundle{
		Version:    version,
		Manager:    bundleImageArtifacts["cluster-api-snow-controller"],
		KubeVip:    bundleImageArtifacts["kube-vip"],
		Components: bundleManifestArtifacts["infrastructure-components.yaml"],
		Metadata:   bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
