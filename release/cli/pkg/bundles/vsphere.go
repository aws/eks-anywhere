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
	"sync"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/cli/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/cli/pkg/version"
)

func GetVsphereBundle(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests sync.Map) (anywherev1alpha1.VSphereBundle, error) {
	projectsInBundle := []string{"cluster-api-provider-vsphere", "kube-rbac-proxy", "kube-vip"}
	vsphereBundleArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, ok := r.BundleArtifactsTable.Load(project)
		if !ok {
			return anywherev1alpha1.VSphereBundle{}, fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		vsphereBundleArtifacts[project] = projectArtifacts.([]releasetypes.Artifact)
	}
	vSphereCloudProviderArtifacts, ok := r.BundleArtifactsTable.Load(fmt.Sprintf("cloud-provider-vsphere-%s", eksDReleaseChannel))
	if !ok {
		return anywherev1alpha1.VSphereBundle{}, fmt.Errorf("artifacts for project cloud-provider-vsphere-%s not found in bundle artifacts table", eksDReleaseChannel)
	}
	vsphereBundleArtifacts["cloud-provider-vsphere"] = vSphereCloudProviderArtifacts.([]releasetypes.Artifact)
	sortedComponentNames := bundleutils.SortArtifactsMap(vsphereBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range vsphereBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api-provider-vsphere" {
					sourceBranch = imageArtifact.SourcedFromBranch
				}
				imageDigest, ok := imageDigests.Load(imageArtifact.ReleaseImageURI)
				if !ok {
					return anywherev1alpha1.VSphereBundle{}, fmt.Errorf("digest for image %s not found in image digests table", imageArtifact.ReleaseImageURI)
				}
				bundleImageArtifact := anywherev1alpha1.Image{
					Name:        imageArtifact.AssetName,
					Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
					OS:          imageArtifact.OS,
					Arch:        imageArtifact.Arch,
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: imageDigest.(string),
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
					return anywherev1alpha1.VSphereBundle{}, err
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
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CapvProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.VSphereBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-vsphere")
	}

	bundle := anywherev1alpha1.VSphereBundle{
		Version:              version,
		ClusterAPIController: bundleImageArtifacts["cluster-api-provider-vsphere"],
		KubeProxy:            bundleImageArtifacts["kube-rbac-proxy"],
		Manager:              bundleImageArtifacts["cloud-provider-vsphere"],
		KubeVip:              bundleImageArtifacts["kube-vip"],
		Components:           bundleManifestArtifacts["infrastructure-components.yaml"],
		ClusterTemplate:      bundleManifestArtifacts["cluster-template.yaml"],
		Metadata:             bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
