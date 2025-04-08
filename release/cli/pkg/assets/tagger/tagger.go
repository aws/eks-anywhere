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

package tagger

import (
	"strings"

	"github.com/pkg/errors"

	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	"github.com/aws/eks-anywhere/release/cli/pkg/git"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

// BuildToolingGitTagAssigner reads the Git tag from the eks-anywhere-build-tooling repository using the branch name.
// If overrideBranch is provided, it takes precedence over the default branch from the release config.
func BuildToolingGitTagAssigner(rc *releasetypes.ReleaseConfig, gitTagPath, overrideBranch string) (string, error) {
	branchName := rc.BuildRepoBranchName
	if overrideBranch != "" {
		branchName = overrideBranch
	}
	gitTag, err := filereader.ReadGitTag(gitTagPath, rc.BuildRepoSource, branchName)
	if err != nil {
		return "", errors.Cause(err)
	}

	return gitTag, nil
}

// CliGitTagAssigner determines the Git tag to use for the CLI repository based on the release configuration.
// If the release is a development release (DevRelease is true), it fetches the list of version tags from the repository
// and returns the most recent (highest) tag in descending semantic version order.
// Otherwise, it uses the explicitly defined ReleaseVersion from the release configuration.
func CliGitTagAssigner(rc *releasetypes.ReleaseConfig, gitTagPath, overrideBranch string) (string, error) {
	var gitTag string

	if rc.DevRelease {
		tagList, err := git.GetRepoTagsDescending(rc.CliRepoSource)
		if err != nil {
			return "", errors.Cause(err)
		}
		gitTag = strings.Split(tagList, "\n")[0]
	} else {
		gitTag = rc.ReleaseVersion
	}

	return gitTag, nil
}

// NonExistentTagAssigner is a placeholder GitTagAssigner that always returns the tag "non-existent".
// This can be used in scenarios where Git tagging is irrelevant.
func NonExistentTagAssigner(rc *releasetypes.ReleaseConfig, gitTagPath, overrideBranch string) (string, error) {
	return "non-existent", nil
}

// GetGitTagAssigner returns the GitTagAssigner function to be used for the asset configuration.
// If a custom GitTagAssigner is defined in the AssetConfig, it returns that.
// Otherwise, it defaults to using the BuildToolingGitTagAssigner.
func GetGitTagAssigner(ac *assettypes.AssetConfig) assettypes.GitTagAssigner {
	if ac.GitTagAssigner != nil {
		return assettypes.GitTagAssigner(ac.GitTagAssigner)
	}
	return assettypes.GitTagAssigner(BuildToolingGitTagAssigner)
}
