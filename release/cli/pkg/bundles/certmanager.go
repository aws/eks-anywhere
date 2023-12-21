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

package bundles

import (
	"fmt"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	"github.com/aws/eks-anywhere/release/cli/pkg/version"
)

func GetCertManagerBundle(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.CertManagerBundle, error) {
	certManagerArtifacts, err := r.BundleArtifactsTable.Load("cert-manager")
	if err != nil {
		return anywherev1alpha1.CertManagerBundle{}, fmt.Errorf("artifacts for project cert-manager not found in bundle artifacts table")
	}

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, artifact := range certManagerArtifacts {
		if artifact.Image != nil {
			imageArtifact := artifact.Image
			sourceBranch = imageArtifact.SourcedFromBranch
			imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
			if err != nil {
				return anywherev1alpha1.CertManagerBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
			}
			bundleArtifact := anywherev1alpha1.Image{
				Name:        imageArtifact.AssetName,
				Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
				OS:          imageArtifact.OS,
				Arch:        imageArtifact.Arch,
				URI:         imageArtifact.ReleaseImageURI,
				ImageDigest: imageDigest,
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

			manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
			if err != nil {
				return anywherev1alpha1.CertManagerBundle{}, err
			}

			artifactHashes = append(artifactHashes, manifestHash)
		}
	}

	if r.DryRun {
		componentChecksum = version.FakeComponentChecksum
	} else {
		componentChecksum = version.GenerateComponentHash(artifactHashes, r.DryRun)
	}
	version, err := version.BuildComponentVersion(
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CertManagerProjectPath, sourceBranch, r),
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
		Ctl:        bundleImageArtifacts["cert-manager-ctl"],
		Webhook:    bundleImageArtifacts["cert-manager-webhook"],
		Manifest:   bundleManifestArtifacts["cert-manager.yaml"],
	}

	return bundle, nil
}
