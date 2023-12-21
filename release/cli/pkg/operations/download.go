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
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/aws/eks-anywhere/release/cli/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	"github.com/aws/eks-anywhere/release/cli/pkg/retrier"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
)

func DownloadArtifacts(ctx context.Context, r *releasetypes.ReleaseConfig, eksaArtifacts releasetypes.ArtifactsTable) error {
	// Retrier for downloading source S3 objects. This retrier has a max timeout of 60 minutes. It
	// checks whether the error occured during download is an ObjectNotFound error and retries the
	// download operation for a maximum of 60 retries, with a wait time of 30 seconds per retry.
	s3Retrier := retrier.NewRetrier(60*time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if r.BuildRepoBranchName == "main" && artifactutils.IsObjectNotFoundError(err) && totalRetries < 60 {
			return true, 30 * time.Second
		}
		return false, 0
	}))
	fmt.Println("==========================================================")
	fmt.Println("                  Artifacts Download")
	fmt.Println("==========================================================")

	errGroup, ctx := errgroup.WithContext(ctx)

	eksaArtifacts.Range(func(k, v interface{}) bool {
		artifacts := v.([]releasetypes.Artifact)
		for _, artifact := range artifacts {
			r, artifact, s3Retrier := r, artifact, s3Retrier
			errGroup.Go(func() error {
				// Check if there is an archive to be downloaded
				if artifact.Archive != nil {
					return handleArchiveDownload(ctx, r, artifact, s3Retrier)
				}

				// Check if there is a manifest to be downloaded
				if artifact.Manifest != nil {
					return handleManifestDownload(ctx, r, artifact, s3Retrier)
				}

				return nil
			})
		}
		return true
	})
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("downloading artifacts: %v", err)
	}
	fmt.Printf("%s Successfully downloaded artifacts\n", constants.SuccessIcon)

	return nil
}

