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

const bootsProjectPath = "projects/tinkerbell/boots"

// GetBootsAssets returns the eks a artifacts for boots
func (r *ReleaseConfig) GetBootsAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(bootsProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "boots"
	repoName := fmt.Sprintf("tinkerbell/%s", name)
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": bootsProjectPath,
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
		ProjectPath:       bootsProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}

	return []Artifact{{Image: imageArtifact}}, nil
}
