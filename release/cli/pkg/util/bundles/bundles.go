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

package bundles

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"

	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	"github.com/aws/eks-anywhere/release/cli/pkg/images"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func SortArtifactsMap(m map[string][]releasetypes.Artifact) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func getKubeRbacProxyImageAttributes(r *releasetypes.ReleaseConfig) (string, string, map[string]string, error) {
	gitTag, err := filereader.ReadGitTag(constants.KubeRbacProxyProjectPath, r.BuildRepoSource, r.BuildRepoBranchName)
	if err != nil {
		return "", "", nil, errors.Cause(err)
	}
	name := "kube-rbac-proxy"
	repoName := fmt.Sprintf("brancz/%s", name)
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": constants.KubeRbacProxyProjectPath,
	}

	return name, repoName, tagOptions, nil
}

func GetKubeRbacProxyImageTagOverride(r *releasetypes.ReleaseConfig) (releasetypes.ImageTagOverride, error) {
	name, repoName, tagOptions, err := getKubeRbacProxyImageAttributes(r)
	if err != nil {
		return releasetypes.ImageTagOverride{}, errors.Cause(err)
	}

	releaseImageUri, err := images.GetReleaseImageURI(r, name, repoName, tagOptions, assettypes.ImageTagConfiguration{}, false, false)
	if err != nil {
		return releasetypes.ImageTagOverride{}, errors.Cause(err)
	}
	imageTagOverride := releasetypes.ImageTagOverride{
		Repository: repoName,
		ReleaseUri: releaseImageUri,
	}

	return imageTagOverride, nil
}
