package semver_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/semver"
)

func TestNewError(t *testing.T) {
	testCases := []struct {
		testName string
		version  string
	}{
		{
			testName: "empty",
			version:  "",
		},
		{
			testName: "only letters",
			version:  "xxx",
		},
		{
			testName: "only mayor",
			version:  "11",
		},
		{
			testName: "no patch",
			version:  "11.1",
		},
		{
			testName: "dot after patch",
			version:  "11.1.1.1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			if _, err := semver.New(tt.version); err == nil {
				t.Fatalf("semver.New(%s) err = nil, want err not nil", tt.version)
			}
		})
	}
}

func TestNewSuccess(t *testing.T) {
	testCases := []struct {
		testName string
		version  string
		want     *semver.Version
	}{
		{
			testName: "only patch",
			version:  "0.0.4",
			want: &semver.Version{
				Major: 0,
				Minor: 0,
				Patch: 4,
			},
		},
		{
			testName: "only patch with double digit numbers",
			version:  "10.20.30",
			want: &semver.Version{
				Major: 10,
				Minor: 20,
				Patch: 30,
			},
		},
		{
			testName: "prerelease and meta",
			version:  "1.1.2-prerelease+meta",
			want: &semver.Version{
				Major:         1,
				Minor:         1,
				Patch:         2,
				Prerelease:    "prerelease",
				Buildmetadata: "meta",
			},
		},
		{
			testName: "only meta with hyphen",
			version:  "1.1.2+meta-valid",
			want: &semver.Version{
				Major:         1,
				Minor:         1,
				Patch:         2,
				Buildmetadata: "meta-valid",
			},
		},
		{
			testName: "prerelease and build with dots",
			version:  "2.0.0-rc.1+build.123",
			want: &semver.Version{
				Major:         2,
				Minor:         0,
				Patch:         0,
				Prerelease:    "rc.1",
				Buildmetadata: "build.123",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			if got, err := semver.New(tt.version); err != nil {
				t.Fatalf("semver.New(%s) err = %s, want err = nil", tt.version, err)
			} else if !got.Equal(tt.want) {
				t.Fatalf("semver.New(%s) semver = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
