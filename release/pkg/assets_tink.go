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

const tinkProjectPath = "projects/tinkerbell/tink"

// GetTinkAssets returns the eks-a artifacts for tinkerbell/tink
func (r *ReleaseConfig) GetTinkAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(tinkProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	tinkImages := []string{
		"tink-controller",
		"tink-server",
		"tink-worker",
		"tink-cli",
	}

	artifacts := []Artifact{}
	imageTagOverrides := []ImageTagOverride{}
	sourceBranch := ""
	for _, image := range tinkImages {
		repoName := fmt.Sprintf("tinkerbell/tink/%s", image)
		tagOptions := map[string]string{
			"gitTag":      gitTag,
			"projectPath": tinkProjectPath,
		}

		sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}
		if sourcedFromBranch != r.BuildRepoBranchName {
			gitTag, err = r.readGitTag(tinkProjectPath, sourcedFromBranch)
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
			ProjectPath:       tinkProjectPath,
			SourcedFromBranch: sourcedFromBranch,
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})

		imageTagOverrides = append(imageTagOverrides, ImageTagOverride{
			Repository: repoName,
			ReleaseUri: imageArtifact.ReleaseImageURI,
		})
		sourceBranch = sourcedFromBranch
	}

	manifest := "tink.yaml"
	var sourceS3Prefix string
	var releaseS3Path string
	latestPath := getLatestUploadDestination(sourceBranch)

	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceS3Prefix = fmt.Sprintf("%s/%s", tinkProjectPath, latestPath)
	} else {
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/tink", r.BundleNumber)
	}

	if r.DevRelease {
		releaseS3Path = fmt.Sprintf("artifacts/%s/tink", r.DevReleaseUriVersion)
	} else {
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/tink", r.BundleNumber)
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
		ArtifactPath:      filepath.Join(r.ArtifactDir, "tink", r.BuildRepoHead),
		ReleaseName:       manifest,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		ImageTagOverrides: imageTagOverrides,
		GitTag:            gitTag,
		ProjectPath:       tinkProjectPath,
	}
	artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})

	return artifacts, nil
}
