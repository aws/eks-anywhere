package release

import (
	"testing"

	. "github.com/onsi/gomega"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestTrim(t *testing.T) {
	tests := []struct {
		name     string
		releases EksAReleases
		maxSize  int
		expected EksAReleases
	}{
		{
			name: "maxSize is -1",
			releases: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1"},
				{Version: "v1.2"},
			},
			maxSize: -1,
			expected: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1"},
				{Version: "v1.2"},
			},
		},
		{
			name: "maxSize is greater than the number of releases",
			releases: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1"},
				{Version: "v1.2"},
			},
			maxSize: 5,
			expected: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1"},
				{Version: "v1.2"},
			},
		},
		{
			name: "maxSize is less than the number of releases",
			releases: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1"},
				{Version: "v1.2"},
				{Version: "v1.3"},
				{Version: "v1.4"},
			},
			maxSize: 3,
			expected: EksAReleases{
				{Version: "v1.2"},
				{Version: "v1.3"},
				{Version: "v1.4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			result := Trim(tt.releases, tt.maxSize)
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestAppendOrUpdateRelease(t *testing.T) {
	tests := []struct {
		name         string
		releases     EksAReleases
		releaseToAdd anywherev1alpha1.EksARelease
		expected     EksAReleases
	}{
		{
			name:     "Adding new release to empty releases",
			releases: EksAReleases{},
			releaseToAdd: anywherev1alpha1.EksARelease{
				Version: "v1.0+build.1",
			},
			expected: EksAReleases{
				{Version: "v1.0+build.1"},
			},
		},
		{
			name: "Adding new release to non-empty releases",
			releases: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1"},
			},
			releaseToAdd: anywherev1alpha1.EksARelease{
				Version: "v1.1+build.1",
			},
			expected: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1"},
				{Version: "v1.1+build.1"},
			},
		},
		{
			name: "Updating existing release in releases",
			releases: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1+build.1"},
				{Version: "v1.1+build.2", Number: 2},
				{Version: "v1.1+build.3"},
			},
			releaseToAdd: anywherev1alpha1.EksARelease{
				Version: "v1.1+build.2",
				Number:  3,
			},
			expected: EksAReleases{
				{Version: "v1.0"},
				{Version: "v1.1+build.1"},
				{Version: "v1.1+build.2", Number: 3},
				{Version: "v1.1+build.3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			result := tt.releases.AppendOrUpdateRelease(tt.releaseToAdd)
			g.Expect(result).To(BeComparableTo(tt.expected))
		})
	}
}
