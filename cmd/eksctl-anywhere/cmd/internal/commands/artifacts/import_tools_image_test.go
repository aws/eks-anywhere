package artifacts_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type importToolsImageTest struct {
	*WithT
	ctx        context.Context
	mover      *mocks.MockImageMover
	unpackager *mocks.MockUnPackager
	command    *artifacts.ImportToolsImage
	bundles    *releasev1.Bundles
}

func newImportToolsImageTest(t *testing.T) *importToolsImageTest {
	downloadFolder := "tmp-folder"
	t.Cleanup(func() {
		os.RemoveAll(downloadFolder)
	})
	ctrl := gomock.NewController(t)
	mover := mocks.NewMockImageMover(ctrl)
	unpackager := mocks.NewMockUnPackager(ctrl)

	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					Eksa: releasev1.EksaBundle{
						CliTools: releasev1.Image{
							URI: "tools:v1.0.0",
						},
					},
				},
			},
		},
	}

	return &importToolsImageTest{
		WithT:      NewWithT(t),
		ctx:        context.Background(),
		mover:      mover,
		unpackager: unpackager,
		command: &artifacts.ImportToolsImage{
			ImageMover:         mover,
			UnPackager:         unpackager,
			TmpArtifactsFolder: downloadFolder,
			Bundles:            bundles,
			InputFile:          "tools_image.tar",
		},
		bundles: bundles,
	}
}

func TestImportToolsImageRun(t *testing.T) {
	tt := newImportToolsImageTest(t)
	tt.unpackager.EXPECT().UnPackage(tt.command.InputFile, tt.command.TmpArtifactsFolder)
	tt.mover.EXPECT().Move(tt.ctx, "tools:v1.0.0")

	tt.Expect(tt.command.Run(tt.ctx)).To(Succeed())
}
