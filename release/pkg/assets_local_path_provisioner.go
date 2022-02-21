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

	"github.com/pkg/errors"
)

const localPathProvisonerProjectPath = "projects/rancher/local-path-provisioner"

// GetLocalPathProvisionerAssets returns the eks-a artifacts for local-path-provisioner
func (r *ReleaseConfig) GetLocalPathProvisionerAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(localPathProvisonerProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "local-path-provisioner"
	repoName := fmt.Sprintf("rancher/%s", name)
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": localPathProvisonerProjectPath,
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
		ProjectPath:       localPathProvisonerProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{{Image: imageArtifact}}

	return artifacts, nil
}
