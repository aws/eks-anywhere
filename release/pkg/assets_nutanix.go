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
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func (r *ReleaseConfig) GetNutanixBundle(eksDReleaseChannel string, imageDigests map[string]string) (anywherev1alpha1.NutanixBundle, error) {
	nutanixBundleArtifacts := map[string][]Artifact{
		"cluster-api-provider-nutanix": r.BundleArtifactsTable["cluster-api-provider-nutanix"],
		"kube-rbac-proxy":              r.BundleArtifactsTable["kube-rbac-proxy"],
		"kube-vip":                     r.BundleArtifactsTable["kube-vip"],
	}
	sortedComponentNames := sortArtifactsMap(nutanixBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range nutanixBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api-provider-nutanix" {
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

				manifestContents, err := ioutil.ReadFile(filepath.Join(manifestArtifact.ArtifactPath, manifestArtifact.ReleaseName))
				if err != nil {
					return anywherev1alpha1.NutanixBundle{}, err
				}
				manifestHash := generateManifestHash(manifestContents)
				artifactHashes = append(artifactHashes, manifestHash)
			}
		}
	}

	nutanixCloudProviderArtifacts := r.BundleArtifactsTable[fmt.Sprintf("cloud-provider-nutanix-%s", eksDReleaseChannel)]

	for _, artifact := range nutanixCloudProviderArtifacts {
		imageArtifact := artifact.Image

		bundleArtifact := anywherev1alpha1.Image{
			Name:        imageArtifact.AssetName,
			Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
			OS:          imageArtifact.OS,
			Arch:        imageArtifact.Arch,
			URI:         imageArtifact.ReleaseImageURI,
			ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
		}
		bundleImageArtifacts[imageArtifact.AssetName] = bundleArtifact
		artifactHashes = append(artifactHashes, bundleArtifact.ImageDigest)
	}

	if r.DryRun {
		componentChecksum = fakeComponentChecksum
	} else {
		componentChecksum = generateComponentHash(artifactHashes)
	}
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(r.BuildRepoSource, capvProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.NutanixBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-nutanix")
	}

	bundle := anywherev1alpha1.NutanixBundle{
		Version:              version,
		ClusterAPIController: bundleImageArtifacts["cluster-api-nutanix-controller"],
		Components:           bundleManifestArtifacts["infrastructure-components.yaml"],
		ClusterTemplate:      bundleManifestArtifacts["cluster-template.yaml"],
		Metadata:             bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
