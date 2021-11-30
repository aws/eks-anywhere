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
	"strings"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// GetCAPIAssets returns the eks-a artifacts for cluster-api
func (r *ReleaseConfig) GetCAPIAssets() ([]Artifact, error) {
	gitTag, err := r.getCAPIGitTag()
	if err != nil {
		return nil, errors.Cause(err)
	}

	capiImages := []string{
		"cluster-api-controller",
		"kubeadm-bootstrap-controller",
		"kubeadm-control-plane-controller",
	}

	componentTagOverrideMap := map[string]ImageTagOverride{}
	artifacts := []Artifact{}
	for _, image := range capiImages {
		repoName := fmt.Sprintf("kubernetes-sigs/cluster-api/%s", image)
		tagOptions := map[string]string{
			"gitTag": gitTag,
		}
		releaseImageUri, err := r.GetReleaseImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}

		imageArtifact := &ImageArtifact{
			AssetName:       image,
			SourceImageURI:  r.GetSourceImageURI(image, repoName, tagOptions),
			ReleaseImageURI: releaseImageUri,
			Arch:            []string{"amd64"},
			OS:              "linux",
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

	for component, manifestList := range componentManifestMap {
		for _, manifest := range manifestList {
			var sourceS3Prefix string
			var releaseS3Path string
			var imageTagOverride ImageTagOverride
			latestPath := r.getLatestUploadDestination()

			if r.DevRelease || r.ReleaseEnvironment == "development" {
				sourceS3Prefix = fmt.Sprintf("projects/kubernetes-sigs/cluster-api/%s/manifests/%s/%s", latestPath, component, gitTag)
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
			}
			artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
		}
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetCoreClusterAPIBundle(imageDigests map[string]string) (anywherev1alpha1.CoreClusterAPI, error) {
	coreClusterAPIBundleArtifactsFuncs := map[string]func() ([]Artifact, error){
		"cluster-api": r.GetCAPIAssets,
		"kube-proxy":  r.GetKubeRbacProxyAssets,
	}

	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, "projects/kubernetes-sigs/cluster-api")),
	)
	if err != nil {
		return anywherev1alpha1.CoreClusterAPI{}, errors.Wrapf(err, "Error getting version for cluster-api")
	}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	for componentName, artifactFunc := range coreClusterAPIBundleArtifactsFuncs {
		artifacts, err := artifactFunc()
		if err != nil {
			return anywherev1alpha1.CoreClusterAPI{}, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
		}

		for _, artifact := range artifacts {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api" {
					if imageArtifact.AssetName != "cluster-api-controller" {
						continue
					}
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
				if !strings.Contains(manifestArtifact.ReleaseName, "core") && !strings.Contains(manifestArtifact.ReleaseName, "metadata") {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
			}
		}
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
	kubeadmBootstrapBundleArtifactsFuncs := map[string]func() ([]Artifact, error){
		"cluster-api": r.GetCAPIAssets,
		"kube-proxy":  r.GetKubeRbacProxyAssets,
	}

	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, "projects/kubernetes-sigs/cluster-api")),
	)
	if err != nil {
		return anywherev1alpha1.KubeadmBootstrapBundle{}, errors.Wrapf(err, "Error getting version for cluster-api")
	}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	for componentName, artifactFunc := range kubeadmBootstrapBundleArtifactsFuncs {
		artifacts, err := artifactFunc()
		if err != nil {
			return anywherev1alpha1.KubeadmBootstrapBundle{}, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
		}

		for _, artifact := range artifacts {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api" {
					if imageArtifact.AssetName != "kubeadm-bootstrap-controller" {
						continue
					}
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
				if !strings.Contains(manifestArtifact.ReleaseName, "bootstrap") && !strings.Contains(manifestArtifact.ReleaseName, "metadata") {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
			}
		}
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
	kubeadmControlPlaneBundleArtifactsFuncs := map[string]func() ([]Artifact, error){
		"cluster-api": r.GetCAPIAssets,
		"kube-proxy":  r.GetKubeRbacProxyAssets,
	}

	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, "projects/kubernetes-sigs/cluster-api")),
	)
	if err != nil {
		return anywherev1alpha1.KubeadmControlPlaneBundle{}, errors.Wrapf(err, "Error getting version for cluster-api")
	}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	for componentName, artifactFunc := range kubeadmControlPlaneBundleArtifactsFuncs {
		artifacts, err := artifactFunc()
		if err != nil {
			return anywherev1alpha1.KubeadmControlPlaneBundle{}, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
		}

		for _, artifact := range artifacts {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api" {
					if imageArtifact.AssetName != "kubeadm-control-plane-controller" {
						continue
					}
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
				if !strings.Contains(manifestArtifact.ReleaseName, "control-plane") && !strings.Contains(manifestArtifact.ReleaseName, "metadata") {
					continue
				}

				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
			}
		}
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

func (r *ReleaseConfig) getCAPIGitTag() (string, error) {
	projectSource := "projects/kubernetes-sigs/cluster-api"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return "", errors.Cause(err)
	}

	return gitTag, nil
}
