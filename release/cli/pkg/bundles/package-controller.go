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
	"strings"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/ecr"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/ecrpublic"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/helm"
	"github.com/aws/eks-anywhere/release/cli/pkg/images"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/cli/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/cli/pkg/version"
)

func GetPackagesBundle(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.PackageBundle, error) {
	projectsInBundle := []string{"eks-anywhere-packages"}
	packagesArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, err := r.BundleArtifactsTable.Load(project)
		if err != nil {
			return anywherev1alpha1.PackageBundle{}, fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		packagesArtifacts[project] = projectArtifacts
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(packagesArtifacts)

	var sourceBranch string
	var componentChecksum string
	var Helmtag, Imagetag, Tokentag string
	var Helmsha, Imagesha, TokenSha string
	var err error
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	artifactHashes := []string{}

	// Find latest packages dev build for the Helm chart and Image
	// If we do find the tag in private ECR but it doesn't exist in public ECR, skopeo copy the image so that the helm chart works correctly
	if r.DevRelease && !r.DryRun {
		Helmtag, Helmsha, err = ecr.GetLatestImage(r.SourceClients.Packages.EcrClient, "eks-anywhere-packages", r.BuildRepoBranchName, true)
		if err != nil {
			fmt.Printf("Error getting dev version helm tag EKS Anywhere package controller, using latest version %v", err)
		}
		Imagetag, Imagesha, err = ecr.GetLatestImage(r.SourceClients.Packages.EcrClient, "eks-anywhere-packages", r.BuildRepoBranchName, false)
		if err != nil {
			fmt.Printf("Error getting dev version Image tag EKS Anywhere package controller, using latest version %v", err)
		}
		PackageImage, err := ecrpublic.CheckImageExistence(fmt.Sprintf("%s/%s:%s", r.ReleaseContainerRegistry, "eks-anywhere-packages", Imagetag), r.ReleaseContainerRegistry, r.ReleaseClients.ECRPublic.Client)
		if err != nil {
			fmt.Printf("Error checking image version existence for EKS Anywhere package controller, using latest version: %v", err)
		}
		if !PackageImage {
			fmt.Printf("Required helm image not found in public ECR... copying image: %v\n", fmt.Sprintf("%s/%s:%s", r.ReleaseContainerRegistry, "eks-anywhere-packages", Imagetag))
			err := images.CopyToDestination(r.SourceClients.Packages.AuthConfig, r.ReleaseClients.ECRPublic.AuthConfig, fmt.Sprintf("%s/%s:%s", r.PackagesSourceContainerRegistry, "eks-anywhere-packages", Imagetag), fmt.Sprintf("%s/%s:%s", r.ReleaseContainerRegistry, "eks-anywhere-packages", Imagetag))
			if err != nil {
				fmt.Printf("Error copying dev EKS Anywhere package controller image, to ECR Public: %v", err)
			}
		}
		Tokentag, TokenSha, err = ecr.GetLatestImage(r.SourceClients.Packages.EcrClient, "ecr-token-refresher", r.BuildRepoBranchName, false)
		if err != nil {
			fmt.Printf("Error getting dev version Image tag EKS Anywhere package token refresher, using latest version %v", err)
		}
		TokenImage, err := ecrpublic.CheckImageExistence(fmt.Sprintf("%s/%s:%s", r.ReleaseContainerRegistry, "ecr-token-refresher", Tokentag), r.ReleaseContainerRegistry, r.ReleaseClients.ECRPublic.Client)
		if err != nil {
			fmt.Printf("Error checking image version existence for EKS Anywhere package ecr token refresher, using latest version: %v", err)
		}
		if !TokenImage {
			fmt.Printf("Required helm image not found in public ECR... copying image: %v\n", fmt.Sprintf("%s/%s:%s", r.ReleaseContainerRegistry, "ecr-token-refresher", Tokentag))
			err := images.CopyToDestination(r.SourceClients.Packages.AuthConfig, r.ReleaseClients.ECRPublic.AuthConfig, fmt.Sprintf("%s/%s:%s", r.PackagesSourceContainerRegistry, "ecr-token-refresher", Tokentag), fmt.Sprintf("%s/%s:%s", r.ReleaseContainerRegistry, "ecr-token-refresher", Tokentag))
			if err != nil {
				fmt.Printf("Error copying dev EKS Anywhere package ecr token refresher image, to ECR Public: %v", err)
			}
		}
	}
	for _, componentName := range sortedComponentNames {
		for _, artifact := range packagesArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				sourceBranch = imageArtifact.SourcedFromBranch
				bundleImageArtifact := anywherev1alpha1.Image{}
				if strings.HasSuffix(imageArtifact.AssetName, "helm") {
					imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
					if err != nil {
						return anywherev1alpha1.PackageBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
					}
					if r.DevRelease && Helmsha != "" && Helmtag != "" {
						imageDigest = Helmsha
						imageArtifact.SourceImageURI = replaceTag(imageArtifact.SourceImageURI, Helmtag)
					}
					assetName := strings.TrimSuffix(imageArtifact.AssetName, "-helm")
					bundleImageArtifact = anywherev1alpha1.Image{
						Name:        assetName,
						Description: fmt.Sprintf("Helm chart for %s", assetName),
						URI:         imageArtifact.ReleaseImageURI,
						ImageDigest: imageDigest,
					}
				} else {
					imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
					if err != nil {
						return anywherev1alpha1.PackageBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
					}
					if strings.HasSuffix(imageArtifact.AssetName, "eks-anywhere-packages") && r.DevRelease && Imagesha != "" && Imagetag != "" {
						imageDigest = Imagesha
						imageArtifact.SourceImageURI = replaceTag(imageArtifact.SourceImageURI, Imagetag)
					} else if strings.HasSuffix(imageArtifact.AssetName, "ecr-token-refresher") && r.DevRelease && TokenSha != "" && Tokentag != "" {
						imageDigest = TokenSha
						imageArtifact.SourceImageURI = replaceTag(imageArtifact.SourceImageURI, Tokentag)
					}
					bundleImageArtifact = anywherev1alpha1.Image{
						Name:        imageArtifact.AssetName,
						Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
						OS:          imageArtifact.OS,
						Arch:        imageArtifact.Arch,
						URI:         imageArtifact.ReleaseImageURI,
						ImageDigest: imageDigest,
					}
				}
				bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
				artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
			}
		}
	}

	if !r.DryRun && r.DevRelease && r.BuildRepoBranchName == "main" {
		for _, componentName := range sortedComponentNames {
			for _, artifact := range packagesArtifacts[componentName] {
				if artifact.Image != nil {
					imageArtifact := artifact.Image
					sourceBranch = imageArtifact.SourcedFromBranch
					if strings.HasSuffix(imageArtifact.AssetName, "helm") {
						trimmedAsset := strings.TrimSuffix(artifact.Image.AssetName, "-helm")
						fmt.Printf("trimmedAsset=%v\n\n", trimmedAsset)
						helmDriver, err := helm.NewHelm()
						if err != nil {
							return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "creating helm client")
						}
						fmt.Printf("Modifying helm chart for %s\n", trimmedAsset)
						helmDest, err := helm.GetHelmDest(helmDriver, r, imageArtifact.SourceImageURI, trimmedAsset)
						if err != nil {
							return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "getting Helm destination:")
						}
						fmt.Printf("helmDest=%v\n", helmDest)
						fmt.Printf("Pulled helm chart locally to %s\n", helmDest)
						err = helm.ModifyAndPushChartYaml(*imageArtifact, r, helmDriver, helmDest, packagesArtifacts, bundleImageArtifacts)
						if err != nil {
							return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "modifying Chart.yaml and pushing Helm chart to destination:")
						}
					}
				}
			}
		}
	}

	if r.DryRun {
		componentChecksum = version.FakeComponentChecksum
	} else {
		componentChecksum = version.GenerateComponentHash(artifactHashes, r.DryRun)
	}
	version, err := version.BuildComponentVersion(
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.PackagesProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.PackageBundle{}, errors.Wrap(err, "Error getting version for EKS Anywhere package controller")
	}

	bundle := anywherev1alpha1.PackageBundle{
		Version:                   version,
		Controller:                bundleImageArtifacts["eks-anywhere-packages"],
		TokenRefresher:            bundleImageArtifacts["ecr-token-refresher"],
		CredentialProviderPackage: bundleImageArtifacts["credential-provider-package"],
		HelmChart:                 bundleImageArtifacts["eks-anywhere-packages-helm"],
	}
	return bundle, nil
}

// replaceTag is used to replace the tag of an Image URI with a string.
func replaceTag(uri, tag string) string {
	NewURIList := strings.Split(uri, ":")
	if len(NewURIList) < 2 {
		return uri
	}
	NewURIList[len(NewURIList)-1] = tag
	uri = strings.Join(NewURIList[:], ":")
	return uri
}
