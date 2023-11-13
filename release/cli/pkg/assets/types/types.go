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

package types

import (
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

type ManifestComponent struct {
	Name                  string
	ReleaseManifestPrefix string
	ManifestFiles         []string
	NoVersionSuffix       bool
}

type ImageTagConfiguration struct {
	SourceLatestTagFromECR      bool
	NonProdSourceImageTagFormat string
	ProdSourceImageTagFormat    string
	ReleaseImageTagFormat       string
}

type Image struct {
	AssetName             string
	RepoName              string
	TrimEksAPrefix        bool
	ImageTagConfiguration ImageTagConfiguration
	TrimVersionSignifier  bool
}

type Archive struct {
	Name                 string
	Format               string
	OSName               string
	OSVersion            string
	ArchitectureOverride string
	ArchiveS3PathGetter  ArchiveS3PathGenerator
}

type AssetConfig struct {
	ProjectName                    string
	ProjectPath                    string
	GitTagAssigner                 GitTagAssigner
	Archives                       []*Archive
	Images                         []*Image
	ImageRepoPrefix                string
	ImageTagOptions                []string
	Manifests                      []*ManifestComponent
	NoGitTag                       bool
	HasReleaseBranches             bool
	HasSeparateTagPerReleaseBranch bool
	OnlyForDevRelease              bool
	UsesKubeRbacProxy              bool
}

type ArchiveS3PathGenerator func(rc *releasetypes.ReleaseConfig, archive *Archive, projectPath, gitTag, eksDReleaseChannel, eksDReleaseNumber, kubeVersion, latestPath, arch string) (string, string, string, string, error)

type GitTagAssigner func(rc *releasetypes.ReleaseConfig, gitTagPath, overrideBranch string) (string, error)
