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

package filereader

import (
	"testing"
)

func TestNewBuildNumberFromLastVersion(t *testing.T) {
	testCases := []struct {
		testName           string
		latestBuildVersion string
		releaseVersion     string
		branch             string
		want               int
	}{
		{
			testName:           "vDev release with latest v0.0.0",
			latestBuildVersion: "v0.0.0-dev+build.5",
			releaseVersion:     "vDev",
			branch:             "main",
			want:               6,
		},
		{
			testName:           "v0.0.0 release with latest v0.0.0",
			latestBuildVersion: "v0.0.0-dev+build.68",
			releaseVersion:     "v0.0.0",
			branch:             "main",
			want:               69,
		},
		{
			testName:           "different semver",
			latestBuildVersion: "v0.0.0-dev+build.5",
			releaseVersion:     "v0.0.1",
			branch:             "main",
			want:               0,
		},
		{
			testName:           "vDev release with latest v0.0.0, non-main",
			latestBuildVersion: "v0.0.0-dev-v1beta+build.5",
			releaseVersion:     "vDev",
			branch:             "v1beta1",
			want:               6,
		},
		{
			testName:           "v0.0.0 release with latest v0.0.0, non-main branch",
			latestBuildVersion: "v0.0.0-dev-v1beta1+build.0",
			releaseVersion:     "v0.0.0",
			branch:             "v1beta1",
			want:               1,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			if got, err := NewBuildNumberFromLastVersion(tt.latestBuildVersion, tt.releaseVersion, tt.branch); err != nil {
				t.Fatalf("NewBuildNumberFromLastVersion err = %s, want err = nil", err)
			} else if got != tt.want {
				t.Fatalf("NewBuildNumberFromLastVersion version = %d, want %d", got, tt.want)
			}
		})
	}
}
