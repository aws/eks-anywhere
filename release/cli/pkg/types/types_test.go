package types_test

import (
	"testing"

	"github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func TestReleaseConfig_BundlesManifestFilepath(t *testing.T) {
	tests := []struct {
		name          string
		releaseConfig *types.ReleaseConfig
		expected      string
	}{
		{
			name: "Dev Release with non-main branch",
			releaseConfig: &types.ReleaseConfig{
				DevRelease:           true,
				ReleaseVersion:       "v0.18.10-dev+build.100",
				DevReleaseUriVersion: "v0.18.10-dev-build.100",
				BuildRepoBranchName:  "release-0.18",
			},
			expected: "release-0.18/v0.18.10-dev-build.100/bundles.yaml",
		},
		{
			name: "Dev Release with main branch",
			releaseConfig: &types.ReleaseConfig{
				DevRelease:           true,
				ReleaseVersion:       "v0.19.0-dev+build.100",
				DevReleaseUriVersion: "v0.19.0-dev-build.100",
				Weekly:               false,
				BuildRepoBranchName:  "main",
			},
			expected: "v0.19.0-dev-build.100/bundles.yaml",
		},
		{
			name: "Dev weekly Release with main branch",
			releaseConfig: &types.ReleaseConfig{
				DevRelease:           true,
				ReleaseVersion:       "v0.19.0-dev+build.100",
				DevReleaseUriVersion: "v0.19.0-dev-build.100",
				BuildRepoBranchName:  "main",
				Weekly:               true,
				ReleaseDate:          "2022-01-01",
			},
			expected: "weekly-releases/2022-01-01/bundle-release.yaml",
		},
		{
			name: "Prod Release",
			releaseConfig: &types.ReleaseConfig{
				DevRelease:     false,
				ReleaseVersion: "v0.19.0",
				BundleNumber:   123,
			},
			expected: "releases/bundles/123/manifest.yaml",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.releaseConfig.BundlesManifestFilepath()
			if result != test.expected {
				t.Errorf("Unexpected result. Expected: %s, Got: %s", test.expected, result)
			}
		})
	}
}

func TestReleaseConfig_ReleaseManifestFilepath(t *testing.T) {
	tests := []struct {
		name          string
		releaseConfig *types.ReleaseConfig
		expected      string
	}{
		{
			name: "Dev Release with non-main branch",
			releaseConfig: &types.ReleaseConfig{
				DevRelease:          true,
				BuildRepoBranchName: "feature-branch",
			},
			expected: "feature-branch/eks-a-release.yaml",
		},
		{
			name: "Dev Release with main branch",
			releaseConfig: &types.ReleaseConfig{
				DevRelease:          true,
				Weekly:              false,
				BuildRepoBranchName: "main",
			},
			expected: "eks-a-release.yaml",
		},
		{
			name: "Dev weekly Release with main branch",
			releaseConfig: &types.ReleaseConfig{
				DevRelease:          true,
				BuildRepoBranchName: "main",
				Weekly:              true,
				ReleaseDate:         "2022-01-01",
			},
			expected: "weekly-releases/2022-01-01/eks-a-release.yaml",
		},
		{
			name: "Non-DevRelease",
			releaseConfig: &types.ReleaseConfig{
				DevRelease: false,
			},
			expected: "releases/eks-a/manifest.yaml",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.releaseConfig.ReleaseManifestFilepath()
			if result != test.expected {
				t.Errorf("Unexpected result. Expected: %s, Got: %s", test.expected, result)
			}
		})
	}
}
