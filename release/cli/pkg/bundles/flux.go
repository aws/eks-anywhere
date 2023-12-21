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

func GetFluxBundle(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.FluxBundle, error) {
	projectsInBundle := []string{"helm-controller", "kustomize-controller", "notification-controller", "source-controller"}
	fluxBundleArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, err := r.BundleArtifactsTable.Load(project)
		if err != nil {
			return anywherev1alpha1.FluxBundle{}, fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		fluxBundleArtifacts[project] = projectArtifacts
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(fluxBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range fluxBundleArtifacts[componentName] {
			imageArtifact := artifact.Image
			sourceBranch = imageArtifact.SourcedFromBranch
			imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
			if err != nil {
				return anywherev1alpha1.FluxBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
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
	}

	if r.DryRun {
		componentChecksum = version.FakeComponentChecksum
	} else {
		componentChecksum = version.GenerateComponentHash(artifactHashes, r.DryRun)
	}
	version, err := version.BuildComponentVersion(
		version.NewMultiProjectVersionerWithGITTAG(r.BuildRepoSource,
			constants.FluxcdRootPath,
			constants.Flux2ProjectPath,
			sourceBranch,
			r,
		),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.FluxBundle{}, errors.Wrap(err, "failed generating version for flux bundle")
	}

	bundle := anywherev1alpha1.FluxBundle{
		Version:                version,
		SourceController:       bundleImageArtifacts["source-controller"],
		KustomizeController:    bundleImageArtifacts["kustomize-controller"],
		HelmController:         bundleImageArtifacts["helm-controller"],
		NotificationController: bundleImageArtifacts["notification-controller"],
	}

	return bundle, nil
}
