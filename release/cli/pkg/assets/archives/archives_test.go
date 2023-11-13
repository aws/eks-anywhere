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

package archives

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

func TestGenerateArchiveAssets(t *testing.T) {
	testCases := []struct {
		testName            string
		archive             *assettypes.Archive
		buildRepoBranchName string
		projectPath         string
		gitTag              string
		eksDReleaseChannel  string
		eksDReleaseNumber   string
		kubeVersion         string
		wantArchiveArtifact *releasetypes.ArchiveArtifact
		wantErr             bool
	}{
		{
			testName:            "Tarball archive for project foo/bar from main",
			buildRepoBranchName: "main",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.1.0",
			eksDReleaseChannel:  "1-21",
			eksDReleaseNumber:   "8",
			kubeVersion:         "1.21.9",
			archive: &assettypes.Archive{
				Name:   "baz",
				Format: "tarball",
			},
			wantArchiveArtifact: &releasetypes.ArchiveArtifact{
				SourceS3Key:       "baz-linux-amd64-v0.1.0.tar.gz",
				SourceS3Prefix:    "projects/foo/bar/latest",
				ArtifactPath:      "artifacts/baz-tarball/1-21",
				ReleaseName:       "baz-v0.0.0-dev-build.0-linux-amd64.tar.gz",
				ReleaseS3Path:     "artifacts/v0.0.0-dev-build.0/baz/v0.1.0",
				ReleaseCdnURI:     "https://release-bucket/artifacts/v0.0.0-dev-build.0/baz/v0.1.0/baz-v0.0.0-dev-build.0-linux-amd64.tar.gz",
				OS:                "linux",
				Arch:              []string{"amd64"},
				GitTag:            "v0.1.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "main",
				ImageFormat:       "tarball",
			},
			wantErr: false,
		},
		{
			testName:            "Tarball archive for project foo/bar from release-branch",
			buildRepoBranchName: "release-branch",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.2.0",
			eksDReleaseChannel:  "1-22",
			eksDReleaseNumber:   "6",
			kubeVersion:         "1.22.6",
			archive: &assettypes.Archive{
				Name:   "baz",
				Format: "tarball",
			},
			wantArchiveArtifact: &releasetypes.ArchiveArtifact{
				SourceS3Key:       "baz-linux-amd64-v0.2.0.tar.gz",
				SourceS3Prefix:    "projects/foo/bar/release-branch",
				ArtifactPath:      "artifacts/baz-tarball/1-22",
				ReleaseName:       "baz-v0.0.0-dev-release-branch-build.0-linux-amd64.tar.gz",
				ReleaseS3Path:     "artifacts/v0.0.0-dev-release-branch-build.0/baz/v0.2.0",
				ReleaseCdnURI:     "https://release-bucket/artifacts/v0.0.0-dev-release-branch-build.0/baz/v0.2.0/baz-v0.0.0-dev-release-branch-build.0-linux-amd64.tar.gz",
				OS:                "linux",
				Arch:              []string{"amd64"},
				GitTag:            "v0.2.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "release-branch",
				ImageFormat:       "tarball",
			},
			wantErr: false,
		},
		{
			testName:            "OS image archive for project foo/bar from main",
			buildRepoBranchName: "main",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.1.0",
			eksDReleaseChannel:  "1-21",
			eksDReleaseNumber:   "8",
			kubeVersion:         "1.21.9",
			archive: &assettypes.Archive{
				Name:                "baz",
				OSName:              "lorem",
				OSVersion:           "v1.2.3",
				Format:              "ova",
				ArchiveS3PathGetter: EksDistroArtifactPathGetter,
			},
			wantArchiveArtifact: &releasetypes.ArchiveArtifact{
				SourceS3Key:       "lorem.ova",
				SourceS3Prefix:    "projects/foo/bar/1-21/ova/lorem/v1.2.3/latest",
				ArtifactPath:      "artifacts/baz-ova/1-21",
				ReleaseName:       "lorem-1.21.9-eks-d-1-21-8-eks-a-v0.0.0-dev-build.0-amd64.ova",
				ReleaseS3Path:     "artifacts/v0.0.0-dev-build.0/eks-distro/ova/1-21/1-21-8",
				ReleaseCdnURI:     "https://release-bucket/artifacts/v0.0.0-dev-build.0/eks-distro/ova/1-21/1-21-8/lorem-1.21.9-eks-d-1-21-8-eks-a-v0.0.0-dev-build.0-amd64.ova",
				OS:                "linux",
				OSName:            "lorem",
				Arch:              []string{"amd64"},
				GitTag:            "v0.1.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "main",
				ImageFormat:       "ova",
			},
			wantErr: false,
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

			if gotArchiveArtifact, err := GetArchiveAssets(releaseConfig, tt.archive, tt.projectPath, tt.gitTag, tt.eksDReleaseChannel, tt.eksDReleaseNumber, tt.kubeVersion); (err != nil) != tt.wantErr {
				t.Fatalf("GetArchiveAssets err = %v, want err = %v", err, tt.wantErr)
			} else if !reflect.DeepEqual(gotArchiveArtifact, tt.wantArchiveArtifact) {
				t.Fatalf("GetArchiveAssets got artifact = %v, expected %v", gotArchiveArtifact, tt.wantArchiveArtifact)
			}
		})
	}
}
