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
	"path/filepath"

	"github.com/pkg/errors"
)

// GetDiagnosticCollectorAssets returns the artifacts for eks-a diagnostic collector
func (r *ReleaseConfig) GetDiagnosticCollectorAssets() ([]Artifact, error) {
	projectSource := "projects/aws/eks-anywhere"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
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
		"gitTag": gitTag,
	}
	imageArtifact := &ImageArtifact{
		AssetName:       name,
		SourceImageURI:  r.GetSourceImageURI(name, sourceRepoName, tagOptions),
		ReleaseImageURI: r.GetReleaseImageURI(name, releaseRepoName, tagOptions),
		Arch:            []string{"amd64"},
		OS:              "linux",
	}

	artifact := Artifact{Image: imageArtifact}

	return []Artifact{artifact}, nil
}
