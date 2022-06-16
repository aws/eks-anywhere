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

const eksAToolsProjectPath = "projects/aws/eks-anywhere-build-tooling"

// GetEksAToolsAssets returns the eks-a artifacts for eks-a-tools image
func (r *ReleaseConfig) GetEksAToolsAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(eksAToolsProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}
	name := "eks-anywhere-cli-tools"

	var sourceRepoName string
	var releaseRepoName string
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceRepoName = "eks-anywhere-cli-tools"
	} else {
		sourceRepoName = "cli-tools"
	}

	if r.DevRelease {
		releaseRepoName = "eks-anywhere-cli-tools"
	} else {
		releaseRepoName = "cli-tools"
	}

	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": eksAToolsProjectPath,
	}

	sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(name, sourceRepoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}
	if sourcedFromBranch != r.BuildRepoBranchName {
		gitTag, err = r.readGitTag(eksAToolsProjectPath, sourcedFromBranch)
		if err != nil {
			return nil, errors.Cause(err)
		}
		tagOptions["gitTag"] = gitTag
	}
	releaseImageUri, err := r.GetReleaseImageURI(name, releaseRepoName, tagOptions)
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
		ProjectPath:       eksAToolsProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}

	artifacts := []Artifact{Artifact{Image: imageArtifact}}

	return artifacts, nil
}

func (r *ReleaseConfig) GetEksaBundle(imageDigests map[string]string) (anywherev1alpha1.EksaBundle, error) {
	eksABundleArtifacts := map[string][]Artifact{
		"eks-a-tools":          r.BundleArtifactsTable["eks-a-tools"],
		"cluster-controller":   r.BundleArtifactsTable["cluster-controller"],
		"diagnostic-collector": r.BundleArtifactsTable["diagnostic-collector"],
	}
	sortedComponentNames := sortArtifactsMap(eksABundleArtifacts)

	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range eksABundleArtifacts[componentName] {
			if artifact.Image != nil {
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

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := r.GenerateManifestHash(manifestArtifact)
				if err != nil {
					return anywherev1alpha1.EksaBundle{}, err
				}
				artifactHashes = append(artifactHashes, manifestHash)
			}
		}
	}

	if r.DryRun {
		componentChecksum = fakeComponentChecksum
	} else {
		componentChecksum = generateComponentHash(artifactHashes)
	}
	version, err := BuildComponentVersion(newCliVersioner(r.ReleaseVersion, r.CliRepoSource), componentChecksum)
	if err != nil {
		return anywherev1alpha1.EksaBundle{}, errors.Wrapf(err, "failed generating version for eksa bundle")
	}

	bundle := anywherev1alpha1.EksaBundle{
		Version:             version,
		CliTools:            bundleImageArtifacts["eks-anywhere-cli-tools"],
		Components:          bundleManifestArtifacts["eksa-components.yaml"],
		ClusterController:   bundleImageArtifacts["eks-anywhere-cluster-controller"],
		DiagnosticCollector: bundleImageArtifacts["eks-anywhere-diagnostic-collector"],
	}

	return bundle, nil
}
