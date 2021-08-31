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

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/pkg/errors"
)

// GetBottlerocketBootstrapAssets returns the eks-a artifacts for Bottlerocket bootstrap container
func (r *ReleaseConfig) GetBottlerocketBootstrapAssets(eksDReleaseChannel, eksDReleaseNumber string) ([]Artifact, error) {
	name := "bottlerocket-bootstrap"
	repoName := name
	tagOptions := map[string]string{
		"eksDReleaseChannel": eksDReleaseChannel,
		"eksDReleaseNumber":  eksDReleaseNumber,
	}

	imageArtifact := &ImageArtifact{
		AssetName:       name,
		SourceImageURI:  r.GetSourceImageURI(name, repoName, tagOptions),
		ReleaseImageURI: r.GetReleaseImageURI(name, repoName, tagOptions),
		Arch:            []string{"amd64"},
		OS:              "linux",
	}

	artifact := Artifact{Image: imageArtifact}

	return []Artifact{artifact}, nil
}

func (r *ReleaseConfig) GetBottlerocketBootstrapBundle(eksDReleaseChannel, eksDReleaseNumber string, imageDigests map[string]string) (anywherev1alpha1.BottlerocketBootstrapBundle, error) {
	artifacts, err := r.GetBottlerocketBootstrapAssets(eksDReleaseChannel, eksDReleaseNumber)
	if err != nil {
		return anywherev1alpha1.BottlerocketBootstrapBundle{}, errors.Cause(err)
	}

	bundleArtifacts := map[string]anywherev1alpha1.Image{}

	for _, artifact := range artifacts {
		imageArtifact := artifact.Image
		bottlerocketBootstrapImage := anywherev1alpha1.Image{
			Name:        imageArtifact.AssetName,
			Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
			OS:          imageArtifact.OS,
			Arch:        imageArtifact.Arch,
			URI:         imageArtifact.ReleaseImageURI,
			ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
		}
		bundleArtifacts[imageArtifact.AssetName] = bottlerocketBootstrapImage
	}

	bundle := anywherev1alpha1.BottlerocketBootstrapBundle{
		Bootstrap: bundleArtifacts["bottlerocket-bootstrap"],
	}

	return bundle, nil
}
