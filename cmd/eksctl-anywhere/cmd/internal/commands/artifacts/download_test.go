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
	ctx                context.Context
	reader             *mocks.MockReader
	mover              *mocks.MockImageMover
	downloader         *mocks.MockChartDownloader
	toolsDownloader    *mocks.MockImageMover
	packager           *mocks.MockPackager
	command            *artifacts.Download
	images, charts     []releasev1.Image
	bundles            *releasev1.Bundles
	manifestDownloader *mocks.MockManifestDownloader
}

func newDownloadArtifactsTest(t *testing.T) *downloadArtifactsTest {
	downloadFolder := "tmp-folder"
	t.Cleanup(func() {
		os.RemoveAll(downloadFolder)
	})
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	mover := mocks.NewMockImageMover(ctrl)
	toolsDownloader := mocks.NewMockImageMover(ctrl)
	downloader := mocks.NewMockChartDownloader(ctrl)
	packager := mocks.NewMockPackager(ctrl)
	manifestDownloader := mocks.NewMockManifestDownloader(ctrl)
	images := []releasev1.Image{
		{
			Name: "image 1",
			URI:  "image1:1",
		},
		{
			Name: "image 2",
			URI:  "image2:1",
		},
		{
			Name: "tools",
			URI:  "tools:v1.0.0",
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

	return &downloadArtifactsTest{
		WithT:           NewWithT(t),
		ctx:             context.Background(),
		reader:          reader,
		mover:           mover,
		toolsDownloader: toolsDownloader,
		downloader:      downloader,
		packager:        packager,
		images:          images,
		charts:          charts,
		command: &artifacts.Download{
			Reader:                   reader,
			BundlesImagesDownloader:  mover,
			EksaToolsImageDownloader: toolsDownloader,
			ChartDownloader:          downloader,
			Packager:                 packager,
			Version:                  version.Info{GitVersion: "v1.0.0"},
			TmpDowloadFolder:         downloadFolder,
			DstFile:                  "artifacts.tar",
			ManifestDownloader:       manifestDownloader,
		},
		bundles: &releasev1.Bundles{
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
		},
		manifestDownloader: manifestDownloader,
	}
}

func TestDownloadRun(t *testing.T) {
	tt := newDownloadArtifactsTest(t)
	tt.reader.EXPECT().ReadBundlesForVersion("v1.0.0").Return(tt.bundles, nil)
	tt.toolsDownloader.EXPECT().Move(tt.ctx, "tools:v1.0.0")
	tt.reader.EXPECT().ReadImagesFromBundles(tt.ctx, tt.bundles).Return(tt.images, nil)
	tt.mover.EXPECT().Move(tt.ctx, "image1:1", "image2:1")
	tt.reader.EXPECT().ReadChartsFromBundles(tt.ctx, tt.bundles).Return(tt.charts)
	tt.downloader.EXPECT().Download(tt.ctx, "chart:v1.0.0", "package-chart:v1.0.0")
	tt.packager.EXPECT().Package("tmp-folder", "artifacts.tar")
	tt.manifestDownloader.EXPECT().Download(tt.ctx, tt.bundles)

	tt.Expect(tt.command.Run(tt.ctx)).To(Succeed())
}

func TestDownloadErrorReadingImages(t *testing.T) {
	tt := newDownloadArtifactsTest(t)
	tt.reader.EXPECT().ReadBundlesForVersion("v1.0.0").Return(tt.bundles, nil)
	tt.toolsDownloader.EXPECT().Move(tt.ctx, "tools:v1.0.0")
	tt.reader.EXPECT().ReadImagesFromBundles(tt.ctx, tt.bundles).Return(nil, errors.New("error reading images"))

	tt.Expect(tt.command.Run(tt.ctx)).To(MatchError(ContainSubstring("downloading images: error reading images")))
}
