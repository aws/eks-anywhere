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

const cloudstackCloudProviderProjectPath = "projects/kubernetes/cloud-provider-cloudstack"

// GetCloudStackCloudProviderAssets returns the eks-a artifacts for cloudstack cloud provider
func (r *ReleaseConfig) GetCloudStackCloudProviderAssets(eksDReleaseChannel string) ([]Artifact, error) {
	gitTagFolder := filepath.Join(cloudstackCloudProviderProjectPath, eksDReleaseChannel)
	gitTag, err := r.readGitTag(gitTagFolder, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "cloud-provider-cloudstack"
	repoName := "kubernetes/cloud-provider-cloudstack/cpi/manager"
	tagOptions := map[string]string{
		"gitTag":             gitTag,
		"eksDReleaseChannel": eksDReleaseChannel,
		"projectPath":        gitTagFolder,
	}

	fmt.Println("Getting cloudStackCloudProvider source image uri")
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
		ProjectPath:       cloudstackCloudProviderProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{{Image: imageArtifact}}

	return artifacts, nil
}
