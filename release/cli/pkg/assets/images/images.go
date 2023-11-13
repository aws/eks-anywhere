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

package images

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/cli/pkg/assets/tagger"
	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
	"github.com/aws/eks-anywhere/release/cli/pkg/images"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func GetImageAssets(rc *releasetypes.ReleaseConfig, ac *assettypes.AssetConfig, image *assettypes.Image, imageRepoPrefix string, imageTagOptions []string, gitTag, projectPath, gitTagPath, eksDReleaseChannel, eksDReleaseNumber, kubeVersion string) (*releasetypes.ImageArtifact, string, error) {
	repoName, assetName := image.RepoName, image.RepoName
	if image.AssetName != "" {
		assetName = image.AssetName
	}

	if imageRepoPrefix != "" {
		repoName = fmt.Sprintf("%s/%s", imageRepoPrefix, repoName)
	}
	sourceRepoName, releaseRepoName := repoName, repoName
	if image.TrimEksAPrefix {
		if rc.ReleaseEnvironment == "production" {
			sourceRepoName = strings.TrimPrefix(repoName, "eks-anywhere-")
		}
		if !rc.DevRelease {
			releaseRepoName = strings.TrimPrefix(repoName, "eks-anywhere-")
		}
	}

	imageTagOptionsMap := map[string]string{}
	for _, opt := range imageTagOptions {
		switch opt {
		case "gitTag":
			imageTagOptionsMap[opt] = gitTag
		case "projectPath":
			imageTagOptionsMap[opt] = projectPath
		case "eksDReleaseChannel":
			imageTagOptionsMap[opt] = eksDReleaseChannel
		case "eksDReleaseNumber":
			imageTagOptionsMap[opt] = eksDReleaseNumber
		case "kubeVersion":
			imageTagOptionsMap[opt] = kubeVersion
		case "buildRepoSourceRevision":
			imageTagOptionsMap[opt] = rc.BuildRepoHead
		default:
			return nil, "", fmt.Errorf("error configuring image tag options: invalid option: %s", opt)
		}
	}

	sourceImageUri, sourcedFromBranch, err := images.GetSourceImageURI(rc, assetName, sourceRepoName, imageTagOptionsMap, image.ImageTagConfiguration, image.TrimVersionSignifier, ac.HasSeparateTagPerReleaseBranch)
	if err != nil {
		return nil, "", errors.Cause(err)
	}
	if sourcedFromBranch != rc.BuildRepoBranchName {
		gitTag, err := tagger.GetGitTagAssigner(ac)(rc, gitTagPath, sourcedFromBranch)
		if err != nil {
			return nil, "", errors.Cause(err)
		}
		imageTagOptionsMap["gitTag"] = gitTag
	}

	releaseImageUri, err := images.GetReleaseImageURI(rc, assetName, releaseRepoName, imageTagOptionsMap, image.ImageTagConfiguration, image.TrimVersionSignifier, ac.HasSeparateTagPerReleaseBranch)
	if err != nil {
		return nil, "", errors.Cause(err)
	}

	imageArtifact := &releasetypes.ImageArtifact{
		AssetName:         assetName,
		SourceImageURI:    sourceImageUri,
		ReleaseImageURI:   releaseImageUri,
		Arch:              []string{"amd64", "arm64"},
		OS:                "linux",
		GitTag:            gitTag,
		ProjectPath:       projectPath,
		SourcedFromBranch: sourcedFromBranch,
	}

	return imageArtifact, sourceRepoName, nil
}
