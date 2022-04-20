package releases_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test/mocks"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestReadReleasesFromURL(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	manifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: release-1`

	wantRelease := &releasev1.Release{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
			Kind:       "Release",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "release-1",
		},
	}

	reader.EXPECT().ReadFile(url).Return([]byte(manifest), nil)

	g.Expect(releases.ReadReleasesFromURL(reader, url)).To(Equal(wantRelease))
}

func TestReadReleasesFromURLErrorReading(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	reader.EXPECT().ReadFile(url).Return(nil, errors.New("error reading"))

	_, err := releases.ReadReleasesFromURL(reader, url)
	g.Expect(err).To(MatchError(ContainSubstring("error reading")))
}

func TestReadReleasesFromURLErrorUnmarshaling(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	manifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: {}`

	reader.EXPECT().ReadFile(url).Return([]byte(manifest), nil)

	_, err := releases.ReadReleasesFromURL(reader, url)
	g.Expect(err).To(MatchError(ContainSubstring("failed to unmarshal release manifest from [url]:")))
}

func TestReleaseForVersionSuccess(t *testing.T) {
	tests := []struct {
		name     string
		releases *releasev1.Release
		version  string
		want     *releasev1.EksARelease
	}{
		{
			name: "multiple releases same patch, different prerelease",
			releases: &releasev1.Release{
				Spec: releasev1.ReleaseSpec{
					Releases: []releasev1.EksARelease{
						{Version: "v0.0.1", Number: 1},
						{Version: "v0.0.1-dev", Number: 2},
						{Version: "v0.0.1-alpha", Number: 3},
						{Version: "v0.0.1-beta", Number: 4},
					},
				},
			},
			version: "v0.0.1-alpha",
			want:    &releasev1.EksARelease{Version: "v0.0.1-alpha", Number: 3},
		},
		{
			name: "multiple releases same patch, same prerelease, different build metadata",
			releases: &releasev1.Release{
				Spec: releasev1.ReleaseSpec{
					Releases: []releasev1.EksARelease{
						{Version: "v0.0.1-alpha+werwe", Number: 1},
						{Version: "v0.0.1-alpha+f4fe", Number: 2},
						{Version: "v0.0.1-alpha+f43fs", Number: 3},
						{Version: "v0.0.1-alpha+f234f", Number: 4},
					},
				},
			},
			version: "v0.0.1-alpha",
			want:    &releasev1.EksARelease{Version: "v0.0.1-alpha+werwe", Number: 1},
		},
		{
			name: "version doesn't exist",
			releases: &releasev1.Release{
				Spec: releasev1.ReleaseSpec{
					Releases: []releasev1.EksARelease{
						{Version: "v0.0.1-alpha+werwe", Number: 1},
						{Version: "v0.0.1-alpha+f4fe", Number: 2},
						{Version: "v0.0.1-alpha+f43fs", Number: 3},
						{Version: "v0.0.1-alpha+f234f", Number: 4},
					},
				},
			},
			version: "v0.0.2-alpha",
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(releases.ReleaseForVersion(tt.releases, tt.version)).To(Equal(tt.want))
		})
	}
}

func TestReleaseForVersionError(t *testing.T) {
	tests := []struct {
		name     string
		releases *releasev1.Release
		version  string
		want     string
	}{
		{
			name:    "invalid version",
			version: "x.x.x",
			want:    "invalid eksa version",
		},
		{
			name: "invalid version in releases",
			releases: &releasev1.Release{
				Spec: releasev1.ReleaseSpec{
					Releases: []releasev1.EksARelease{
						{Version: "v0.0.1", Number: 1},
						{Version: "vx.x.x", Number: 2},
					},
				},
			},
			version: "1.1.1",
			want:    "invalid version for release 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			_, err := releases.ReleaseForVersion(tt.releases, tt.version)
			g.Expect(err).To(MatchError(ContainSubstring(tt.want)))
		})
	}
}
