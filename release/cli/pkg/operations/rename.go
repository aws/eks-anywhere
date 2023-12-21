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
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func RenameArtifacts(ctx context.Context, r *releasetypes.ReleaseConfig, eksaArtifacts releasetypes.ArtifactsTable) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                    Artifacts Rename")
	fmt.Println("==========================================================")

	errGroup, ctx := errgroup.WithContext(ctx)

	eksaArtifacts.Range(func(k, v interface{}) bool {
		artifacts := v.([]releasetypes.Artifact)
		for _, artifact := range artifacts {
			r, artifact := r, artifact
			errGroup.Go(func() error {
				// Change the name of the archive along with the checksum files
				if artifact.Archive != nil {
					return handleArchiveRename(ctx, r, artifact)
				}

				// Override images in the manifest with release URIs
				if artifact.Manifest != nil {
					return handleManifestRename(ctx, r, artifact)
				}

				return nil
			})
		}
		return true
	})
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("renaming artifacts: %v", err)
	}
	fmt.Printf("%s Successfully renamed artifacts\n", constants.SuccessIcon)

	return nil
}

func handleArchiveRename(_ context.Context, r *releasetypes.ReleaseConfig, artifact releasetypes.Artifact) error {
	if r.DryRun && artifact.Archive.ImageFormat != "tarball" {
		fmt.Println("Skipping OS image renames in dry-run mode")
	} else {
		archiveArtifact := artifact.Archive
		oldArtifactFile := filepath.Join(archiveArtifact.ArtifactPath, archiveArtifact.SourceS3Key)
		newArtifactFile := filepath.Join(archiveArtifact.ArtifactPath, archiveArtifact.ReleaseName)
		fmt.Printf("Renaming archive - %s\n", newArtifactFile)
		err := os.Rename(oldArtifactFile, newArtifactFile)
		if err != nil {
			return errors.Cause(err)
		}

		// Change the names of the checksum files
		checksumExtensions := []string{".sha256", ".sha512"}

		// Adding a special case for tinkerbell/hook project.
		// The project builds linux kernel files that are not stored as tarballs and currently do not have SHA checksums.
		// TODO(pokearu): Add logic to generate SHA for hook project
		if artifact.Archive.ProjectPath == constants.HookProjectPath {
			checksumExtensions = []string{}
		}

		for _, extension := range checksumExtensions {
			oldChecksumFile := oldArtifactFile + extension
			newChecksumFile := newArtifactFile + extension
			fmt.Printf("Renaming checksum file - %s\n", newChecksumFile)
			err = os.Rename(oldChecksumFile, newChecksumFile)
			if err != nil {
				return errors.Cause(err)
			}
		}
	}

	return nil
}

func handleManifestRename(_ context.Context, r *releasetypes.ReleaseConfig, artifact releasetypes.Artifact) error {
	manifestArtifact := artifact.Manifest
	oldArtifactFile := filepath.Join(manifestArtifact.ArtifactPath, manifestArtifact.SourceS3Key)
	newArtifactFile := filepath.Join(manifestArtifact.ArtifactPath, manifestArtifact.ReleaseName)
	fmt.Printf("Renaming manifest - %s\n", newArtifactFile)
	err := os.Rename(oldArtifactFile, newArtifactFile)
	if err != nil {
		return errors.Cause(err)
	}

	for _, imageTagOverride := range manifestArtifact.ImageTagOverrides {
		manifestFileContents, err := os.ReadFile(newArtifactFile)
		if err != nil {
			return errors.Cause(err)
		}
		regex := fmt.Sprintf("%s/.*%s.*", r.SourceContainerRegistry, imageTagOverride.Repository)
		compiledRegex := regexp.MustCompile(regex)
		fmt.Printf("Overriding image to %s in manifest %s\n", imageTagOverride.ReleaseUri, newArtifactFile)
		updatedManifestFileContents := compiledRegex.ReplaceAllString(string(manifestFileContents), imageTagOverride.ReleaseUri)
		err = os.WriteFile(newArtifactFile, []byte(updatedManifestFileContents), 0o644)
		if err != nil {
			return errors.Cause(err)
		}
	}

	return nil
}
