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

package operations

import (
	"fmt"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/pkg/assets"
	"github.com/aws/eks-anywhere/release/pkg/aws/ecrpublic"
	"github.com/aws/eks-anywhere/release/pkg/bundles"
	"github.com/aws/eks-anywhere/release/pkg/constants"
	"github.com/aws/eks-anywhere/release/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/pkg/util/artifacts"
)

func GenerateBundleArtifactsTable(r *releasetypes.ReleaseConfig) (map[string][]releasetypes.Artifact, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("              Bundle Artifacts Table Generation")
	fmt.Println("==========================================================")

	eksDReleaseMap, err := filereader.ReadEksDReleases(r)
	if err != nil {
		return nil, err
	}

	supportedK8sVersions, err := filereader.GetSupportedK8sVersions(r)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting supported Kubernetes versions for bottlerocket")
	}

	artifactsTable, err := assets.GetBundleReleaseAssets(supportedK8sVersions, eksDReleaseMap, r)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle release assets")
	}

	fmt.Printf("%s Successfully generated bundle artifacts table\n", constants.SuccessIcon)

	return artifactsTable, nil
}

func BundleArtifactsRelease(r *releasetypes.ReleaseConfig) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                  Bundle Artifacts Release")
	fmt.Println("==========================================================")
	err := DownloadArtifacts(r, r.BundleArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	err = RenameArtifacts(r, r.BundleArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	err = UploadArtifacts(r, r.BundleArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	return nil
}

func GenerateImageDigestsTable(r *releasetypes.ReleaseConfig) (map[string]string, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("                 Image Digests Table Generation")
	fmt.Println("==========================================================")
	imageDigests := make(map[string]string)

	for _, artifacts := range r.BundleArtifactsTable {
		for _, artifact := range artifacts {
			if artifact.Image != nil {
				var imageDigestStr string
				var err error
				if r.DryRun {
					sha256sum, err := artifactutils.GetFakeSHA(256)
					if err != nil {
						return nil, errors.Cause(err)
					}
					imageDigestStr = fmt.Sprintf("sha256:%s", sha256sum)
				} else {
					imageDigestStr, err = ecrpublic.GetImageDigest(artifact.Image.ReleaseImageURI, r.ReleaseContainerRegistry, r.ReleaseClients.ECRPublic.Client)
					if err != nil {
						return nil, errors.Cause(err)
					}
				}

				imageDigests[artifact.Image.ReleaseImageURI] = imageDigestStr
				fmt.Printf("Image digest for %s - %s\n", artifact.Image.ReleaseImageURI, imageDigestStr)
			}
		}
	}
	fmt.Printf("%s Successfully generated image digests table\n", constants.SuccessIcon)

	return imageDigests, nil
}

func GenerateBundleSpec(r *releasetypes.ReleaseConfig, bundle *anywherev1alpha1.Bundles, imageDigests map[string]string) error {
	fmt.Println("\n==========================================================")
	fmt.Println("               Bundles Manifest Spec Generation")
	fmt.Println("==========================================================")
	versionsBundles, err := bundles.GetVersionsBundles(r, imageDigests)
	if err != nil {
		return err
	}

	bundle.Spec.VersionsBundles = versionsBundles

	fmt.Printf("%s Successfully generated bundle manifest spec\n", constants.SuccessIcon)
	return nil
}
