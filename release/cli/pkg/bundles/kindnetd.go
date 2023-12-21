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

func GetKindnetdBundle(r *releasetypes.ReleaseConfig) (anywherev1alpha1.KindnetdBundle, error) {
	kindnetdArtifacts, err := r.BundleArtifactsTable.Load("kindnetd")
	if err != nil {
		return anywherev1alpha1.KindnetdBundle{}, fmt.Errorf("artifacts for project kindnetd not found in bundle artifacts table")
	}

	var sourceBranch string
	var componentChecksum string
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, artifact := range kindnetdArtifacts {
		if artifact.Manifest != nil {
			manifestArtifact := artifact.Manifest
			sourceBranch = manifestArtifact.SourcedFromBranch

			bundleManifestArtifact := anywherev1alpha1.Manifest{
				URI: manifestArtifact.ReleaseCdnURI,
			}

			bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

			manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
			if err != nil {
				return anywherev1alpha1.KindnetdBundle{}, err
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
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.KindProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.KindnetdBundle{}, errors.Wrapf(err, "Error getting version for kind")
	}

	bundle := anywherev1alpha1.KindnetdBundle{
		Version:  version,
		Manifest: bundleManifestArtifacts["kindnetd.yaml"],
	}

	return bundle, nil
}
