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
	"github.com/pkg/errors"
)

// GetDiagnosticCollectorAssets returns the artifacts for eks-a diagnostic collector
func (r *ReleaseConfig) GetDiagnosticCollectorAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(eksAnywhereProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}
	name := "eks-anywhere-diagnostic-collector"

	var sourceRepoName string
	var releaseRepoName string
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceRepoName = "eks-anywhere-diagnostic-collector"
	} else {
		sourceRepoName = "diagnostic-collector"
	}

	if r.DevRelease {
		releaseRepoName = "eks-anywhere-diagnostic-collector"
	} else {
		releaseRepoName = "diagnostic-collector"
	}

	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": eksAnywhereProjectPath,
	}

	sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(name, sourceRepoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}
	if sourcedFromBranch != r.BuildRepoBranchName {
		gitTag, err = r.readGitTag(eksAnywhereProjectPath, sourcedFromBranch)
		if err != nil {
			return nil, errors.Cause(err)
		}
		tagOptions["gitTag"] = gitTag
	}
	releaseImageUri, err := r.GetReleaseImageURI(name, releaseRepoName, tagOptions)
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
		ProjectPath:       eksAnywhereProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}

	artifacts := []Artifact{{Image: imageArtifact}}

	return artifacts, nil
}
