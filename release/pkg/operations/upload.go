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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/pkg/aws/ecrpublic"
	"github.com/aws/eks-anywhere/release/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/pkg/bundles"
	"github.com/aws/eks-anywhere/release/pkg/constants"
	"github.com/aws/eks-anywhere/release/pkg/images"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
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
				if artifact.Archive.ProjectPath == bundles.HookProjectPath {
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
				sourceImageUri := artifact.Image.SourceImageURI
				releaseImageUri := artifact.Image.ReleaseImageURI
				fmt.Printf("Source Image - %s\n", sourceImageUri)
				fmt.Printf("Destination Image - %s\n", releaseImageUri)
				exists, err := ecrpublic.CheckImageExistence(releaseImageUri, r.ReleaseContainerRegistry, r.ReleaseClients.ECRPublic.Client)
				if err != nil {
					return fmt.Errorf("checking for image existence in ECR Public: %v", err)
				}
				if !exists {
					err := images.CopyToDestination(sourceEcrAuthConfig, releaseEcrAuthConfig, sourceImageUri, releaseImageUri)
					if err != nil {
						return fmt.Errorf("copying image from source to destination: %v", err)
					}
				}
			}
		}
	}
	fmt.Printf("%s Successsfully uploaded artifacts\n", constants.SuccessIcon)

	return nil
}
