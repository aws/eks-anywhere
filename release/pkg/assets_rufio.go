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

const rufioProjectPath = "projects/tinkerbell/rufio"

// GetRufioAssets returns the eks a artifacts for rufio
func (r *ReleaseConfig) GetRufioAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(rufioProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "rufio"
	repoName := fmt.Sprintf("tinkerbell/%s", name)
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": rufioProjectPath,
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
		ProjectPath:       rufioProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{{Image: imageArtifact}}

	imageTagOverrides := []ImageTagOverride{
		{
			Repository: repoName,
			ReleaseUri: imageArtifact.ReleaseImageURI,
		},
	}

	manifest := "manifest.yaml"
	var sourceS3Prefix string
	var releaseS3Path string
	latestPath := getLatestUploadDestination(sourcedFromBranch)

	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceS3Prefix = fmt.Sprintf("%s/%s/manifests/infrastructure-rufio/%s", rufioProjectPath, latestPath, gitTag)
	} else {
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/rufio/manifests/infrastructure-rufio/%s", r.BundleNumber, gitTag)
	}

	if r.DevRelease {
		releaseS3Path = fmt.Sprintf("artifacts/%s/rufio/manifests/infrastructure-rufio/%s", r.DevReleaseUriVersion, gitTag)
	} else {
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/rufio/manifests/infrastructure-rufio/%s", r.BundleNumber, gitTag)
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
		ArtifactPath:      filepath.Join(r.ArtifactDir, "rufio-manifests", r.BuildRepoHead),
		ReleaseName:       manifest,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		ImageTagOverrides: imageTagOverrides,
		GitTag:            gitTag,
		ProjectPath:       rufioProjectPath,
	}
	artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})

	return artifacts, nil
}
