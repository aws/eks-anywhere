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

// GetLocalPathProvisionerAssets returns the eks-a artifacts for local-path-provisioner
func (r *ReleaseConfig) GetLocalPathProvisionerAssets() ([]Artifact, error) {
	// Get Git tag for the project
	projectSource := "projects/rancher/local-path-provisioner"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "local-path-provisioner"
	repoName := fmt.Sprintf("rancher/%s", name)
	tagOptions := map[string]string{
		"gitTag": gitTag,
	}
	releaseImageUri, err := r.GetReleaseImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}

	imageArtifact := &ImageArtifact{
		AssetName:       name,
		SourceImageURI:  r.GetSourceImageURI(name, repoName, tagOptions),
		ReleaseImageURI: releaseImageUri,
		Arch:            []string{"amd64"},
		OS:              "linux",
		GitTag:          gitTag,
	}
	artifacts := []Artifact{Artifact{Image: imageArtifact}}

	return artifacts, nil
}
