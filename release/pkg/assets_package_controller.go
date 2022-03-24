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

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	packagesRootPath = "projects/aws/eks-anywhere-packages"
)

// GetPackagesAssets returns the eks-a artifacts for package-controller
func (r *ReleaseConfig) GetPackagesAssets() ([]Artifact, error) {
	name := "eks-anywhere-packages"
	repoName := name
	gitTag, err := r.readGitTag(packagesRootPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": packagesRootPath,
	}

	sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}
	if sourcedFromBranch != r.BuildRepoBranchName {
		gitTag, err = r.readGitTag(packagesRootPath, sourcedFromBranch)
		if err != nil {
			return nil, errors.Cause(err)
		}
		tagOptions["gitTag"] = gitTag
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
		ProjectPath:       packagesRootPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{{Image: imageArtifact}}
	return artifacts, nil
}

func (r *ReleaseConfig) GetPackagesBundle(imageDigests map[string]string) (anywherev1alpha1.PackageBundle, error) {
	artifacts := r.BundleArtifactsTable["eks-anywhere-packages"]

	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	artifactHashes := []string{}

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

		bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
		artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
	}

	componentChecksum := generateComponentHash(artifactHashes)
	version, err := BuildComponentVersion(newCliVersioner(r.ReleaseVersion, r.CliRepoSource), componentChecksum)
	if err != nil {
		return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "failed generating version for package bundle")
	}

	bundle := anywherev1alpha1.PackageBundle{
		Version:    version,
		Controller: bundleImageArtifacts["eks-anywhere-packages"],
	}
	return bundle, nil
}
