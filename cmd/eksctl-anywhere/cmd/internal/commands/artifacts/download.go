package artifacts

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Reader interface {
	ReadImages(eksaVersion string) ([]releasev1.Image, error)
	ReadCharts(eksaVersion string) ([]releasev1.Image, error)
	ReadBundlesForVersion(version string) (*releasev1.Bundles, error)
}

type ImageMover interface {
	Move(ctx context.Context, artifacts ...string) error
}

type ChartDownloader interface {
	Download(ctx context.Context, artifacts ...string) error
}

type Packager interface {
	Package(folder string, dstFile string) error
}

type Download struct {
	Reader           Reader
	Version          version.Info
	ImageMover       ImageMover
	ChartDownloader  ChartDownloader
	Packager         Packager
	TmpDowloadFolder string
	DstFile          string
}

func (d Download) Run(ctx context.Context) error {
	if err := os.MkdirAll(d.TmpDowloadFolder, os.ModePerm); err != nil {
		return fmt.Errorf("creating tmp artifact download folder: %v", err)
	}

	images, err := d.Reader.ReadImages(d.Version.GitVersion)
	if err != nil {
		return fmt.Errorf("downloading images: %v", err)
	}

	if err = d.ImageMover.Move(ctx, artifactNames(images)...); err != nil {
		return err
	}

	charts, err := d.Reader.ReadCharts(d.Version.GitVersion)
	if err != nil {
		return fmt.Errorf("downloading charts: %v", err)
	}

	if err := d.ChartDownloader.Download(ctx, artifactNames(charts)...); err != nil {
		return err
	}

	logger.Info("Packaging artifacts", "dst", d.DstFile)
	if err := d.Packager.Package(d.TmpDowloadFolder, d.DstFile); err != nil {
		return err
	}

	if err := os.RemoveAll(d.TmpDowloadFolder); err != nil {
		return fmt.Errorf("deleting tmp artifact download folder: %v", err)
	}

	return nil
}

func artifactNames(artifacts []releasev1.Image) []string {
	taggedArtifacts := make([]string, 0, len(artifacts))
	for _, a := range artifacts {
		taggedArtifacts = append(taggedArtifacts, a.VersionedImage())
	}

	return taggedArtifacts
}