func handleArchiveDownload(_ context.Context, r *releasetypes.ReleaseConfig, artifact releasetypes.Artifact, s3Retrier *retrier.Retrier) error {
	sourceS3Prefix := artifact.Archive.SourceS3Prefix
	sourceS3Key := artifact.Archive.SourceS3Key
	artifactPath := artifact.Archive.ArtifactPath
	objectKey := filepath.Join(sourceS3Prefix, sourceS3Key)
	objectLocalFilePath := filepath.Join(artifactPath, sourceS3Key)
	fmt.Printf("Archive - %s\n", objectKey)
	if r.DryRun && artifact.Archive.ImageFormat != "tarball" {
		fmt.Println("Skipping OS image downloads in dry-run mode")
	} else {
		err := s3Retrier.Retry(func() error {
			if !s3.KeyExists(r.SourceBucket, objectKey) {
				return fmt.Errorf("requested object not found")
			}
			return nil
		})
		if err != nil {
			if r.BuildRepoBranchName != "main" {
				var latestSourceS3PrefixFromMain string
				fmt.Printf("Artifact corresponding to %s branch not found for %s archive. Using artifact from main\n", r.BuildRepoBranchName, sourceS3Key)
				if strings.Contains(sourceS3Key, "eksctl-anywhere") {
					latestSourceS3PrefixFromMain = strings.NewReplacer(r.CliRepoBranchName, "latest").Replace(sourceS3Prefix)
				} else {
					gitTagFromMain, err := filereader.ReadGitTag(artifact.Archive.ProjectPath, r.BuildRepoSource, "main")
					if err != nil {
						return errors.Cause(err)
					}
					latestSourceS3PrefixFromMain = strings.NewReplacer(r.BuildRepoBranchName, "latest", artifact.Archive.GitTag, gitTagFromMain).Replace(sourceS3Prefix)
				}
				objectKey = filepath.Join(latestSourceS3PrefixFromMain, sourceS3Key)
			} else {
				return fmt.Errorf("retries exhausted waiting for archive to be uploaded to source location: %v", err)
			}
		}

		err = s3.DownloadFile(objectLocalFilePath, r.SourceBucket, objectKey)
		if err != nil {
			return errors.Cause(err)
		}

		// Download checksum files for the archive
		checksumExtensions := []string{
			".sha256",
			".sha512",
		}

		// Adding a special case for tinkerbell/hook project.
		// The project builds linux kernel files that are not stored as tarballs and currently do not have SHA checksums.
		// TODO(pokearu): Add logic to generate SHA for hook project
		if artifact.Archive.ProjectPath == constants.HookProjectPath {
			checksumExtensions = []string{}
		}

		for _, extension := range checksumExtensions {
			objectShasumFileName := fmt.Sprintf("%s%s", sourceS3Key, extension)
			objectShasumFileKey := filepath.Join(sourceS3Prefix, objectShasumFileName)
			objectShasumFileLocalFilePath := filepath.Join(artifactPath, objectShasumFileName)
			fmt.Printf("Checksum file - %s\n", objectShasumFileKey)

			err := s3Retrier.Retry(func() error {
				if !s3.KeyExists(r.SourceBucket, objectShasumFileKey) {
					return fmt.Errorf("requested object not found")
				}
				return nil
			})
			if err != nil {
				if r.BuildRepoBranchName != "main" {
					var latestSourceS3PrefixFromMain string
					fmt.Printf("Artifact corresponding to %s branch not found for %s archive. Using artifact from main\n", r.BuildRepoBranchName, sourceS3Key)
					if strings.Contains(sourceS3Key, "eksctl-anywhere") {
						latestSourceS3PrefixFromMain = strings.NewReplacer(r.CliRepoBranchName, "latest").Replace(sourceS3Prefix)
					} else {
						gitTagFromMain, err := filereader.ReadGitTag(artifact.Archive.ProjectPath, r.BuildRepoSource, "main")
						if err != nil {
							return errors.Cause(err)
						}
						latestSourceS3PrefixFromMain = strings.NewReplacer(r.BuildRepoBranchName, "latest", artifact.Archive.GitTag, gitTagFromMain).Replace(sourceS3Prefix)
					}
					objectShasumFileKey = filepath.Join(latestSourceS3PrefixFromMain, objectShasumFileName)
				} else {
					return fmt.Errorf("retries exhausted waiting for checksum file to be uploaded to source location: %v", err)
				}
			}

			err = s3.DownloadFile(objectShasumFileLocalFilePath, r.SourceBucket, objectShasumFileKey)
			if err != nil {
				return errors.Cause(err)
			}
		}
	}

	return nil
}

func handleManifestDownload(_ context.Context, r *releasetypes.ReleaseConfig, artifact releasetypes.Artifact, s3Retrier *retrier.Retrier) error {
	sourceS3Prefix := artifact.Manifest.SourceS3Prefix
	sourceS3Key := artifact.Manifest.SourceS3Key
	artifactPath := artifact.Manifest.ArtifactPath
	objectKey := filepath.Join(sourceS3Prefix, sourceS3Key)
	objectLocalFilePath := filepath.Join(artifactPath, sourceS3Key)
	fmt.Printf("Manifest - %s\n", objectKey)

	err := s3Retrier.Retry(func() error {
		if !s3.KeyExists(r.SourceBucket, objectKey) {
			return fmt.Errorf("Requested object not found")
		}
		return nil
	})
	if err != nil {
		if r.BuildRepoBranchName != "main" {
			fmt.Printf("Artifact corresponding to %s branch not found for %s manifest. Using artifact from main\n", r.BuildRepoBranchName, sourceS3Key)
			gitTagFromMain, err := filereader.ReadGitTag(artifact.Manifest.ProjectPath, r.BuildRepoSource, "main")
			if err != nil {
				return errors.Cause(err)
			}
			latestSourceS3PrefixFromMain := strings.NewReplacer(r.BuildRepoBranchName, "latest", artifact.Manifest.GitTag, gitTagFromMain).Replace(sourceS3Prefix)
			objectKey = filepath.Join(latestSourceS3PrefixFromMain, sourceS3Key)
		} else {
			return fmt.Errorf("retries exhausted waiting for archive to be uploaded to source location: %v", err)
		}
	}

	err = s3.DownloadFile(objectLocalFilePath, r.SourceBucket, objectKey)
	if err != nil {
		return errors.Cause(err)
	}

	return nil
}
