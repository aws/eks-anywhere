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

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const capasProjectPath = "projects/aws/cluster-api-provider-aws-snow"

// GetCapasAssets returns the eks-a artifacts for cluster-api-provider-aws-snow
func (r *ReleaseConfig) GetCapasAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(capasProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	artifacts := []Artifact{}

	name := "cluster-api-snow-controller"
	repoName := "aws/cluster-api-provider-aws-snow/manager"
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": capasProjectPath,
	}

	sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}
	if sourcedFromBranch != r.BuildRepoBranchName {
		gitTag, err = r.readGitTag(capasProjectPath, sourcedFromBranch)
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
		ProjectPath:       capasProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts = append(artifacts, Artifact{Image: imageArtifact})

	imageTagOverrides := []ImageTagOverride{
		{
			Repository: repoName,
			ReleaseUri: imageArtifact.ReleaseImageURI,
		},
	}

	manifestList := []string{
		"infrastructure-components.yaml",
		"metadata.yaml",
	}

	for _, manifest := range manifestList {
		var sourceS3Prefix string
		var releaseS3Path string
		sourcedFromBranch := r.BuildRepoBranchName
		latestPath := getLatestUploadDestination(sourcedFromBranch)

		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Prefix = fmt.Sprintf("%s/%s/manifests/infrastructure-snow/%s", capasProjectPath, latestPath, gitTag)
		} else {
			sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/cluster-api-provider-aws-snow/manifests/infrastructure-snow/%s", r.BundleNumber, gitTag)
		}

		if r.DevRelease {
			releaseS3Path = fmt.Sprintf("artifacts/%s/cluster-api-provider-aws-snow/manifests/infrastructure-snow/%s", r.DevReleaseUriVersion, gitTag)
		} else {
			releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/cluster-api-provider-aws-snow/manifests/infrastructure-snow/%s", r.BundleNumber, gitTag)
		}

		cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, manifest))
		if err != nil {
			return nil, errors.Cause(err)
		}

		manifestArtifact := &ManifestArtifact{
			SourceS3Key:       manifest,
			SourceS3Prefix:    sourceS3Prefix,
			ArtifactPath:      filepath.Join(r.ArtifactDir, "capas-manifests", r.BuildRepoHead),
			ReleaseName:       manifest,
			ReleaseS3Path:     releaseS3Path,
			ReleaseCdnURI:     cdnURI,
			GitTag:            gitTag,
			ProjectPath:       capasProjectPath,
			SourcedFromBranch: sourcedFromBranch,
			Component:         "cluster-api-provider-aws-snow",
			ImageTagOverrides: imageTagOverrides,
		}
		artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetSnowBundle(imageDigests map[string]string) (anywherev1alpha1.SnowBundle, error) {
	capasBundleArtifacts := map[string][]Artifact{
		"cluster-api-provider-aws-snow": r.BundleArtifactsTable["cluster-api-provider-aws-snow"],
		"kube-rbac-proxy":               r.BundleArtifactsTable["kube-rbac-proxy"],
		"kube-vip":                      r.BundleArtifactsTable["kube-vip"],
	}
	sortedComponentNames := sortArtifactsMap(capasBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range capasBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image

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
				if componentName == "cluster-api-provider-aws-snow" {
					sourceBranch = manifestArtifact.SourcedFromBranch
				}
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := r.GenerateManifestHash(manifestArtifact)
				if err != nil {
					return anywherev1alpha1.SnowBundle{}, err
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
		newVersionerWithGITTAG(r.BuildRepoSource, capasProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.SnowBundle{}, errors.Wrapf(err, "Error getting version for CAPAS")
	}

	bundle := anywherev1alpha1.SnowBundle{
		Version:    version,
		Manager:    bundleImageArtifacts["cluster-api-snow-controller"],
		KubeVip:    bundleImageArtifacts["kube-vip"],
		Components: bundleManifestArtifacts["infrastructure-components.yaml"],
		Metadata:   bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
