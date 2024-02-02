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

func TestCompareSuccess(t *testing.T) {
	testCases := []struct {
		testName string
		v1       *semver.Version
		v2       *semver.Version
		want     int
	}{
		{
			testName: "equal",
			v1: &semver.Version{
				Major: 1,
				Minor: 0,
				Patch: 4,
			},
			v2: &semver.Version{
				Major: 1,
				Minor: 0,
				Patch: 4,
			},
			want: 0,
		},
		{
			testName: "less than",
			v1: &semver.Version{
				Major: 1,
				Minor: 0,
				Patch: 3,
			},
			v2: &semver.Version{
				Major: 1,
				Minor: 0,
				Patch: 4,
			},
			want: -1,
		},
		{
			testName: "less than, diff major",
			v1: &semver.Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			v2: &semver.Version{
				Major: 2,
				Minor: 0,
				Patch: 4,
			},
			want: -1,
		},
		{
			testName: "greater than",
			v1: &semver.Version{
				Major: 1,
				Minor: 0,
				Patch: 5,
			},
			v2: &semver.Version{
				Major: 1,
				Minor: 0,
				Patch: 4,
			},
			want: 1,
		},
		{
			testName: "greater than, diff major",
			v1: &semver.Version{
				Major: 2,
				Minor: 1,
				Patch: 3,
			},
			v2: &semver.Version{
				Major: 1,
				Minor: 2,
				Patch: 4,
			},
			want: 1,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			got := tt.v1.Compare(tt.v2)
			if got != tt.want {
				t.Fatalf("semver.Compare() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareBuildMetadata(t *testing.T) {
	testCases := []struct {
		testName string
		v1       *semver.Version
		v2       *semver.Version
		want     int
	}{
		{
			testName: "equal build metadata only strings",
			v1: &semver.Version{
				Buildmetadata: "f234f.werwe",
			},
			v2: &semver.Version{
				Buildmetadata: "f234f.werwe",
			},
			want: 0,
		},
		{
			testName: "equal build metadata strings and numbers",
			v1: &semver.Version{
				Buildmetadata: "build.1234",
			},
			v2: &semver.Version{
				Buildmetadata: "build.1234",
			},
			want: 0,
		},
		{
			testName: "different build metadata only strings",
			v1: &semver.Version{
				Buildmetadata: "f4fe.f234f",
			},
			v2: &semver.Version{
				Buildmetadata: "f4fe.werwe",
			},
			want: 2,
		},
		{
			testName: "lower with build metadata strings and numbers",
			v1: &semver.Version{
				Buildmetadata: "build.1234",
			},
			v2: &semver.Version{
				Buildmetadata: "build.1235",
			},
			want: -1,
		},
		{
			testName: "lower as a subset with v2 longer",
			v1: &semver.Version{
				Buildmetadata: "build.1235",
			},
			v2: &semver.Version{
				Buildmetadata: "build.1235.custom",
			},
			want: -1,
		},
		{
			testName: "lower because v2 is string and v1 is number",
			v1: &semver.Version{
				Buildmetadata: "build.1235",
			},
			v2: &semver.Version{
				Buildmetadata: "build.custom",
			},
			want: -1,
		},
		{
			testName: "greater with build metadata strings and numbers",
			v1: &semver.Version{
				Buildmetadata: "build.2340",
			},
			v2: &semver.Version{
				Buildmetadata: "build.423",
			},
			want: 1,
		},
		{
			testName: "greater as a subset with v1 longer",
			v1: &semver.Version{
				Buildmetadata: "build.1235.custom.v2",
			},
			v2: &semver.Version{
				Buildmetadata: "build.1235.custom",
			},
			want: 1,
		},
		{
			testName: "greater because v1 is string and v2 is number",
			v1: &semver.Version{
				Buildmetadata: "build.v1235",
			},
			v2: &semver.Version{
				Buildmetadata: "build.222",
			},
			want: 1,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			got := tt.v1.CompareBuildMetadata(tt.v2)
			if got != tt.want {
				t.Fatalf("Version.CompareBuildMetadata() got = %v, want %v", got, tt.want)
			}
		})
	}
}
