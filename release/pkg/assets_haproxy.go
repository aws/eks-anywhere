// Copyright 2021 Amazon.com Inc. or its affiliates. All Rights Reserved.
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

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// GetKindnetdAssets returns the eks-a artifacts for kindnetd
func (r *ReleaseConfig) GetHaproxyAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(kindProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "haproxy"
	repoName := fmt.Sprintf("kubernetes-sigs/kind/%s", name)
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": kindProjectPath,
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
		ProjectPath:       kindProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{{Image: imageArtifact}}

	return artifacts, nil
}

func (r *ReleaseConfig) GetHaproxyBundle(imageDigests map[string]string) (anywherev1alpha1.HaproxyBundle, error) {
	artifacts := r.BundleArtifactsTable["haproxy"]

	bundleArtifacts := map[string]anywherev1alpha1.Image{}

	for _, artifact := range artifacts {
		imageArtifact := artifact.Image
		bundleImageArtifact := anywherev1alpha1.Image{
			Name:        imageArtifact.AssetName,
			Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
			OS:          imageArtifact.OS,
			Arch:        imageArtifact.Arch,
			URI:         imageArtifact.ReleaseImageURI,
			ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
		}
		bundleArtifacts[imageArtifact.AssetName] = bundleImageArtifact
	}

	bundle := anywherev1alpha1.HaproxyBundle{
		Image: bundleArtifacts["haproxy"],
	}

	return bundle, nil
}
