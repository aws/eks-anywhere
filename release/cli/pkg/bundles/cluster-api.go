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

func GetCoreClusterAPIBundle(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.CoreClusterAPI, error) {
	projectsInBundle := []string{"cluster-api", "kube-rbac-proxy"}
	coreClusterAPIBundleArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, err := r.BundleArtifactsTable.Load(project)
		if err != nil {
			return anywherev1alpha1.CoreClusterAPI{}, fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		coreClusterAPIBundleArtifacts[project] = projectArtifacts
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(coreClusterAPIBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range coreClusterAPIBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api" {
					if imageArtifact.AssetName != "cluster-api-controller" {
						continue
					}
					sourceBranch = imageArtifact.SourcedFromBranch
				}

				imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
				if err != nil {
					return anywherev1alpha1.CoreClusterAPI{}, fmt.Errorf("loading digest from image digests table: %v", err)
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
				if manifestArtifact.Component != "cluster-api" {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}
				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
				if err != nil {
					return anywherev1alpha1.CoreClusterAPI{}, err
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
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CapiProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.CoreClusterAPI{}, errors.Wrapf(err, "Error getting version for cluster-api")
	}

	bundle := anywherev1alpha1.CoreClusterAPI{
		Version:    version,
		Controller: bundleImageArtifacts["cluster-api-controller"],
		KubeProxy:  bundleImageArtifacts["kube-rbac-proxy"],
		Components: bundleManifestArtifacts["core-components.yaml"],
		Metadata:   bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}

func GetKubeadmBootstrapBundle(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.KubeadmBootstrapBundle, error) {
	projectsInBundle := []string{"cluster-api", "kube-rbac-proxy"}
	kubeadmBootstrapBundleArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, err := r.BundleArtifactsTable.Load(project)
		if err != nil {
			return anywherev1alpha1.KubeadmBootstrapBundle{}, fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		kubeadmBootstrapBundleArtifacts[project] = projectArtifacts
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(kubeadmBootstrapBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range kubeadmBootstrapBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api" {
					if imageArtifact.AssetName != "kubeadm-bootstrap-controller" {
						continue
					}
					sourceBranch = imageArtifact.SourcedFromBranch
				}
				imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
				if err != nil {
					return anywherev1alpha1.KubeadmBootstrapBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
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
				if manifestArtifact.Component != "bootstrap-kubeadm" {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}
				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
				if err != nil {
					return anywherev1alpha1.KubeadmBootstrapBundle{}, err
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
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CapiProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.KubeadmBootstrapBundle{}, errors.Wrapf(err, "Error getting version for cluster-api")
	}

	bundle := anywherev1alpha1.KubeadmBootstrapBundle{
		Version:    version,
		Controller: bundleImageArtifacts["kubeadm-bootstrap-controller"],
		KubeProxy:  bundleImageArtifacts["kube-rbac-proxy"],
		Components: bundleManifestArtifacts["bootstrap-components.yaml"],
		Metadata:   bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}

func GetKubeadmControlPlaneBundle(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.KubeadmControlPlaneBundle, error) {
	projectsInBundle := []string{"cluster-api", "kube-rbac-proxy"}
	kubeadmControlPlaneBundleArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, err := r.BundleArtifactsTable.Load(project)
		if err != nil {
			return anywherev1alpha1.KubeadmControlPlaneBundle{}, fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		kubeadmControlPlaneBundleArtifacts[project] = projectArtifacts
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(kubeadmControlPlaneBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range kubeadmControlPlaneBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api" {
					if imageArtifact.AssetName != "kubeadm-control-plane-controller" {
						continue
					}
					sourceBranch = imageArtifact.SourcedFromBranch
				}
				imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
				if err != nil {
					return anywherev1alpha1.KubeadmControlPlaneBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
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
				if manifestArtifact.Component != "control-plane-kubeadm" {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}
				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
				if err != nil {
					return anywherev1alpha1.KubeadmControlPlaneBundle{}, err
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
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CapiProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.KubeadmControlPlaneBundle{}, errors.Wrapf(err, "Error getting version for cluster-api")
	}
	bundle := anywherev1alpha1.KubeadmControlPlaneBundle{
		Version:    version,
		Controller: bundleImageArtifacts["kubeadm-control-plane-controller"],
		KubeProxy:  bundleImageArtifacts["kube-rbac-proxy"],
		Components: bundleManifestArtifacts["control-plane-components.yaml"],
		Metadata:   bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
