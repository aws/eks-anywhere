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
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/cli/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/helm"
	"github.com/aws/eks-anywhere/release/cli/pkg/images"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func UploadArtifacts(r *releasetypes.ReleaseConfig, eksArtifacts map[string][]releasetypes.Artifact) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                  Artifacts Upload")
	fmt.Println("==========================================================")
	if r.DryRun {
		fmt.Println("Skipping artifacts upload in dry-run mode")
		return nil
	}

	sourceEcrAuthConfig := r.SourceClients.ECR.AuthConfig
	releaseEcrAuthConfig := r.ReleaseClients.ECRPublic.AuthConfig

	for _, artifacts := range eksArtifacts {
		for _, artifact := range artifacts {
			if artifact.Archive != nil {
				archiveFile := filepath.Join(artifact.Archive.ArtifactPath, artifact.Archive.ReleaseName)
				fmt.Printf("Archive - %s\n", archiveFile)
				key := filepath.Join(artifact.Archive.ReleaseS3Path, artifact.Archive.ReleaseName)
				err := s3.UploadFile(archiveFile, aws.String(r.ReleaseBucket), aws.String(key), r.ReleaseClients.S3.Uploader)
				if err != nil {
					return errors.Cause(err)
				}

				checksumExtensions := []string{".sha256", ".sha512"}
				// Adding a special case for tinkerbell/hook project.
				// The project builds linux kernel files that are not stored as tarballs and currently do not have SHA checksums.
				// TODO(pokearu): Add logic to generate SHA for hook project
				if artifact.Archive.ProjectPath == constants.HookProjectPath {
					checksumExtensions = []string{}
				}

				for _, extension := range checksumExtensions {
					checksumFile := filepath.Join(artifact.Archive.ArtifactPath, artifact.Archive.ReleaseName) + extension
					fmt.Printf("Checksum - %s\n", checksumFile)
					key := filepath.Join(artifact.Archive.ReleaseS3Path, artifact.Archive.ReleaseName) + extension
					err := s3.UploadFile(checksumFile, aws.String(r.ReleaseBucket), aws.String(key), r.ReleaseClients.S3.Uploader)
					if err != nil {
						return errors.Cause(err)
					}
				}
			}

			if artifact.Manifest != nil {
				manifestFile := filepath.Join(artifact.Manifest.ArtifactPath, artifact.Manifest.ReleaseName)
				fmt.Printf("Manifest - %s\n", manifestFile)
				key := filepath.Join(artifact.Manifest.ReleaseS3Path, artifact.Manifest.ReleaseName)
				err := s3.UploadFile(manifestFile, aws.String(r.ReleaseBucket), aws.String(key), r.ReleaseClients.S3.Uploader)
				if err != nil {
					return errors.Cause(err)
				}
			}

			if artifact.Image != nil {
				// If the artifact is a helm chart, skip the skopeo copy. Instead, modify the Chart.yaml to match the release tag
				// and then use Helm package and push commands to upload chart to ECR Public
				if !r.DryRun && ((strings.HasSuffix(artifact.Image.AssetName, "helm") && !r.DevRelease) || strings.HasSuffix(artifact.Image.AssetName, "chart")) {
					// Trim -helm on the packages helm chart, but don't need to trim tinkerbell chart since the AssetName is the same as the repoName
					trimmedAsset := strings.TrimSuffix(artifact.Image.AssetName, "-helm")

					helmDriver, err := helm.NewHelm()
					if err != nil {
						return fmt.Errorf("creating helm client: %v", err)
					}

					fmt.Printf("Modifying helm chart for %s\n", trimmedAsset)
					helmDest, err := helm.GetHelmDest(helmDriver, r, artifact.Image.SourceImageURI, trimmedAsset)
					if err != nil {
						return fmt.Errorf("getting Helm destination: %v", err)
					}

					fmt.Printf("Pulled helm chart locally to %s\n", helmDest)
					err = helm.ModifyAndPushChartYaml(*artifact.Image, r, helmDriver, helmDest, eksArtifacts, nil)
					if err != nil {
						return fmt.Errorf("modifying Chart.yaml and pushing Helm chart to destination: %v", err)
					}

					continue
				}
				sourceImageUri := artifact.Image.SourceImageURI
				releaseImageUri := artifact.Image.ReleaseImageURI
				fmt.Printf("Source Image - %s\n", sourceImageUri)
				fmt.Printf("Destination Image - %s\n", releaseImageUri)
				err := images.CopyToDestination(sourceEcrAuthConfig, releaseEcrAuthConfig, sourceImageUri, releaseImageUri)
				if err != nil {
					return fmt.Errorf("copying image from source to destination: %v", err)
				}
			}
		}
	}
	fmt.Printf("%s Successfully uploaded artifacts\n", constants.SuccessIcon)

	return nil
}
