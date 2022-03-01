package pkg

import (
	"testing"
)

func TestGenerateNewDevReleaseVersion(t *testing.T) {
	testCases := []struct {
		testName           string
		latestBuildVersion string
		releaseVersion     string
		branch             string
		want               string
	}{
		{
			testName:           "vDev release",
			latestBuildVersion: "vDev.build.68",
			releaseVersion:     "vDev",
			branch:             "main",
			want:               "v0.0.0-dev+build.69",
		},
		{
			testName:           "vDev release with latest v0.0.0",
			latestBuildVersion: "v0.0.0-dev+build.5",
			releaseVersion:     "vDev",
			branch:             "main",
			want:               "v0.0.0-dev+build.6",
		},
		{
			testName:           "v0.0.0 release with latest vDev",
			latestBuildVersion: "vDev.build.5",
			releaseVersion:     "v0.0.0",
			branch:             "main",
			want:               "v0.0.0-dev+build.6",
		},
		{
			testName:           "v0.0.0 release with latest v0.0.0",
			latestBuildVersion: "v0.0.0-dev+build.68",
			releaseVersion:     "v0.0.0",
			branch:             "main",
			want:               "v0.0.0-dev+build.69",
		},
		{
			testName:           "different semver",
			latestBuildVersion: "v0.0.0-dev+build.5",
			releaseVersion:     "v0.0.1",
			branch:             "main",
			want:               "v0.0.1-dev+build.0",
		},
		{
			testName:           "vDev release, non-main",
			latestBuildVersion: "vDev.build.68",
			releaseVersion:     "vDev",
			branch:             "v1beta1",
			want:               "v0.0.0-dev-v1beta1+build.69",
		},
		{
			testName:           "vDev release with latest v0.0.0, non-main",
			latestBuildVersion: "v0.0.0-dev-v1beta+build.5",
			releaseVersion:     "vDev",
			branch:             "v1beta1",
			want:               "v0.0.0-dev-v1beta1+build.6",
		},
		{
			testName:           "v0.0.0 release with latest v0.0.0, non-main branch",
			latestBuildVersion: "v0.0.0-dev-v1beta1+build.0",
			releaseVersion:     "v0.0.0",
			branch:             "v1beta1",
			want:               "v0.0.0-dev-v1beta1+build.1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			if got, err := generateNewDevReleaseVersion(tt.latestBuildVersion, tt.releaseVersion, tt.branch); err != nil {
				t.Fatalf("generateNewDevReleaseVersion err = %s, want err = nil", err)
			} else if got != tt.want {
				t.Fatalf("generateNewDevReleaseVersion version = %s, want %s", got, tt.want)
			}
		})
	}
}
