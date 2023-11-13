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
	"reflect"
	"strings"
	"testing"
	"time"

	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

var releaseConfig = &releasetypes.ReleaseConfig{
	ArtifactDir:              "artifacts",
	CliRepoSource:            "eks-a-build",
	BuildRepoSource:          "eks-a-cli",
	CliRepoBranchName:        "main",
	CliRepoUrl:               "https://github.com/aws/eks-anywhere.git",
	BuildRepoUrl:             "https://github.com/aws/eks-anywhere-build-tooling.git",
	SourceBucket:             "projectbuildpipeline-857-pipelineoutputartifactsb-10ajmk30khe3f",
	ReleaseBucket:            "release-bucket",
	SourceContainerRegistry:  "source-container-registry",
	ReleaseContainerRegistry: "release-container-registry",
	CDN:                      "https://release-bucket",
	BundleNumber:             1,
	ReleaseNumber:            1,
	ReleaseVersion:           "vDev",
	ReleaseTime:              time.Unix(0, 0),
	DevRelease:               true,
	DryRun:                   true,
}

func TestGenerateImageAssets(t *testing.T) {
	testCases := []struct {
		testName            string
		image               *assettypes.Image
		imageRepoPrefix     string
		imageTagOptions     []string
		assetConfig         *assettypes.AssetConfig
		buildRepoBranchName string
		projectPath         string
		gitTag              string
		eksDReleaseChannel  string
		eksDReleaseNumber   string
		kubeVersion         string
		wantImageArtifact   *releasetypes.ImageArtifact
		wantErr             bool
	}{
		{
			testName:            "Image artifact for project foo/bar from main",
			buildRepoBranchName: "main",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.1.0",
			eksDReleaseChannel:  "1-21",
			eksDReleaseNumber:   "8",
			kubeVersion:         "1.21.9",
			assetConfig:         &assettypes.AssetConfig{},
			image: &assettypes.Image{
				RepoName: "bar",
			},
			imageRepoPrefix: "foo",
			imageTagOptions: []string{"gitTag"},
			wantImageArtifact: &releasetypes.ImageArtifact{
				AssetName:         "bar",
				SourceImageURI:    "source-container-registry/foo/bar:latest",
				ReleaseImageURI:   "release-container-registry/foo/bar:v0.1.0-eks-a-v0.0.0-dev-build.1",
				OS:                "linux",
				Arch:              []string{"amd64", "arm64"},
				GitTag:            "v0.1.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "main",
			},
			wantErr: false,
		},
		{
			testName:            "Image artifact for project foo/bar from release-branch",
			buildRepoBranchName: "release-branch",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.2.0",
			eksDReleaseChannel:  "1-22",
			eksDReleaseNumber:   "5",
			kubeVersion:         "1.22.4",
			assetConfig:         &assettypes.AssetConfig{},
			image: &assettypes.Image{
				RepoName: "bar",
			},
			imageRepoPrefix: "foo",
			imageTagOptions: []string{"gitTag"},
			wantImageArtifact: &releasetypes.ImageArtifact{
				AssetName:         "bar",
				SourceImageURI:    "source-container-registry/foo/bar:release-branch",
				ReleaseImageURI:   "release-container-registry/foo/bar:v0.2.0-eks-a-v0.0.0-dev-release-branch-build.1",
				OS:                "linux",
				Arch:              []string{"amd64", "arm64"},
				GitTag:            "v0.2.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "release-branch",
			},
		},
		{
			testName:            "Image artifact for project foo/bar from main with asset name override",
			buildRepoBranchName: "main",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.1.0",
			eksDReleaseChannel:  "1-21",
			eksDReleaseNumber:   "8",
			kubeVersion:         "1.21.9",
			assetConfig:         &assettypes.AssetConfig{},
			image: &assettypes.Image{
				RepoName:  "bar",
				AssetName: "lorem-ipsum",
			},
			imageRepoPrefix: "foo",
			imageTagOptions: []string{"gitTag"},
			wantImageArtifact: &releasetypes.ImageArtifact{
				AssetName:         "lorem-ipsum",
				SourceImageURI:    "source-container-registry/foo/bar:latest",
				ReleaseImageURI:   "release-container-registry/foo/bar:v0.1.0-eks-a-v0.0.0-dev-build.1",
				OS:                "linux",
				Arch:              []string{"amd64", "arm64"},
				GitTag:            "v0.1.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "main",
			},
			wantErr: false,
		},
		{
			testName:            "Image artifact for project foo/bar from main with custom tagging configurations",
			buildRepoBranchName: "main",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.3.0",
			eksDReleaseChannel:  "1-21",
			eksDReleaseNumber:   "8",
			kubeVersion:         "1.21.9",
			assetConfig:         &assettypes.AssetConfig{},
			image: &assettypes.Image{
				RepoName:  "bar",
				AssetName: "custom-bar",
				ImageTagConfiguration: assettypes.ImageTagConfiguration{
					NonProdSourceImageTagFormat: "<gitTag>-<kubeVersion>-baz-<eksDReleaseChannel>-bar",
					ReleaseImageTagFormat:       "<eksDReleaseChannel>-<eksDReleaseNumber>-<kubeVersion>-baz-bar",
				},
			},
			imageRepoPrefix: "foo",
			imageTagOptions: []string{"gitTag", "eksDReleaseChannel", "eksDReleaseNumber", "kubeVersion"},
			wantImageArtifact: &releasetypes.ImageArtifact{
				AssetName:         "custom-bar",
				SourceImageURI:    "source-container-registry/foo/bar:v0.3.0-1.21.9-baz-1-21-bar-latest",
				ReleaseImageURI:   "release-container-registry/foo/bar:1-21-8-1.21.9-baz-bar-eks-a-v0.0.0-dev-build.1",
				OS:                "linux",
				Arch:              []string{"amd64", "arm64"},
				GitTag:            "v0.3.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "main",
			},
			wantErr: false,
		},
		{
			testName:            "Image artifact for project foo/bar from main with incorrect image tag option",
			buildRepoBranchName: "main",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.1.0",
			eksDReleaseChannel:  "1-21",
			eksDReleaseNumber:   "8",
			kubeVersion:         "1.21.9",
			assetConfig:         &assettypes.AssetConfig{},
			image: &assettypes.Image{
				RepoName: "bar",
			},
			imageRepoPrefix:   "foo",
			imageTagOptions:   []string{"non-existent-option"},
			wantImageArtifact: nil,
			wantErr:           true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			releaseConfig.BuildRepoBranchName = tt.buildRepoBranchName

			releaseVersion, err := filereader.GetCurrentEksADevReleaseVersion(releaseConfig.ReleaseVersion, releaseConfig, 0)
			if err != nil {
				t.Fatalf("Error getting previous EKS-A dev release number: %v\n", err)
			}

			releaseConfig.ReleaseVersion = releaseVersion
			releaseConfig.DevReleaseUriVersion = strings.ReplaceAll(releaseVersion, "+", "-")

			if gotImageArtifact, _, err := GetImageAssets(releaseConfig, tt.assetConfig, tt.image, tt.imageRepoPrefix, tt.imageTagOptions, tt.gitTag, tt.projectPath, tt.projectPath, tt.eksDReleaseChannel, tt.eksDReleaseNumber, tt.kubeVersion); (err != nil) != tt.wantErr {
				t.Fatalf("GetImageAssets got err = %v, want err = %v", err, tt.wantErr)
			} else if !reflect.DeepEqual(gotImageArtifact, tt.wantImageArtifact) {
				t.Fatalf("GetImageAssets got artifact = %v, expected %v", gotImageArtifact, tt.wantImageArtifact)
			}
		})
	}
}
