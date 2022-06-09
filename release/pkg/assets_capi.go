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
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const capiProjectPath = "projects/kubernetes-sigs/cluster-api"

// GetCAPIAssets returns the eks-a artifacts for cluster-api
func (r *ReleaseConfig) GetCAPIAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(capiProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	capiImages := []string{
		"cluster-api-controller",
		"kubeadm-bootstrap-controller",
		"kubeadm-control-plane-controller",
	}

	componentTagOverrideMap := map[string]ImageTagOverride{}
	sourcedFromBranch := r.BuildRepoBranchName
	artifacts := []Artifact{}
	for _, image := range capiImages {
		repoName := fmt.Sprintf("kubernetes-sigs/cluster-api/%s", image)
		tagOptions := map[string]string{
			"gitTag":      gitTag,
			"projectPath": capiProjectPath,
		}

		sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}
		if sourcedFromBranch != r.BuildRepoBranchName {
			gitTag, err = r.readGitTag(capiProjectPath, sourcedFromBranch)
			if err != nil {
				return nil, errors.Cause(err)
			}
			tagOptions["gitTag"] = gitTag
		}
		releaseImageUri, err := r.GetReleaseImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}

		imageArtifact := &ImageArtifact{
			AssetName:         image,
			SourceImageURI:    sourceImageUri,
			ReleaseImageURI:   releaseImageUri,
			Arch:              []string{"amd64"},
			OS:                "linux",
			GitTag:            gitTag,
			ProjectPath:       capiProjectPath,
			SourcedFromBranch: sourcedFromBranch,
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})

		componentTagOverrideMap[image] = ImageTagOverride{
			Repository: repoName,
			ReleaseUri: imageArtifact.ReleaseImageURI,
		}
	}

	var imageTagOverrides []ImageTagOverride

	kubeRbacProxyImageTagOverride, err := r.GetKubeRbacProxyImageTagOverride()
	if err != nil {
		return nil, errors.Cause(err)
	}

	componentManifestMap := map[string][]string{
		"bootstrap-kubeadm":     {"bootstrap-components.yaml", "metadata.yaml"},
		"cluster-api":           {"core-components.yaml", "metadata.yaml"},
		"control-plane-kubeadm": {"control-plane-components.yaml", "metadata.yaml"},
	}
	sortedComponentNames := sortManifestMap(componentManifestMap)

	for _, component := range sortedComponentNames {
		manifestList := componentManifestMap[component]
		for _, manifest := range manifestList {
			var sourceS3Prefix string
			var releaseS3Path string
			var imageTagOverride ImageTagOverride
			latestPath := getLatestUploadDestination(sourcedFromBranch)

			if r.DevRelease || r.ReleaseEnvironment == "development" {
				sourceS3Prefix = fmt.Sprintf("%s/%s/manifests/%s/%s", capiProjectPath, latestPath, component, gitTag)
			} else {
				sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/cluster-api/manifests/%s/%s", r.BundleNumber, component, gitTag)
			}

			if r.DevRelease {
				releaseS3Path = fmt.Sprintf("artifacts/%s/cluster-api/manifests/%s/%s", r.DevReleaseUriVersion, component, gitTag)
			} else {
				releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/cluster-api/manifests/%s/%s", r.BundleNumber, component, gitTag)
			}

			cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, manifest))
			if err != nil {
				return nil, errors.Cause(err)
			}

			if component == "bootstrap-kubeadm" {
				imageTagOverride = componentTagOverrideMap["kubeadm-bootstrap-controller"]
			} else if component == "cluster-api" {
				imageTagOverride = componentTagOverrideMap["cluster-api-controller"]
			} else {
				imageTagOverride = componentTagOverrideMap["kubeadm-control-plane-controller"]
			}

			imageTagOverrides = append(imageTagOverrides, imageTagOverride, kubeRbacProxyImageTagOverride)

			manifestArtifact := &ManifestArtifact{
				SourceS3Key:       manifest,
				SourceS3Prefix:    sourceS3Prefix,
				ArtifactPath:      filepath.Join(r.ArtifactDir, fmt.Sprintf("%s-manifests", component), r.BuildRepoHead),
				ReleaseName:       manifest,
				ReleaseS3Path:     releaseS3Path,
				ReleaseCdnURI:     cdnURI,
				ImageTagOverrides: imageTagOverrides,
				GitTag:            gitTag,
				ProjectPath:       capiProjectPath,
				SourcedFromBranch: sourcedFromBranch,
				Component:         component,
			}
			artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
		}
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetCoreClusterAPIBundle(imageDigests map[string]string) (anywherev1alpha1.CoreClusterAPI, error) {
	coreClusterAPIBundleArtifacts := map[string][]Artifact{
		"cluster-api":     r.BundleArtifactsTable["cluster-api"],
		"kube-rbac-proxy": r.BundleArtifactsTable["kube-rbac-proxy"],
	}
	sortedComponentNames := sortArtifactsMap(coreClusterAPIBundleArtifacts)

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
				if manifestArtifact.Component != "cluster-api" {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}
				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := r.GenerateManifestHash(manifestArtifact)
				if err != nil {
					return anywherev1alpha1.CoreClusterAPI{}, err
				}
				artifactHashes = append(artifactHashes, manifestHash)
			}
		}
	}

	if r.DryRun {
		componentChecksum = fakeComponentChecksum
	} else {
		componentChecksum = generateComponentHash(artifactHashes)
	}
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(r.BuildRepoSource, capiProjectPath, sourceBranch, r),
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

