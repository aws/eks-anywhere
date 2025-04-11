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
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	anywhereconstants "github.com/aws/eks-anywhere/pkg/constants"
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/assets"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/ecrpublic"
	"github.com/aws/eks-anywhere/release/cli/pkg/bundles"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	"github.com/aws/eks-anywhere/release/cli/pkg/images"
	sig "github.com/aws/eks-anywhere/release/cli/pkg/signature"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
	commandutils "github.com/aws/eks-anywhere/release/cli/pkg/util/command"
)

func GenerateBundleArtifactsTable(r *releasetypes.ReleaseConfig) (releasetypes.ArtifactsTable, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("              Bundle Artifacts Table Generation")
	fmt.Println("==========================================================")

	eksDReleaseMap, err := filereader.ReadEksDReleases(r)
	if err != nil {
		return releasetypes.ArtifactsTable{}, err
	}

	supportedK8sVersions, err := filereader.GetSupportedK8sVersions(r)
	if err != nil {
		return releasetypes.ArtifactsTable{}, errors.Wrapf(err, "Error getting supported Kubernetes versions for bottlerocket")
	}

	artifactsTable, err := assets.GetBundleReleaseAssets(supportedK8sVersions, eksDReleaseMap, r)
	if err != nil {
		return releasetypes.ArtifactsTable{}, errors.Wrapf(err, "Error getting bundle release assets")
	}

	fmt.Printf("%s Successfully generated bundle artifacts table\n", constants.SuccessIcon)

	return artifactsTable, nil
}

