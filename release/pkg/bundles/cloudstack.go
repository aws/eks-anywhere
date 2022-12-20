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
	"github.com/aws/eks-anywhere/release/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	"github.com/aws/eks-anywhere/release/pkg/version"
)

func GetCloudStackBundle(r *releasetypes.ReleaseConfig, imageDigests map[string]string) (anywherev1alpha1.CloudStackBundle, error) {
	cloudstackBundleArtifacts := map[string][]releasetypes.Artifact{
		"cluster-api-provider-cloudstack": r.BundleArtifactsTable["cluster-api-provider-cloudstack"],
		"kube-vip":                        r.BundleArtifactsTable["kube-vip"],
		"kube-rbac-proxy":                 r.BundleArtifactsTable["kube-rbac-proxy"],
	}

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}
	for componentName, artifacts := range cloudstackBundleArtifacts {
		for _, artifact := range artifacts {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api-provider-cloudstack" {
					sourceBranch = imageArtifact.SourcedFromBranch
				}
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
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
				if err != nil {
					return anywherev1alpha1.CloudStackBundle{}, err
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
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CapcProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.CloudStackBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-cloudstack")
	}

	bundle := anywherev1alpha1.CloudStackBundle{
		Version:              version,
		ClusterAPIController: bundleImageArtifacts["cluster-api-provider-cloudstack"],
		KubeVip:              bundleImageArtifacts["kube-vip"],
		KubeRbacProxy:        bundleImageArtifacts["kube-rbac-proxy"],
		Components:           bundleManifestArtifacts["infrastructure-components.yaml"],
		Metadata:             bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
