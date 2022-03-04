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
)

const hookProjectPath = "projects/tinkerbell/hook"

// GethookAssets returns the eks-a artifacts for tinkerbell/hook
func (r *ReleaseConfig) GethookAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(hookProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	var sourceS3Key string
	var sourceS3Prefix string
	var releaseS3Path string
	var releaseName string
	artifacts := []Artifact{}
	sourcedFromBranch := r.BuildRepoBranchName
	latestPath := getLatestUploadDestination(sourcedFromBranch)

	hookImages := []string{
		"hook-bootkit",
		"hook-docker",
		"hook-kernel",
	}
	for _, image := range hookImages {
		repoName := fmt.Sprintf("tinkerbell/%s", image)
		tagOptions := map[string]string{
			"gitTag":      gitTag,
			"projectPath": hookProjectPath,
		}

		sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}
		if sourcedFromBranch != r.BuildRepoBranchName {
			gitTag, err = r.readGitTag(hookProjectPath, sourcedFromBranch)
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
			ProjectPath:       hookProjectPath,
			SourcedFromBranch: sourcedFromBranch,
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})
	}

	hookArchives := []string{
		"initramfs-aarch64",
		"initramfs-x86_64",
		"vmlinuz-aarch64",
		"vmlinuz-x86_64",
	}
	for _, archive := range hookArchives {
		sourceS3Key = archive
		releaseName = archive
		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Prefix = fmt.Sprintf("%s/%s/%s", hookProjectPath, latestPath, gitTag)
		} else {
			sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/hook/%s", r.BundleNumber, gitTag)
		}

		if r.DevRelease {
			releaseS3Path = fmt.Sprintf("artifacts/%s/hook/%s", r.DevReleaseUriVersion, gitTag)
		} else {
			releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/hook/%s", r.BundleNumber, gitTag)
		}

		cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, releaseName))
		if err != nil {
			return nil, errors.Cause(err)
		}

		archiveArtifact := &ArchiveArtifact{
			SourceS3Key:       sourceS3Key,
			SourceS3Prefix:    sourceS3Prefix,
			ArtifactPath:      filepath.Join(r.ArtifactDir, "hook", r.BuildRepoHead),
			ReleaseName:       releaseName,
			ReleaseS3Path:     releaseS3Path,
			ReleaseCdnURI:     cdnURI,
			GitTag:            gitTag,
			ProjectPath:       hookProjectPath,
			SourcedFromBranch: sourcedFromBranch,
			ImageFormat:       "kernel",
		}
		artifacts = append(artifacts, Artifact{Archive: archiveArtifact})
	}

	return artifacts, nil
}
