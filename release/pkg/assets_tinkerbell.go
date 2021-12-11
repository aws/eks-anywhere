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

package pkg

import (
	"fmt"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func (r *ReleaseConfig) GetTinkerbellBundle(eksDReleaseChannel string, imageDigests map[string]string) (anywherev1alpha1.TinkerbellBundle, error) {
	tinkerbellBundleArtifacts := map[string][]Artifact{
		"cluster-api-provider-tinkerbell": r.BundleArtifactsTable["cluster-api-provider-tinkerbell"],
	}

	var version string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	for componentName, artifacts := range tinkerbellBundleArtifacts {
		for _, artifact := range artifacts {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api-provider-tinkerbell" {
					componentVersion, err := BuildComponentVersion(
						newVersionerWithGITTAG(r.BuildRepoSource, captProjectPath, imageArtifact.SourcedFromBranch, r),
					)
					if err != nil {
						return anywherev1alpha1.TinkerbellBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-tinkerbell")
					}
					version = componentVersion
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
			}

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
			}
		}
	}

	bundle := anywherev1alpha1.TinkerbellBundle{
		Version:              version,
		ClusterAPIController: bundleImageArtifacts["cluster-api-provider-tinkerbell"],
		Components:           bundleManifestArtifacts["infrastructure-components.yaml"],
		ClusterTemplate:      bundleManifestArtifacts["cluster-template.yaml"],
		Metadata:             bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
