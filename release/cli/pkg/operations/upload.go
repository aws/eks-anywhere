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
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/aws/eks-anywhere/release/cli/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/helm"
	"github.com/aws/eks-anywhere/release/cli/pkg/images"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func UploadArtifacts(ctx context.Context, r *releasetypes.ReleaseConfig, eksaArtifacts releasetypes.ArtifactsTable) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                  Artifacts Upload")
	fmt.Println("==========================================================")
	if r.DryRun {
		fmt.Println("Skipping artifacts upload in dry-run mode")
		return nil
	}

	errGroup, ctx := errgroup.WithContext(ctx)

	sourceEcrAuthConfig := r.SourceClients.ECR.AuthConfig
	releaseEcrAuthConfig := r.ReleaseClients.ECRPublic.AuthConfig

	projectsInBundle := []string{"eks-anywhere-packages"}
	packagesArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, err := r.BundleArtifactsTable.Load(project)
		if err != nil {
			return fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		packagesArtifacts[project] = projectArtifacts
	}

	eksaArtifacts.Range(func(k, v interface{}) bool {
		artifacts := v.([]releasetypes.Artifact)
		for _, artifact := range artifacts {
			r, packagesArtifacts, artifact, sourceEcrAuthConfig, releaseEcrAuthConfig := r, packagesArtifacts, artifact, sourceEcrAuthConfig, releaseEcrAuthConfig
			errGroup.Go(func() error {
				if artifact.Archive != nil {
					return handleArchiveUpload(ctx, r, artifact)
				}

				if artifact.Manifest != nil {
					return handleManifestUpload(ctx, r, artifact)
				}

				if artifact.Image != nil {
					return handleImageUpload(ctx, r, packagesArtifacts, artifact, sourceEcrAuthConfig, releaseEcrAuthConfig)
				}

				return nil
			})
		}
		return true
	})
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("uploading artifacts: %v", err)
	}
	fmt.Printf("%s Successfully uploaded artifacts\n", constants.SuccessIcon)

	return nil
}

func handleArchiveUpload(_ context.Context, r *releasetypes.ReleaseConfig, artifact releasetypes.Artifact) error {
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

	return nil
}

func handleManifestUpload(_ context.Context, r *releasetypes.ReleaseConfig, artifact releasetypes.Artifact) error {
	manifestFile := filepath.Join(artifact.Manifest.ArtifactPath, artifact.Manifest.ReleaseName)
	fmt.Printf("Manifest - %s\n", manifestFile)
	key := filepath.Join(artifact.Manifest.ReleaseS3Path, artifact.Manifest.ReleaseName)
	err := s3.UploadFile(manifestFile, aws.String(r.ReleaseBucket), aws.String(key), r.ReleaseClients.S3.Uploader)
	if err != nil {
		return errors.Cause(err)
	}

	return nil
}

func handleImageUpload(_ context.Context, r *releasetypes.ReleaseConfig, packagesArtifacts map[string][]releasetypes.Artifact, artifact releasetypes.Artifact, sourceEcrAuthConfig, releaseEcrAuthConfig *docker.AuthConfiguration) error {
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
		err = helm.ModifyAndPushChartYaml(*artifact.Image, r, helmDriver, helmDest, packagesArtifacts, nil)
		if err != nil {
			return fmt.Errorf("modifying Chart.yaml and pushing Helm chart to destination: %v", err)
		}
	} else {
		sourceImageUri := artifact.Image.SourceImageURI
		releaseImageUri := artifact.Image.ReleaseImageURI
		fmt.Printf("Source Image - %s\n", sourceImageUri)
		fmt.Printf("Destination Image - %s\n", releaseImageUri)
		err := images.CopyToDestination(sourceEcrAuthConfig, releaseEcrAuthConfig, sourceImageUri, releaseImageUri)
		if err != nil {
			return fmt.Errorf("copying image from source to destination: %v", err)
		}
	}

	return nil
}
