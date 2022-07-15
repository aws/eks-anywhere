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

package assets

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/pkg/assets/archives"
	assetconfig "github.com/aws/eks-anywhere/release/pkg/assets/config"
	"github.com/aws/eks-anywhere/release/pkg/assets/images"
	"github.com/aws/eks-anywhere/release/pkg/assets/manifests"
	"github.com/aws/eks-anywhere/release/pkg/assets/tagger"
	assettypes "github.com/aws/eks-anywhere/release/pkg/assets/types"
	"github.com/aws/eks-anywhere/release/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	sliceutils "github.com/aws/eks-anywhere/release/pkg/util/slices"
)

func getAssetsFromConfig(ac *assettypes.AssetConfig, rc *releasetypes.ReleaseConfig, eksDReleaseChannel, eksDReleaseNumber, kubeVersion string) ([]releasetypes.Artifact, error) {
	var artifacts []releasetypes.Artifact
	var imageTagOverrides []releasetypes.ImageTagOverride
	projectName := ac.ProjectName
	projectPath := ac.ProjectPath
	sourcedFromBranch := rc.BuildRepoBranchName
	gitTagPath := projectPath
	if ac.HasSeparateTagPerReleaseBranch {
		gitTagPath = filepath.Join(projectPath, eksDReleaseChannel)
	}
	// Get git tag for project if exists
	gitTag, err := tagger.GetGitTagAssigner(ac)(rc, gitTagPath)
	if err != nil {
		return nil, fmt.Errorf("error getting git tag for project %s: %v", projectName, err)
	}

	// Add project images to artifacts list
	for _, image := range ac.Images {
		imageArtifact, sourceRepoName, err := images.GetImageAssets(rc, image, ac.ImageRepoPrefix, ac.ImageTagOptions, gitTag, projectPath, gitTagPath, eksDReleaseChannel, eksDReleaseNumber, kubeVersion)
		if err != nil {
			return nil, fmt.Errorf("error getting image artifact: %v", err)
		}

		artifacts = append(artifacts, releasetypes.Artifact{Image: imageArtifact})

		imageTagOverrides = append(imageTagOverrides, releasetypes.ImageTagOverride{
			Repository: sourceRepoName,
			ReleaseUri: imageArtifact.ReleaseImageURI,
		})
	}

	// Add manifests to artifacts list
	for _, manifestComponent := range ac.Manifests {
		for _, manifestFile := range manifestComponent.ManifestFiles {
			manifestArtifact, err := manifests.GetManifestAssets(rc, manifestComponent, manifestFile, projectName, projectPath, gitTag, sourcedFromBranch, imageTagOverrides)
			if err != nil {
				return nil, fmt.Errorf("error getting manifest artifact: %v", err)
			}

			artifacts = append(artifacts, releasetypes.Artifact{Manifest: manifestArtifact})
		}
	}

	// Add archives to artifacts list
	for _, archive := range ac.Archives {
		archiveArtifact, err := archives.GetArchiveAssets(rc, archive, projectPath, gitTag, eksDReleaseChannel, eksDReleaseNumber, kubeVersion)
		if err != nil {
			return nil, fmt.Errorf("error getting archive artifact: %v", err)
		}

		artifacts = append(artifacts, releasetypes.Artifact{Archive: archiveArtifact})
	}

	return artifacts, nil
}

func GetBundleReleaseAssets(supportedK8sVersions []string, eksDReleaseMap *filereader.EksDLatestReleases, rc *releasetypes.ReleaseConfig) (map[string][]releasetypes.Artifact, error) {
	artifactsTable := map[string][]releasetypes.Artifact{}
	assetConfigs := assetconfig.GetBundleReleaseAssetsConfigMap()
	for _, release := range eksDReleaseMap.Releases {
		channel := release.Branch
		number := strconv.Itoa(release.Number)
		kubeVersion := release.KubeVersion

		if !sliceutils.SliceContains(supportedK8sVersions, channel) {
			continue
		}

		for _, assetConfig := range assetConfigs {
			if !rc.DevRelease && assetConfig.OnlyForDevRelease {
				continue
			}
			projectName := assetConfig.ProjectName
			if assetConfig.HasReleaseBranches {
				projectName = fmt.Sprintf("%s-%s", projectName, channel)
			}
			artifactsList, err := getAssetsFromConfig(&assetConfig, rc, channel, number, kubeVersion)
			if err != nil {
				return nil, errors.Wrapf(err, "Error getting artifacts for project %s", projectName)
			}
			artifactsTable[projectName] = artifactsList
		}
	}

	return artifactsTable, nil
}
