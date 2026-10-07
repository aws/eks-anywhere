// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package images

import (
	"testing"

	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func TestHelmCompatibleGitTag(t *testing.T) {
	tests := map[string]struct {
		gitTag string
		want   string
	}{
		"release tag": {
			gitTag: "v0.26.0",
			want:   "0.26.0",
		},
		"prerelease tag": {
			gitTag: "v0.26.0-rc.1",
			want:   "0.26.0-rc.1",
		},
		"commit SHA": {
			gitTag: "0123456789abcdef",
			want:   "0.0.1-0123456789abcdef",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := helmCompatibleGitTag(test.gitTag); got != test.want {
				t.Fatalf("helmCompatibleGitTag() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestGetSourceImageURIUsesHelmCompatibleTag(t *testing.T) {
	releaseConfig := &releasetypes.ReleaseConfig{
		BuildRepoBranchName:     "main",
		SourceContainerRegistry: "source.example",
		ReleaseEnvironment:      "development",
		DevRelease:              true,
		DryRun:                  true,
		BuildRepoSource:         "build-tooling",
	}
	tagConfig := assettypes.ImageTagConfiguration{
		NonProdSourceImageTagFormat: "<gitTag>",
		UseHelmCompatibleSourceTag:  true,
	}

	tests := map[string]struct {
		gitTag string
		want   string
	}{
		"release tag": {
			gitTag: "v0.26.0",
			want:   "source.example/tinkerbell/tinkerbell-crds:0.26.0-latest-helm",
		},
		"commit SHA": {
			gitTag: "0123456789abcdef",
			want:   "source.example/tinkerbell/tinkerbell-crds:0.0.1-0123456789abcdef-latest-helm",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, branch, err := GetSourceImageURI(
				releaseConfig,
				"tinkerbell-crds-helm",
				"tinkerbell/tinkerbell-crds",
				map[string]string{
					"gitTag":      test.gitTag,
					"projectPath": "projects/tinkerbell/tinkerbell-crds",
				},
				tagConfig,
				true,
				false,
			)
			if err != nil {
				t.Fatalf("GetSourceImageURI() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("GetSourceImageURI() = %q, want %q", got, test.want)
			}
			if branch != "main" {
				t.Fatalf("GetSourceImageURI() branch = %q, want main", branch)
			}
		})
	}
}
