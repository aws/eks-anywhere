// Copyright 2021 Amazon.com Inc. or its affiliates. All Rights Reserved.
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

// GetKindnetdAssets returns the eks-a artifacts for kindnetd
func (r *ReleaseConfig) GetKindnetdAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(kindProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "kindnetd"
	repoName := fmt.Sprintf("kubernetes-sigs/kind/%s", name)
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": kindProjectPath,
	}

	sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
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
		ProjectPath:       kindProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{Artifact{Image: imageArtifact}}

	manifestList := []string{
		"kindnetd.yaml",
	}

	for _, manifest := range manifestList {
		var sourceS3Prefix string
		var releaseS3Path string
		latestPath := getLatestUploadDestination(sourcedFromBranch)

		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Prefix = fmt.Sprintf("%s/%s/manifests/kindnetd/%s", kindProjectPath, latestPath, gitTag)
		} else {
			sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/kind/manifests/kindnetd/%s", r.BundleNumber, gitTag)
		}

		if r.DevRelease {
			releaseS3Path = fmt.Sprintf("artifacts/%s/kind/manifests/kindnetd/%s", r.DevReleaseUriVersion, gitTag)
		} else {
			releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/kind/manifests/kindnetd/%s", r.BundleNumber, gitTag)
		}

		cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, manifest))
		if err != nil {
			return nil, errors.Cause(err)
		}

		manifestArtifact := &ManifestArtifact{
			SourceS3Key:       manifest,
			SourceS3Prefix:    sourceS3Prefix,
			ArtifactPath:      filepath.Join(r.ArtifactDir, "kind-manifests", r.BuildRepoHead),
			ReleaseName:       manifest,
			ReleaseS3Path:     releaseS3Path,
			ReleaseCdnURI:     cdnURI,
			GitTag:            gitTag,
			ProjectPath:       kindProjectPath,
			SourcedFromBranch: sourcedFromBranch,
		}
		artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetKindnetdBundle() (anywherev1alpha1.KindnetdBundle, error) {
	artifacts := r.BundleArtifactsTable["kindnetd"]

	var sourceBranch string
	var componentChecksum string
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, artifact := range artifacts {
		if artifact.Manifest != nil {
			manifestArtifact := artifact.Manifest
			sourceBranch = manifestArtifact.SourcedFromBranch

			bundleManifestArtifact := anywherev1alpha1.Manifest{
				URI: manifestArtifact.ReleaseCdnURI,
			}

			bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

			manifestHash, err := r.GenerateManifestHash(manifestArtifact)
			if err != nil {
				return anywherev1alpha1.KindnetdBundle{}, err
			}
			artifactHashes = append(artifactHashes, manifestHash)
		}
	}

	if r.DryRun {
		componentChecksum = fakeComponentChecksum
	} else {
		componentChecksum = generateComponentHash(artifactHashes)
	}
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(r.BuildRepoSource, kindProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.KindnetdBundle{}, errors.Wrapf(err, "Error getting version for kind")
	}

	bundle := anywherev1alpha1.KindnetdBundle{
		Version:  version,
		Manifest: bundleManifestArtifacts["kindnetd.yaml"],
	}

	return bundle, nil
}
