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

const vSphereCsiProjectPath = "projects/kubernetes-sigs/vsphere-csi-driver"

// GetVsphereCsiAssets returns the eks-a artifacts for vSphere CSI
func (r *ReleaseConfig) GetVsphereCsiAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(vSphereCsiProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	vsphereCsiImages := []string{
		"driver",
		"syncer",
	}

	artifacts := []Artifact{}
	for _, image := range vsphereCsiImages {
		repoName := fmt.Sprintf("kubernetes-sigs/vsphere-csi-driver/csi/%s", image)
		tagOptions := map[string]string{
			"gitTag":      gitTag,
			"projectPath": vSphereCsiProjectPath,
		}

		sourceImageUri, err := r.GetSourceImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}
		releaseImageUri, err := r.GetReleaseImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}

		imageArtifact := &ImageArtifact{
			AssetName:       fmt.Sprintf("vsphere-csi-%s", image),
			SourceImageURI:  sourceImageUri,
			ReleaseImageURI: releaseImageUri,
			Arch:            []string{"amd64"},
			OS:              "linux",
			GitTag:          gitTag,
			ProjectPath:     vSphereCsiProjectPath,
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})
	}

	return artifacts, nil
}
