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

package manifests

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

func TestGenerateManifestAssets(t *testing.T) {
	testCases := []struct {
		testName             string
		manifestComponent    *assettypes.ManifestComponent
		manifestFile         string
		imageTagOverrides    []releasetypes.ImageTagOverride
		buildRepoBranchName  string
		projectName          string
		projectPath          string
		gitTag               string
		wantManifestArtifact *releasetypes.ManifestArtifact
		wantErr              bool
	}{
		{
			testName:            "Manifest artifact for project foo/bar from main",
			buildRepoBranchName: "main",
			projectName:         "bar",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.1.0",
			manifestComponent: &assettypes.ManifestComponent{
				Name:                  "bar",
				ReleaseManifestPrefix: "bar-manifests",
			},
			manifestFile: "components.yaml",
			imageTagOverrides: []releasetypes.ImageTagOverride{
				{
					Repository: "foo/bar",
					ReleaseUri: "release-container-registry/foo/bar:v0.1.0-eks-a-v0.0.0-dev-build.1",
				},
			},
			wantManifestArtifact: &releasetypes.ManifestArtifact{
				SourceS3Prefix: "projects/foo/bar/latest/manifests/bar/v0.1.0",
				SourceS3Key:    "components.yaml",
				ArtifactPath:   "artifacts/bar-manifests",
				ReleaseName:    "components.yaml",
				ReleaseS3Path:  "artifacts/v0.0.0-dev-build.0/bar-manifests/manifests/bar/v0.1.0",
				ReleaseCdnURI:  "https://release-bucket/artifacts/v0.0.0-dev-build.0/bar-manifests/manifests/bar/v0.1.0/components.yaml",
				ImageTagOverrides: []releasetypes.ImageTagOverride{
					{
						Repository: "foo/bar",
						ReleaseUri: "release-container-registry/foo/bar:v0.1.0-eks-a-v0.0.0-dev-build.1",
					},
				},
				GitTag:            "v0.1.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "main",
				Component:         "bar",
			},
			wantErr: false,
		},
		{
			testName:            "Manifest artifact for project foo/bar from release-branch",
			buildRepoBranchName: "release-branch",
			projectName:         "bar",
			projectPath:         "projects/foo/bar",
			gitTag:              "v0.1.0",
			manifestComponent: &assettypes.ManifestComponent{
				Name:                  "bar",
				ReleaseManifestPrefix: "bar-manifests",
			},
			manifestFile: "components.yaml",
			imageTagOverrides: []releasetypes.ImageTagOverride{
				{
					Repository: "foo/bar",
					ReleaseUri: "release-container-registry/foo/bar:v0.1.0-eks-a-v0.0.0-dev-release-branch-build.1",
				},
			},
			wantManifestArtifact: &releasetypes.ManifestArtifact{
				SourceS3Prefix: "projects/foo/bar/release-branch/manifests/bar/v0.1.0",
				SourceS3Key:    "components.yaml",
				ArtifactPath:   "artifacts/bar-manifests",
				ReleaseName:    "components.yaml",
				ReleaseS3Path:  "artifacts/v0.0.0-dev-release-branch-build.0/bar-manifests/manifests/bar/v0.1.0",
				ReleaseCdnURI:  "https://release-bucket/artifacts/v0.0.0-dev-release-branch-build.0/bar-manifests/manifests/bar/v0.1.0/components.yaml",
				ImageTagOverrides: []releasetypes.ImageTagOverride{
					{
						Repository: "foo/bar",
						ReleaseUri: "release-container-registry/foo/bar:v0.1.0-eks-a-v0.0.0-dev-release-branch-build.1",
					},
				},
				GitTag:            "v0.1.0",
				ProjectPath:       "projects/foo/bar",
				SourcedFromBranch: "release-branch",
				Component:         "bar",
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

			if gotManifestArtifact, err := GetManifestAssets(releaseConfig, tt.manifestComponent, tt.manifestFile, tt.projectName, tt.projectPath, tt.gitTag, tt.buildRepoBranchName, tt.imageTagOverrides); (err != nil) != tt.wantErr {
				t.Fatalf("GetManifestAssets got err = %v, want err = %v", err, tt.wantErr)
			} else if !reflect.DeepEqual(gotManifestArtifact, tt.wantManifestArtifact) {
				t.Fatalf("GetManifestAssets got artifact = %v, expected %v", gotManifestArtifact, tt.wantManifestArtifact)
			}
		})
	}
}
