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

type importArtifactsTest struct {
	*WithT
	ctx            context.Context
	reader         *mocks.MockReader
	mover          *mocks.MockImageMover
	importer       *mocks.MockChartImporter
	command        *artifacts.Import
	images, charts []releasev1.Image
	bundles        *releasev1.Bundles
	fileImporter   *mocks.MockFileImporter
}

func newImportArtifactsTest(t *testing.T) *importArtifactsTest {
	downloadFolder := "tmp-folder"
	t.Cleanup(func() {
		os.RemoveAll(downloadFolder)
	})
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	mover := mocks.NewMockImageMover(ctrl)
	importer := mocks.NewMockChartImporter(ctrl)
	fileImporter := mocks.NewMockFileImporter(ctrl)
	images := []releasev1.Image{
		{
			Name: "image 1",
			URI:  "image1:1",
		},
		{
			Name: "image 2",
			URI:  "image2:1",
		},
	}

	charts := []releasev1.Image{
		{
			Name: "chart 1",
			URI:  "chart:v1.0.0",
		},
		{
			Name: "chart 2",
			URI:  "package-chart:v1.0.0",
		},
	}

	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{},
			},
		},
	}

	return &importArtifactsTest{
		WithT:    NewWithT(t),
		ctx:      context.Background(),
		reader:   reader,
		mover:    mover,
		images:   images,
		charts:   charts,
		importer: importer,
		command: &artifacts.Import{
			Reader:             reader,
			ImageMover:         mover,
			ChartImporter:      importer,
			TmpArtifactsFolder: downloadFolder,
			Bundles:            bundles,
			FileImporter:       fileImporter,
		},
		bundles:      bundles,
		fileImporter: fileImporter,
	}
}

func TestImportRun(t *testing.T) {
	tt := newImportArtifactsTest(t)
	tt.reader.EXPECT().ReadImagesFromBundles(tt.ctx, tt.bundles).Return(tt.images, nil)
	tt.mover.EXPECT().Move(tt.ctx, "image1:1", "image2:1")
	tt.reader.EXPECT().ReadChartsFromBundles(tt.ctx, tt.bundles).Return(tt.charts)
	tt.fileImporter.EXPECT().Push(tt.ctx, tt.bundles)
	tt.importer.EXPECT().Import(tt.ctx, "chart:v1.0.0", "package-chart:v1.0.0")

	tt.Expect(tt.command.Run(tt.ctx)).To(Succeed())
}
