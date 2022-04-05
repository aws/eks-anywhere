package artifacts_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts/mocks"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type downloadArtifactsTest struct {
	*WithT
	ctx            context.Context
	reader         *mocks.MockReader
	mover          *mocks.MockImageMover
	downloader     *mocks.MockChartDownloader
	packager       *mocks.MockPackager
	command        *artifacts.Download
	images, charts []releasev1.Image
}

func newDownloadArtifactsTest(t *testing.T) *downloadArtifactsTest {
	downloadFolder := "tmp-folder"
	t.Cleanup(func() {
		os.RemoveAll(downloadFolder)
	})
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	mover := mocks.NewMockImageMover(ctrl)
	downloader := mocks.NewMockChartDownloader(ctrl)
	packager := mocks.NewMockPackager(ctrl)
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
			URI:  "chart1:1",
		},
		{
			Name: "image 2",
			URI:  "chart2:1",
		},
	}

	return &downloadArtifactsTest{
		WithT:      NewWithT(t),
		ctx:        context.Background(),
		reader:     reader,
		mover:      mover,
		downloader: downloader,
		packager:   packager,
		images:     images,
		command: &artifacts.Download{
			Reader:           reader,
			ImageMover:       mover,
			ChartDownloader:  downloader,
			Packager:         packager,
			Version:          version.Info{GitVersion: "v1.0.0"},
			TmpDowloadFolder: downloadFolder,
			DstFile:          "artifacts.tar",
		},
		charts: charts,
	}
}

func TestDownloadRun(t *testing.T) {
	tt := newDownloadArtifactsTest(t)
	tt.reader.EXPECT().ReadImages("v1.0.0").Return(tt.images, nil)
	tt.reader.EXPECT().ReadCharts("v1.0.0").Return(tt.charts, nil)
	tt.mover.EXPECT().Move(tt.ctx, "image1:1", "image2:1")
	tt.downloader.EXPECT().Download(tt.ctx, "chart1:1", "chart2:1")
	tt.packager.EXPECT().Package("tmp-folder", "artifacts.tar")

	tt.Expect(tt.command.Run(tt.ctx)).To(Succeed())
}

func TestDownloadErrorReadingImages(t *testing.T) {
	tt := newDownloadArtifactsTest(t)
	tt.reader.EXPECT().ReadImages("v1.0.0").Return(nil, errors.New("error reading images"))

	tt.Expect(tt.command.Run(tt.ctx)).To(MatchError(ContainSubstring("downloading images: error reading images")))
}
