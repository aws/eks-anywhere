package kindnetd_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/networking/kindnetd"
)

func TestInstallerInstallErrorGeneratingManifest(t *testing.T) {
	tt := newKindnetdTest(t)
	tt.spec.VersionsBundles["1.19"].Kindnetd.Manifest.URI = "testdata/missing_manifest.yaml"

	tt.Expect(
		tt.k.Installer.Install(tt.ctx, tt.cluster, tt.spec),
	).To(
		MatchError(ContainSubstring("generating kindnetd manifest for install")),
	)
}

func TestInstallerInstallErrorApplyingManifest(t *testing.T) {
	tt := newKindnetdTest(t)
	tt.client.EXPECT().ApplyKubeSpecFromBytes(
		tt.ctx,
		tt.cluster,
		test.MatchFile("testdata/expected_kindnetd_manifest.yaml"),
	).Return(errors.New("generating yaml"))

	tt.Expect(
		tt.k.Installer.Install(tt.ctx, tt.cluster, tt.spec),
	).To(
		MatchError(ContainSubstring("applying kindnetd manifest for install: generating yaml")),
	)
}

func TestInstallerInstallSuccess(t *testing.T) {
	tt := newKindnetdTest(t)
	tt.client.EXPECT().ApplyKubeSpecFromBytes(
		tt.ctx,
		tt.cluster,
		test.MatchFile("testdata/expected_kindnetd_manifest.yaml"),
	)

	tt.Expect(
		tt.k.Installer.Install(tt.ctx, tt.cluster, tt.spec),
	).To(Succeed())
}

func TestInstallForSpecInstallSuccess(t *testing.T) {
	tt := newKindnetdTest(t)
	installerForSpec := kindnetd.NewInstallerForSpec(tt.client, tt.reader, tt.spec)
	tt.client.EXPECT().ApplyKubeSpecFromBytes(
		tt.ctx,
		tt.cluster,
		test.MatchFile("testdata/expected_kindnetd_manifest.yaml"),
	)

	tt.Expect(
		installerForSpec.Install(tt.ctx, tt.cluster),
	).To(Succeed())
}
