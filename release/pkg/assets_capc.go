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

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"

	"github.com/pkg/errors"
)

const capcProjectPath = "projects/aws/cluster-api-provider-cloudstack"

// GetCapcAssets returns the eks-a artifacts for CAPC
func (r *ReleaseConfig) GetCapcAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(capcProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "cluster-api-cloudstack-controller"
	repoName := "aws/cluster-api-provider-cloudstack/release/manager"
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": capcProjectPath,
	}

	sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}
	if sourcedFromBranch != r.BuildRepoBranchName {
		gitTag, err = r.readGitTag(capcProjectPath, sourcedFromBranch)
		if err != nil {
			return nil, errors.Cause(err)
		}
		tagOptions["gitTag"] = gitTag
	}
	releaseImageUri, err := r.GetReleaseImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}

	imageArtifact := &ImageArtifact{
		AssetName:         name,
		SourceImageURI:    sourceImageUri,
		ReleaseImageURI:   releaseImageUri,
		Arch:              []string{"amd64"},
		OS:                "linux",
		GitTag:            gitTag,
		ProjectPath:       capcProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{{Image: imageArtifact}}

	var imageTagOverrides []ImageTagOverride

	kubeRbacProxyImageTagOverride, err := r.GetKubeRbacProxyImageTagOverride()
	if err != nil {
		return nil, errors.Cause(err)
	}

	imageTagOverride := ImageTagOverride{
		Repository: repoName,
		ReleaseUri: imageArtifact.ReleaseImageURI,
	}
	imageTagOverrides = append(imageTagOverrides, imageTagOverride, kubeRbacProxyImageTagOverride)

	manifestList := []string{
		"infrastructure-components.yaml",
		"cluster-template.yaml",
		"metadata.yaml",
	}

	for _, manifest := range manifestList {
		var sourceS3Prefix string
		var releaseS3Path string
		latestPath := getLatestUploadDestination(sourcedFromBranch)

		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Prefix = fmt.Sprintf("projects/aws/cluster-api-provider-cloudstack/%s/manifests/infrastructure-cloudstack/%s", latestPath, gitTag)
		} else {
			sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/cluster-api-provider-cloudstack/manifests/infrastructure-cloudstack/%s", r.BundleNumber, gitTag)
		}

		if r.DevRelease {
			releaseS3Path = fmt.Sprintf("artifacts/%s/cluster-api-provider-cloudstack/manifests/infrastructure-cloudstack/%s", r.DevReleaseUriVersion, gitTag)
		} else {
			releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/cluster-api-provider-cloudstack/manifests/infrastructure-cloudstack/%s", r.BundleNumber, gitTag)
		}

		cdnURI, err := r.GetURI(filepath.Join(
			releaseS3Path,
			manifest))
		if err != nil {
			return nil, errors.Cause(err)
		}

		manifestArtifact := &ManifestArtifact{
			SourceS3Key:       manifest,
			SourceS3Prefix:    sourceS3Prefix,
			ArtifactPath:      filepath.Join(r.ArtifactDir, "capc-manifests", r.BuildRepoHead),
			ReleaseName:       manifest,
			ReleaseS3Path:     releaseS3Path,
			ReleaseCdnURI:     cdnURI,
			ImageTagOverrides: imageTagOverrides,
			GitTag:            gitTag,
			ProjectPath:       capcProjectPath,
		}
		artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetCloudStackBundle(imageDigests map[string]string) (anywherev1alpha1.CloudStackBundle, error) {
	cloudstackBundleArtifacts := map[string][]Artifact{
		"cluster-api-provider-cloudstack": r.BundleArtifactsTable["cluster-api-provider-cloudstack"],
		"kube-rbac-proxy":                 r.BundleArtifactsTable["kube-rbac-proxy"],
	}

	var version string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	for componentName, artifacts := range cloudstackBundleArtifacts {
		for _, artifact := range artifacts {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api-provider-cloudstack" {
					componentVersion, err := BuildComponentVersion(
						newVersionerWithGITTAG(r.BuildRepoSource, capcProjectPath, imageArtifact.SourcedFromBranch, r),
					)
					if err != nil {
						return anywherev1alpha1.CloudStackBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-cloudstack")
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

	bundle := anywherev1alpha1.CloudStackBundle{
		Version:         version,
		KubeProxy:       bundleImageArtifacts["kube-rbac-proxy"],
		Manager:         bundleImageArtifacts["cluster-api-cloudstack-controller"],
		Components:      bundleManifestArtifacts["infrastructure-components.yaml"],
		ClusterTemplate: bundleManifestArtifacts["cluster-template.yaml"],
		Metadata:        bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
