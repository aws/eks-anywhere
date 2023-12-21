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

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func GetUpgraderBundle(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.UpgraderBundle, error) {
	upgraderArtifacts, err := r.BundleArtifactsTable.Load(fmt.Sprintf("upgrader-%s", eksDReleaseChannel))
	if err != nil {
		return anywherev1alpha1.UpgraderBundle{}, fmt.Errorf("artifacts for project upgrader-%s not found in bundle artifacts table", eksDReleaseChannel)
	}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}

	for _, artifact := range upgraderArtifacts {
		if artifact.Image != nil {
			imageArtifact := artifact.Image
			imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
			if err != nil {
				return anywherev1alpha1.UpgraderBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
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
		}
	}

	bundle := anywherev1alpha1.UpgraderBundle{
		Upgrader: bundleImageArtifacts["upgrader"],
	}

	return bundle, nil
}
