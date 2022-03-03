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
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const harborProjectPath = "projects/goharbor/harbor"

// GetHarborAssets returns the eks-a artifacts for harbor
func (r *ReleaseConfig) GetHarborAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(harborProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	harborImages := []string{
		"harbor-core",
		"harbor-db",
		"harbor-exporter",
		"harbor-jobservice",
		"harbor-log",
		"harbor-nginx",
		"harbor-portal",
		"harbor-redis",
		"harbor-registry",
		"harbor-registryctl",
	}

	artifacts := []Artifact{}
	for _, image := range harborImages {
		repoName := fmt.Sprintf("goharbor/%s", image)
		tagOptions := map[string]string{
			"gitTag":      gitTag,
			"projectPath": harborProjectPath,
		}

		sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}
		if sourcedFromBranch != r.BuildRepoBranchName {
			gitTag, err = r.readGitTag(harborProjectPath, sourcedFromBranch)
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
			ProjectPath:       harborProjectPath,
			SourcedFromBranch: sourcedFromBranch,
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})
	}

	return artifacts, nil
}
