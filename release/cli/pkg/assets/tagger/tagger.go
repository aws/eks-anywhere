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

func NonExistentTagAssigner(rc *releasetypes.ReleaseConfig, gitTagPath, overrideBranch string) (string, error) {
	return "non-existent", nil
}

func GetGitTagAssigner(ac *assettypes.AssetConfig) assettypes.GitTagAssigner {
	if ac.GitTagAssigner != nil {
		return assettypes.GitTagAssigner(ac.GitTagAssigner)
	}
	return assettypes.GitTagAssigner(BuildToolingGitTagAssigner)
}
