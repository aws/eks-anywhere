package pkg

import (
	"testing"
)

func TestGenerateNewDevReleaseVersion(t *testing.T) {
	testCases := []struct {
		testName           string
		latestBuildVersion string
		releaseVersion     string
		want               string
	}{
		{
			testName:           "vDev release",
			latestBuildVersion: "vDev.build.68",
			releaseVersion:     "vDev",
			want:               "v0.0.0-dev+build.69",
		},
		{
			testName:           "vDev release with latest v0.0.0",
			latestBuildVersion: "v0.0.0-dev+build.5",
			releaseVersion:     "vDev",
			want:               "v0.0.0-dev+build.6",
		},
		{
			testName:           "v0.0.0 release with latest vDev",
			latestBuildVersion: "vDev.build.5",
			releaseVersion:     "v0.0.0",
			want:               "v0.0.0-dev+build.6",
		},
		{
			testName:           "v0.0.0 release with latest v0.0.0",
			latestBuildVersion: "v0.0.0-dev+build.68",
			releaseVersion:     "v0.0.0",
			want:               "v0.0.0-dev+build.69",
		},
		{
			testName:           "different semver",
			latestBuildVersion: "v0.0.0-dev+build.5",
			releaseVersion:     "v0.0.1",
			want:               "v0.0.1-dev+build.0",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			if got, err := generateNewDevReleaseVersion(tt.latestBuildVersion, tt.releaseVersion); err != nil {
				t.Fatalf("generateNewDevReleaseVersion err = %s, want err = nil", err)
			} else if got != tt.want {
				t.Fatalf("generateNewDevReleaseVersion version = %s, want %s", got, tt.want)
			}
		})
	}
}
