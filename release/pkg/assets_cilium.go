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

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/pkg/errors"
)

// GetCiliumAssets returns the eks-a artifacts for Cilium
func (r *ReleaseConfig) GetCiliumAssets() ([]Artifact, error) {
	gitTag, err := r.getCiliumGitTag()
	if err != nil {
		return nil, errors.Cause(err)
	}

	manifestName := "cilium.yaml"
	artifacts := []Artifact{}

	var sourceS3Prefix string
	var releaseS3Path string

	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceS3Prefix = fmt.Sprintf("projects/cilium/cilium/latest/manifests/cilium/%s", gitTag)
	} else {
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/cilium/manifests/cilium/%s", r.BundleNumber, gitTag)
	}

	if r.DevRelease {
		releaseS3Path = fmt.Sprintf("artifacts/%s/cilium/manifests/cilium/%s", r.DevReleaseUriVersion, gitTag)
	} else {
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/cilium/manifests/cilium/%s", r.BundleNumber, gitTag)
	}

	cdnURI, err := r.GetURI(filepath.Join(
		releaseS3Path,
		manifestName))
	if err != nil {
		return nil, errors.Cause(err)
	}

	manifestArtifact := &ManifestArtifact{
		SourceS3Key:       manifestName,
		SourceS3Prefix:    sourceS3Prefix,
		ArtifactPath:      filepath.Join(r.ArtifactDir, "cilium-manifests", r.BuildRepoHead),
		ReleaseName:       manifestName,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		ImageTagOverrides: []ImageTagOverride{},
	}
	artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})

	return artifacts, nil
}

func (r *ReleaseConfig) GetCiliumBundle(imageDigests map[string]string) (anywherev1alpha1.CiliumBundle, error) {
	artifacts, err := r.GetCiliumAssets()
	if err != nil {
		return anywherev1alpha1.CiliumBundle{}, errors.Cause(err)
	}

	ciliumContainerRegistry := "public.ecr.aws/isovalent"
	ciliumGitTag, err := r.getCiliumGitTag()
	if err != nil {
		return anywherev1alpha1.CiliumBundle{}, errors.Cause(err)
	}
	ciliumImageTagMap := map[string]string{
		"cilium":           ciliumGitTag,
		"operator-generic": "v1.9.9-beta1",
	}

	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}

	for image, tag := range ciliumImageTagMap {
		imageDigest, err := r.getCiliumImageDigest(image)
		if err != nil {
			return anywherev1alpha1.CiliumBundle{}, errors.Cause(err)
		}

		bundleImageArtifact := anywherev1alpha1.Image{
			Name:        image,
			Description: fmt.Sprintf("Container image for %s image", image),
			OS:          "linux",
			Arch:        []string{"amd64"},
			URI:         fmt.Sprintf("%s/%s-eksa:%s", ciliumContainerRegistry, image, tag),
			ImageDigest: imageDigest,
		}

		bundleImageArtifacts[image] = bundleImageArtifact
	}

	for _, artifact := range artifacts {
		if artifact.Manifest != nil {
			manifestArtifact := artifact.Manifest
			bundleManifestArtifact := anywherev1alpha1.Manifest{
				URI: manifestArtifact.ReleaseCdnURI,
			}

			bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
		}
	}

	bundle := anywherev1alpha1.CiliumBundle{
		Cilium:   bundleImageArtifacts["cilium"],
		Operator: bundleImageArtifacts["operator-generic"],
		Manifest: bundleManifestArtifacts["cilium.yaml"],
	}

	return bundle, nil
}

func (r *ReleaseConfig) getCiliumGitTag() (string, error) {
	projectSource := "projects/cilium/cilium"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return "", errors.Cause(err)
	}

	return gitTag, nil
}

func (r *ReleaseConfig) getCiliumImageDigest(imageName string) (string, error) {
	projectSource := "projects/cilium/cilium"
	imageDigestFileName := fmt.Sprintf("images/%s/IMAGE_DIGEST", imageName)
	imageDigestFile := filepath.Join(r.BuildRepoSource, projectSource, imageDigestFileName)
	imageDigest, err := readFile(imageDigestFile)
	if err != nil {
		return "", errors.Cause(err)
	}

	return imageDigest, nil
}
