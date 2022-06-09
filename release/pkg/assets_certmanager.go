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

const certManagerProjectPath = "projects/jetstack/cert-manager"

// GetCertManagerAssets returns the eks-a artifacts for certmanager
func (r *ReleaseConfig) GetCertManagerAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(certManagerProjectPath, r.BuildRepoBranchName)
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
			"gitTag":      gitTag,
			"projectPath": certManagerProjectPath,
		}

		sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}
		if sourcedFromBranch != r.BuildRepoBranchName {
			gitTag, err = r.readGitTag(certManagerProjectPath, sourcedFromBranch)
			if err != nil {
				return nil, errors.Cause(err)
			}
			tagOptions["gitTag"] = gitTag
		}
		releaseImageUri, err := r.GetReleaseImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}

		imageArtifact := &ImageArtifact{
			AssetName:         image,
			SourceImageURI:    sourceImageUri,
			ReleaseImageURI:   releaseImageUri,
			Arch:              []string{"amd64"},
			OS:                "linux",
			GitTag:            gitTag,
			ProjectPath:       certManagerProjectPath,
			SourcedFromBranch: sourcedFromBranch,
		}
		artifacts = append(artifacts, Artifact{Image: imageArtifact})
	}

	manifestName := "cert-manager.yaml"

	var sourceS3Prefix string
	var releaseS3Path string
	sourcedFromBranch := r.BuildRepoBranchName
	latestPath := getLatestUploadDestination(sourcedFromBranch)

	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceS3Prefix = fmt.Sprintf("%s/%s/manifests/%s", certManagerProjectPath, latestPath, gitTag)
	} else {
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/cert-manager/manifests/%s", r.BundleNumber, gitTag)
	}

	if r.DevRelease {
		releaseS3Path = fmt.Sprintf("artifacts/%s/cert-manager/manifests/%s", r.DevReleaseUriVersion, gitTag)
	} else {
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/cert-manager/manifests/%s", r.BundleNumber, gitTag)
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
		ArtifactPath:      filepath.Join(r.ArtifactDir, "cert-manager-manifests", r.BuildRepoHead),
		ReleaseName:       manifestName,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		ImageTagOverrides: []ImageTagOverride{},
		GitTag:            gitTag,
		ProjectPath:       certManagerProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})

	return artifacts, nil
}

func (r *ReleaseConfig) GetCertManagerBundle(imageDigests map[string]string) (anywherev1alpha1.CertManagerBundle, error) {
	artifacts := r.BundleArtifactsTable["cert-manager"]

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, artifact := range artifacts {
		if artifact.Image != nil {
			imageArtifact := artifact.Image
			sourceBranch = imageArtifact.SourcedFromBranch

			bundleArtifact := anywherev1alpha1.Image{
				Name:        imageArtifact.AssetName,
				Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
				OS:          imageArtifact.OS,
				Arch:        imageArtifact.Arch,
				URI:         imageArtifact.ReleaseImageURI,
				ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
			}

			bundleImageArtifacts[imageArtifact.AssetName] = bundleArtifact
			artifactHashes = append(artifactHashes, bundleArtifact.ImageDigest)
		}
		if artifact.Manifest != nil {
			manifestArtifact := artifact.Manifest
			bundleManifestArtifact := anywherev1alpha1.Manifest{
				URI: manifestArtifact.ReleaseCdnURI,
			}

			bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

			manifestHash, err := r.GenerateManifestHash(manifestArtifact)
			if err != nil {
				return anywherev1alpha1.CertManagerBundle{}, err
			}
			artifactHashes = append(artifactHashes, manifestHash)
		}
	}

	if r.DryRun {
		componentChecksum = fakeComponentChecksum
	} else {
		componentChecksum = generateComponentHash(artifactHashes)
	}
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(r.BuildRepoSource, certManagerProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.CertManagerBundle{}, errors.Wrapf(err, "Error getting version for cert-manager")
	}

	bundle := anywherev1alpha1.CertManagerBundle{
		Version:    version,
		Acmesolver: bundleImageArtifacts["cert-manager-acmesolver"],
		Cainjector: bundleImageArtifacts["cert-manager-cainjector"],
		Controller: bundleImageArtifacts["cert-manager-controller"],
		Webhook:    bundleImageArtifacts["cert-manager-webhook"],
		Manifest:   bundleManifestArtifacts["cert-manager.yaml"],
	}

	return bundle, nil
}
