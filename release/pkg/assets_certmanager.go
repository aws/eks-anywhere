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

// GetCertManagerAssets returns the eks-a artifacts for certmanager
func (r *ReleaseConfig) GetCertManagerAssets() ([]Artifact, error) {
	projectSource := "projects/jetstack/cert-manager"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return nil, errors.Cause(err)
	}

	certImages := []string{
		"cert-manager-acmesolver",
		"cert-manager-cainjector",
		"cert-manager-controller",
		"cert-manager-webhook",
	}

	artifacts := []Artifact{}
	for _, image := range certImages {
		repoName := fmt.Sprintf("jetstack/%s", image)
		tagOptions := map[string]string{
			"gitTag": gitTag,
		}

		imageArtifact := &ImageArtifact{
			AssetName:       image,
			SourceImageURI:  r.GetSourceImageURI(image, repoName, tagOptions),
			ReleaseImageURI: r.GetReleaseImageURI(image, repoName, tagOptions),
			Arch:            []string{"amd64"},
			OS:              "linux",
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetCertManagerBundle(imageDigests map[string]string) (anywherev1alpha1.CertManagerBundle, error) {
	artifacts, err := r.GetCertManagerAssets()
	if err != nil {
		return anywherev1alpha1.CertManagerBundle{}, errors.Cause(err)
	}

	bundleArtifacts := map[string]anywherev1alpha1.Image{}
	bundleObjects := []string{}

	for _, artifact := range artifacts {
		imageArtifact := artifact.Image
		bundleArtifact := anywherev1alpha1.Image{
			Name:        imageArtifact.AssetName,
			Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
			OS:          imageArtifact.OS,
			Arch:        imageArtifact.Arch,
			URI:         imageArtifact.ReleaseImageURI,
			ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
		}

		bundleArtifacts[imageArtifact.AssetName] = bundleArtifact
		bundleObjects = append(bundleObjects, bundleArtifact.ImageDigest)
	}

	componentChecksum := GenerateComponentChecksum(bundleObjects)
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, "projects/jetstack/cert-manager")),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.CertManagerBundle{}, errors.Wrapf(err, "Error getting version for cert-manager")
	}

	bundle := anywherev1alpha1.CertManagerBundle{
		Version:    version,
		Acmesolver: bundleArtifacts["cert-manager-acmesolver"],
		Cainjector: bundleArtifacts["cert-manager-cainjector"],
		Controller: bundleArtifacts["cert-manager-controller"],
		Webhook:    bundleArtifacts["cert-manager-webhook"],
	}

	return bundle, nil
}