func (r *ReleaseConfig) GetKubeadmBootstrapBundle(imageDigests map[string]string) (anywherev1alpha1.KubeadmBootstrapBundle, error) {
	kubeadmBootstrapBundleArtifacts := map[string][]Artifact{
		"cluster-api":     r.BundleArtifactsTable["cluster-api"],
		"kube-rbac-proxy": r.BundleArtifactsTable["kube-rbac-proxy"],
	}
	sortedComponentNames := sortArtifactsMap(kubeadmBootstrapBundleArtifacts)

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
				if manifestArtifact.Component != "bootstrap-kubeadm" {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}
				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := r.GenerateManifestHash(manifestArtifact)
				if err != nil {
					return anywherev1alpha1.KubeadmBootstrapBundle{}, err
				}
				artifactHashes = append(artifactHashes, manifestHash)
			}
		}
	}

	if r.DryRun {
		componentChecksum = fakeComponentChecksum
	} else {
		componentChecksum = generateComponentHash(artifactHashes)
	}
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(r.BuildRepoSource, capiProjectPath, sourceBranch, r),
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

func (r *ReleaseConfig) GetKubeadmControlPlaneBundle(imageDigests map[string]string) (anywherev1alpha1.KubeadmControlPlaneBundle, error) {
	kubeadmControlPlaneBundleArtifacts := map[string][]Artifact{
		"cluster-api":     r.BundleArtifactsTable["cluster-api"],
		"kube-rbac-proxy": r.BundleArtifactsTable["kube-rbac-proxy"],
	}
	sortedComponentNames := sortArtifactsMap(kubeadmControlPlaneBundleArtifacts)

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
				if manifestArtifact.Component != "control-plane-kubeadm" {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}
				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := r.GenerateManifestHash(manifestArtifact)
				if err != nil {
					return anywherev1alpha1.KubeadmControlPlaneBundle{}, err
				}
				artifactHashes = append(artifactHashes, manifestHash)
			}
		}
	}

	if r.DryRun {
		componentChecksum = fakeComponentChecksum
	} else {
		componentChecksum = generateComponentHash(artifactHashes)
	}
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(r.BuildRepoSource, capiProjectPath, sourceBranch, r),
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

func sortManifestMap(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
