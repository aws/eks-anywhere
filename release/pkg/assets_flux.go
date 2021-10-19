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
	fluxGitTag, err := r.getFluxGitTag("flux2")
	if err != nil {
		return nil, errors.Cause(err)
	}
	for _, project := range fluxControllerProjects {
		gitTag, err := r.getFluxGitTag(project)
		if err != nil {
			return nil, errors.Cause(err)
		}

		repoName, tagOptions := r.getFluxImageAttributes(project, gitTag)

		imageArtifact := &ImageArtifact{
			AssetName:       project,
			SourceImageURI:  r.GetSourceImageURI(project, repoName, tagOptions),
			ReleaseImageURI: r.GetReleaseImageURI(project, repoName, tagOptions),
			Arch:            []string{"amd64"},
			OS:              "linux",
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})
	}

	manifest := "gotk-components.yaml"

	imageTagOverrides, err := r.getFluxControllerTagOverrides(fluxControllerProjects)
	if err != nil {
		return nil, errors.Cause(err)
	}

	var sourceS3Prefix string
	var releaseS3Path string

	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceS3Prefix = fmt.Sprintf("projects/fluxcd/flux2/latest/manifests/gotk/%s", fluxGitTag)
	} else {
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/flux2/manifests/gotk/%s", r.BundleNumber, fluxGitTag)
	}

	if r.DevRelease {
		releaseS3Path = fmt.Sprintf("artifacts/%s/flux2/manifests/gotk/%s", r.DevReleaseUriVersion, fluxGitTag)
	} else {
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/flux2/manifests/gotk/%s", r.BundleNumber, fluxGitTag)
	}

	cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, manifest))
	if err != nil {
		return nil, errors.Cause(err)
	}

	manifestArtifact := &ManifestArtifact{
		SourceS3Key:       manifest,
		SourceS3Prefix:    sourceS3Prefix,
		ArtifactPath:      filepath.Join(r.ArtifactDir, "flux-manifests", r.BuildRepoHead),
		ReleaseName:       manifest,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		ImageTagOverrides: imageTagOverrides,
	}
	artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})

	return artifacts, nil
}

func (r *ReleaseConfig) GetFluxBundle(imageDigests map[string]string) (anywherev1alpha1.FluxBundle, error) {
	artifacts, err := r.GetFluxAssets()
	if err != nil {
		return anywherev1alpha1.FluxBundle{}, errors.Cause(err)
	}

	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	for _, artifact := range artifacts {
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
		}

		if artifact.Manifest != nil {
			manifestArtifact := artifact.Manifest
			bundleManifestArtifact := anywherev1alpha1.Manifest{
				URI: manifestArtifact.ReleaseCdnURI,
			}

			bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
		}
	}

	version, err := r.GenerateComponentBundleVersion(
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
		Components:             bundleManifestArtifacts["gotk-components.yaml"],
	}

	return bundle, nil
}

func (r *ReleaseConfig) getFluxGitTag(project string) (string, error) {
	projectSource := fmt.Sprintf("projects/fluxcd/%s", project)
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return "", errors.Cause(err)
	}

	return gitTag, nil
}

func (r *ReleaseConfig) getFluxControllerTagOverrides(projects []string) ([]ImageTagOverride, error) {
	imageTagOverrides := []ImageTagOverride{}
	for _, project := range projects {
		gitTag, err := r.getFluxGitTag(project)
		if err != nil {
			return nil, errors.Cause(err)
		}

		repoName, tagOptions := r.getFluxImageAttributes(project, gitTag)

		imageTagOverride := ImageTagOverride{
			Repository: repoName,
			ReleaseUri: r.GetReleaseImageURI(project, repoName, tagOptions),
		}

		imageTagOverrides = append(imageTagOverrides, imageTagOverride)
	}

	return imageTagOverrides, nil
}

func (r *ReleaseConfig) getFluxImageAttributes(project, gitTag string) (string, map[string]string) {
	repoName := fmt.Sprintf("fluxcd/%s", project)
	tagOptions := map[string]string{
		"gitTag": gitTag,
	}

	return repoName, tagOptions
}
