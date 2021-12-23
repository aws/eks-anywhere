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

func TestGetEksDKubeVersion(t *testing.T) {
	testCases := []struct {
		testName       string
		releaseChannel string
		releaseNumber  string
		want           string
	}{
		{
			testName:       "EKS Distro 1-18-1 release",
			releaseChannel: "1-18",
			releaseNumber:  "1",
			want:           "v1.18.9",
		},
		{
			testName:       "EKS Distro 1-18-13 release",
			releaseChannel: "1-18",
			releaseNumber:  "13",
			want:           "v1.18.20",
		},
		{
			testName:       "EKS Distro 1-19-1 release",
			releaseChannel: "1-19",
			releaseNumber:  "1",
			want:           "v1.19.6",
		},
		{
			testName:       "EKS Distro 1-19-12 release",
			releaseChannel: "1-19",
			releaseNumber:  "12",
			want:           "v1.19.15",
		},
		{
			testName:       "EKS Distro 1-20-1 release",
			releaseChannel: "1-20",
			releaseNumber:  "1",
			want:           "v1.20.4",
		},
		{
			testName:       "EKS Distro 1-20-9 release",
			releaseChannel: "1-20",
			releaseNumber:  "9",
			want:           "v1.20.11",
		},
		{
			testName:       "EKS Distro 1-21-1 release",
			releaseChannel: "1-21",
			releaseNumber:  "1",
			want:           "v1.21.2",
		},
		{
			testName:       "EKS Distro 1-21-7 release",
			releaseChannel: "1-21",
			releaseNumber:  "7",
			want:           "v1.21.5",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			if got, err := getEksDKubeVersion(tt.releaseChannel, tt.releaseNumber); err != nil {
				t.Fatalf("getEksDKubeVersion err = %s, want err = nil", err)
			} else if got != tt.want {
				t.Fatalf("getEksDKubeVersion kubeVersion = %s, want %s", got, tt.want)
			}
		})
	}
}
