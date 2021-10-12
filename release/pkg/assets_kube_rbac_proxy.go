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

// GetKubeRbacProxyAssets returns the eks-a artifacts for kube-rbac-proxy
func (r *ReleaseConfig) GetKubeRbacProxyAssets() ([]Artifact, error) {
	// Get Git tag for the project
	gitTag, err := r.getKubeRbacProxyGitTag()
	if err != nil {
		return nil, errors.Cause(err)
	}

	name, repoName, tagOptions := r.getKubeRbacProxyImageAttributes(gitTag)

	imageArtifact := &ImageArtifact{
		AssetName:       name,
		SourceImageURI:  r.GetSourceImageURI(name, repoName, tagOptions),
		ReleaseImageURI: r.GetReleaseImageURI(name, repoName, tagOptions),
		Arch:            []string{"amd64"},
		OS:              "linux",
	}

	artifact := &Artifact{Image: imageArtifact}

	return []Artifact{*artifact}, nil
}

func (r *ReleaseConfig) getKubeRbacProxyGitTag() (string, error) {
	projectSource := "projects/brancz/kube-rbac-proxy"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return "", errors.Cause(err)
	}

	return gitTag, nil
}

func (r *ReleaseConfig) getKubeRbacProxyImageAttributes(gitTag string) (string, string, map[string]string) {
	name := "kube-rbac-proxy"
	repoName := fmt.Sprintf("brancz/%s", name)
	tagOptions := map[string]string{
		"gitTag": gitTag,
	}

	return name, repoName, tagOptions
}

func (r *ReleaseConfig) GetKubeRbacProxyImageTagOverride() (ImageTagOverride, error) {
	gitTag, err := r.getKubeRbacProxyGitTag()
	if err != nil {
		return ImageTagOverride{}, errors.Cause(err)
	}

	name, repoName, tagOptions := r.getKubeRbacProxyImageAttributes(gitTag)

	imageTagOverride := ImageTagOverride{
		Repository: repoName,
		ReleaseUri: r.GetReleaseImageURI(name, repoName, tagOptions),
	}

	return imageTagOverride, nil
}
