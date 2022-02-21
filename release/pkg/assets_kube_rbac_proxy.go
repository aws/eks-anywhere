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

const kubeRbacProxyProjectPath = "projects/brancz/kube-rbac-proxy"

// GetKubeRbacProxyAssets returns the eks-a artifacts for kube-rbac-proxy
func (r *ReleaseConfig) GetKubeRbacProxyAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(kubeRbacProxyProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name, repoName, tagOptions := r.getKubeRbacProxyImageAttributes(gitTag)

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
		ProjectPath:       kubeRbacProxyProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{{Image: imageArtifact}}

	return artifacts, nil
}

func (r *ReleaseConfig) getKubeRbacProxyImageAttributes(gitTag string) (string, string, map[string]string) {
	name := "kube-rbac-proxy"
	repoName := fmt.Sprintf("brancz/%s", name)
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": kubeRbacProxyProjectPath,
	}

	return name, repoName, tagOptions
}

func (r *ReleaseConfig) GetKubeRbacProxyImageTagOverride() (ImageTagOverride, error) {
	gitTag, err := r.readGitTag(kubeRbacProxyProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return ImageTagOverride{}, errors.Cause(err)
	}

	name, repoName, tagOptions := r.getKubeRbacProxyImageAttributes(gitTag)

	releaseImageUri, err := r.GetReleaseImageURI(name, repoName, tagOptions)
	if err != nil {
		return ImageTagOverride{}, errors.Cause(err)
	}
	imageTagOverride := ImageTagOverride{
		Repository: repoName,
		ReleaseUri: releaseImageUri,
	}

	return imageTagOverride, nil
}
