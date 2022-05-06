package curatedpackages_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type defaultRegistryTest struct {
	*WithT
	ctx             context.Context
	releaseManifest *mocks.MockReader
	KubeVersion     string
	CliVersion      version.Info
	Command         *curatedpackages.DefaultRegistry
	bundles         *releasev1.Bundles
}

func newDefaultRegistryTest(t *testing.T) *defaultRegistryTest {
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	kubeVersion := "1.21"
	return &defaultRegistryTest{
		WithT:           NewWithT(t),
		ctx:             context.Background(),
		releaseManifest: reader,
		bundles: &releasev1.Bundles{
			Spec: releasev1.BundlesSpec{
				VersionsBundles: []releasev1.VersionsBundle{
					{
						PackageController: releasev1.PackageBundle{
							Controller: releasev1.Image{
								URI: "test_host/test_env/test_repository:test-version",
							},
						},
						KubeVersion: kubeVersion,
					},
				},
			},
		},
	}
}

func TestDefaultRegistrySucceeds(t *testing.T) {
	tt := newDefaultRegistryTest(t)
	tt.releaseManifest.EXPECT().ReadBundlesForVersion("v1.0.0").Return(tt.bundles, nil)
	tt.Command = curatedpackages.NewDefaultRegistry(
		tt.releaseManifest,
		"1.21",
		version.Info{GitVersion: "v1.0.0"},
	)
	result, err := tt.Command.GetRegistryBaseRef(tt.ctx)
	tt.Expect(err).To(BeNil())
	tt.Expect(result).To(BeEquivalentTo("test_host/test_env/" + curatedpackages.ImageRepositoryName))
}

func TestDefaultRegistryUnknownKubeVersionFails(t *testing.T) {
	tt := newDefaultRegistryTest(t)
	tt.releaseManifest.EXPECT().ReadBundlesForVersion("v1.0.0").Return(tt.bundles, nil)
	tt.Command = curatedpackages.NewDefaultRegistry(
		tt.releaseManifest,
		"1.22",
		version.Info{GitVersion: "v1.0.0"},
	)
	_, err := tt.Command.GetRegistryBaseRef(tt.ctx)
	tt.Expect(err).To(MatchError(ContainSubstring("is not supported by bundles manifest")))
}

func TestDefaultRegistryUnknownGitVersion(t *testing.T) {
	tt := newDefaultRegistryTest(t)
	tt.releaseManifest.EXPECT().ReadBundlesForVersion("v1.0.0").Return(nil, errors.New("unknown git version"))
	tt.Command = curatedpackages.NewDefaultRegistry(
		tt.releaseManifest,
		"1.21",
		version.Info{GitVersion: "v1.0.0"},
	)
	_, err := tt.Command.GetRegistryBaseRef(tt.ctx)
	tt.Expect(err).To(MatchError(ContainSubstring("unable to parse the release manifest")))
}
