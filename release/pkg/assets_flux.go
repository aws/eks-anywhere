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

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// GetFluxAssets returns the eks-a artifacts for Flux
func (r *ReleaseConfig) GetFluxAssets() ([]Artifact, error) {
	fluxControllerProjects := []string{"source-controller", "kustomize-controller", "helm-controller", "notification-controller"}
	artifacts := []Artifact{}

	for _, project := range fluxControllerProjects {
		fluxControllerProjectPath := fmt.Sprintf("projects/fluxcd/%s", project)
		gitTag, err := r.readGitTag(fluxControllerProjectPath, r.BuildRepoBranchName)
		if err != nil {
			return nil, errors.Cause(err)
		}
		repoName := fmt.Sprintf("fluxcd/%s", project)
		tagOptions := map[string]string{
			"gitTag":      gitTag,
			"projectPath": fluxControllerProjectPath,
		}

		sourceImageUri, err := r.GetSourceImageURI(project, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}
		releaseImageUri, err := r.GetReleaseImageURI(project, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}

		imageArtifact := &ImageArtifact{
			AssetName:       project,
			SourceImageURI:  sourceImageUri,
			ReleaseImageURI: releaseImageUri,
			Arch:            []string{"amd64"},
			OS:              "linux",
			GitTag:          gitTag,
			ProjectPath:     fluxControllerProjectPath,
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})
	}
	return artifacts, nil
}

func (r *ReleaseConfig) GetFluxBundle(imageDigests map[string]string) (anywherev1alpha1.FluxBundle, error) {
	artifacts := r.BundleArtifactsTable["flux"]

	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
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
	}

	version, err := BuildComponentVersion(
		newMultiProjectVersionerWithGITTAG(
			filepath.Join(r.BuildRepoSource, "projects/fluxcd"),
			filepath.Join(r.BuildRepoSource, "projects/fluxcd/flux2"),
		),
	)
	if err != nil {
		return anywherev1alpha1.FluxBundle{}, errors.Wrap(err, "failed generating version for flux bundle")
	}

	bundle := anywherev1alpha1.FluxBundle{
		Version:                version,
		SourceController:       bundleImageArtifacts["source-controller"],
		KustomizeController:    bundleImageArtifacts["kustomize-controller"],
		HelmController:         bundleImageArtifacts["helm-controller"],
		NotificationController: bundleImageArtifacts["notification-controller"],
	}

	return bundle, nil
}