func BundleArtifactsRelease(r *releasetypes.ReleaseConfig) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                  Bundle Artifacts Release")
	fmt.Println("==========================================================")
	err := DownloadArtifacts(context.Background(), r, r.BundleArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	err = RenameArtifacts(context.Background(), r, r.BundleArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	err = UploadArtifacts(context.Background(), r, r.BundleArtifactsTable, true)
	if err != nil {
		return errors.Cause(err)
	}

	return nil
}

func GenerateImageDigestsTable(ctx context.Context, r *releasetypes.ReleaseConfig) (releasetypes.ImageDigestsTable, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("                 Image Digests Table Generation")
	fmt.Println("==========================================================")
	var imageDigests releasetypes.ImageDigestsTable

	errGroup, ctx := errgroup.WithContext(ctx)
	r.BundleArtifactsTable.Range(func(k, v interface{}) bool {
		artifacts := v.([]releasetypes.Artifact)
		for _, artifact := range artifacts {
			r, artifact := r, artifact
			errGroup.Go(func() error {
				if artifact.Image != nil {
					imageDigest, err := getImageDigest(ctx, r, artifact)
					if err != nil {
						return errors.Wrapf(err, "getting image digest for image %s", artifact.Image.ReleaseImageURI)
					}
					imageDigests.Store(artifact.Image.ReleaseImageURI, imageDigest)
					fmt.Printf("Image digest for %s - %s\n", artifact.Image.ReleaseImageURI, imageDigest)
				}

				return nil
			})
		}
		return true
	})
	if err := errGroup.Wait(); err != nil {
		return releasetypes.ImageDigestsTable{}, fmt.Errorf("generating image digests table: %v", err)
	}
	fmt.Printf("%s Successfully generated image digests table\n", constants.SuccessIcon)

	return imageDigests, nil
}

func SignImagesNotation(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) error {
	if r.DryRun {
		fmt.Println("Skipping image signing in dry-run mode")
		return nil
	}
	releaseRegistryUsername := r.ReleaseClients.ECRPublic.AuthConfig.Username
	releaseRegistryPassword := r.ReleaseClients.ECRPublic.AuthConfig.Password
	var rangeErr error
	imageDigests.Range(func(k, v interface{}) bool {
		image := k.(string)
		digest := v.(string)
		imageURI := fmt.Sprintf("%s@%s", image, digest)
		cmd := exec.Command("notation", "list", imageURI, "-u", releaseRegistryUsername, "-p", releaseRegistryPassword)
		out, err := commandutils.ExecCommand(cmd)
		if err != nil {
			rangeErr = fmt.Errorf("listing signatures associated with image %s: %v", imageURI, err)
			return false
		}
		// Skip signing image if it is already signed.
		if strings.Contains(out, "no associated signature") {
			// Sign public ECR image using AWS signer and notation CLI
			// notation sign <registry>/<repository>@<sha256:shasum> --plugin com.amazonaws.signer.notation.plugin --id <signer_profile_arn>
			cmd := exec.Command("notation", "sign", imageURI, "--plugin", "com.amazonaws.signer.notation.plugin", "--id", r.AwsSignerProfileArn, "-u", releaseRegistryUsername, "-p", releaseRegistryPassword)
			out, err := commandutils.ExecCommand(cmd)
			fmt.Println(out)
			if err != nil {
				rangeErr = fmt.Errorf("signing container image with Notation CLI: %v", err)
				return false
			}
		} else {
			rangeErr = nil
			fmt.Printf("Skipping image signing for image %s since it has already been signed\n", imageURI)
		}
		return true
	})

	return rangeErr
}

// Copy image signatures to production account from staging account
func CopyImageSignatureUsingOras(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) error {
	if r.DryRun {
		fmt.Println("Skipping image signature copy in dry-run mode")
		return nil
	}
	sourceRegistryClient := r.SourceClients.ECR.EcrPublicClient
	sourceRegistryUsername := r.SourceClients.ECR.AuthConfig.Username
	sourceRegistryPassword := r.SourceClients.ECR.AuthConfig.Password
	releaseRegistryClient := r.ReleaseClients.ECRPublic.Client
	releaseRegistryUsername := r.ReleaseClients.ECRPublic.AuthConfig.Username
	releaseRegistryPassword := r.ReleaseClients.ECRPublic.AuthConfig.Password
	var rangeErr error
	imageDigests.Range(func(k, v interface{}) bool {
		image := k.(string)
		digest := v.(string)

		// Get imageRespository name since we have a different source and release registry.
		imageRepository, imageTag := artifactutils.SplitImageUri(image, r.ReleaseContainerRegistry)
		// Compute image digest for source and destination images from manifest contents. The digest will be in the
		// form sha256:digest, so we are changing it to Notation's image index and signatures format, sha256-digest.
		sourceImageDigest, err := images.ComputeImageDigestFromManifest(sourceRegistryClient, r.SourceContainerRegistry, imageRepository, imageTag)
		if err != nil {
			rangeErr = fmt.Errorf("computing digest for source image %s from manifest: %v\n", image, err)
			return false
		}
		sourceImageSignatureDigest := fmt.Sprintf("sha256-%s", sourceImageDigest)
		releaseImageDigest, err := images.ComputeImageDigestFromManifest(releaseRegistryClient, r.ReleaseContainerRegistry, imageRepository, imageTag)
		if err != nil {
			rangeErr = fmt.Errorf("computing digest for destination image %s from manifest: %v\n", image, err)
			return false
		}
		releaseImageSignatureDigest := fmt.Sprintf("sha256-%s", releaseImageDigest)
		// Form releaseImageURI in the form <source-registry>/<repository>:<sha256-digest>
		sourceImageURI := fmt.Sprintf("%s/%s:%s", r.SourceContainerRegistry, imageRepository, sourceImageSignatureDigest)
		// Form releaseImageURI in the form <release-registry>/<repository>:<sha256-digest>
		releaseImageURI := fmt.Sprintf("%s/%s:%s", r.ReleaseContainerRegistry, imageRepository, releaseImageSignatureDigest)

		cmd := exec.Command("oras", "copy", "--from-username", sourceRegistryUsername, "--from-password", sourceRegistryPassword, sourceImageURI, "--to-username", releaseRegistryUsername, "--to-password", releaseRegistryPassword, releaseImageURI)
		out, err := commandutils.ExecCommand(cmd)
		fmt.Println(out)
		if err != nil {
			rangeErr = fmt.Errorf("copying signatures associated with image %s: %v\n", fmt.Sprintf("%s@%s", image, digest), err)
			return false
		}
		return true
	})

	return rangeErr
}

func GenerateBundleSpec(r *releasetypes.ReleaseConfig, bundle *anywherev1alpha1.Bundles, imageDigests releasetypes.ImageDigestsTable) error {
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

// SignBundleManifest is the top-level function that computes the Bundles
// manifest signature using AWS KMS and attaches that signature as an
// annotation on the Bundles object.
func SignBundleManifest(ctx context.Context, bundle *anywherev1alpha1.Bundles) error {
	fmt.Println("\n==========================================================")
	fmt.Println("               Bundles Manifest Signing")
	fmt.Println("==========================================================")

	if bundle.Annotations == nil {
		bundle.Annotations = make(map[string]string, 1)
	}
	bundle.Annotations[anywhereconstants.ExcludesAnnotation] = anywhereconstants.Excludes

	fmt.Printf("Generating bundle signature with key: %s\n", constants.KmsKey)

	signature, err := sig.GetBundleSignature(ctx, bundle, constants.KmsKey)
	if err != nil {
		return err
	}
	bundle.Annotations[anywhereconstants.SignatureAnnotation] = signature

	fmt.Printf("%s Successfully signed bundle manifest\n", constants.SuccessIcon)
	return nil
}

func getImageDigest(_ context.Context, r *releasetypes.ReleaseConfig, artifact releasetypes.Artifact) (string, error) {
	var imageDigest string
	var err error
	if r.DryRun {
		sha256sum, err := artifactutils.GetFakeSHA(256)
		if err != nil {
			return "", errors.Cause(err)
		}
		imageDigest = fmt.Sprintf("sha256:%s", sha256sum)
	} else {
		releaseImageUri := artifact.Image.ReleaseImageURI
		releaseContainerRegistry := r.ReleaseContainerRegistry
		ecrPublicClient := r.ReleaseClients.ECRPublic.Client
		if r.DevRelease  && (strings.Contains(releaseImageUri, "eks-anywhere-packages") || strings.Contains(releaseImageUri, "ecr-token-refresher") || strings.Contains(releaseImageUri, "credential-provider-package")) {
			ecrPublicClient = r.ReleaseClients.Packages.Client
			releaseContainerRegistry = r.PackagesReleaseContainerRegistry
		}
		imageDigest, err = ecrpublic.GetImageDigest(releaseImageUri, releaseContainerRegistry, ecrPublicClient)
		if err != nil {
			return "", errors.Cause(err)
		}
	}

	return imageDigest, nil
}
