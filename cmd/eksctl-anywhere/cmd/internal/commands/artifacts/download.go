package artifacts

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Reader interface {
	ReadBundlesForVersion(eksaVersion string) (*releasev1.Bundles, error)
	ReadImagesFromBundles(ctx context.Context, bundles *releasev1.Bundles) ([]releasev1.Image, error)
	ReadChartsFromBundles(ctx context.Context, bundles *releasev1.Bundles) []releasev1.Image
}

type ImageMover interface {
	Move(ctx context.Context, artifacts ...string) error
}

type ChartDownloader interface {
	Download(ctx context.Context, artifacts ...string) error
}

type ManifestDownloader interface {
	Download(ctx context.Context, bundles *releasev1.Bundles)
}

type Packager interface {
	Package(folder string, dstFile string) error
}

type Download struct {
	Reader                   Reader
	FileReader               *files.Reader
	Version                  version.Info
	BundlesImagesDownloader  ImageMover
	EksaToolsImageDownloader ImageMover
	ChartDownloader          ChartDownloader
	Packager                 Packager
	TmpDowloadFolder         string
	DstFile                  string
	ManifestDownloader       ManifestDownloader
	BundlesOverride          string
}

func (d Download) Run(ctx context.Context) error {
	if err := os.MkdirAll(d.TmpDowloadFolder, os.ModePerm); err != nil {
		return fmt.Errorf("creating tmp artifact download folder: %v", err)
	}

	var b *releasev1.Bundles
	var err error
	if d.BundlesOverride != "" {
		b, err = bundles.Read(d.FileReader, d.BundlesOverride)
		if err != nil {
			return fmt.Errorf("reading bundles override: %v", err)
		}
	} else {
		b, err = d.Reader.ReadBundlesForVersion(d.Version.GitVersion)
		if err != nil {
			return fmt.Errorf("reading bundles for version %s: %v", d.Version.GitVersion, err)
		}
	}

	toolsImage := b.DefaultEksAToolsImage().VersionedImage()
	if err = d.EksaToolsImageDownloader.Move(ctx, toolsImage); err != nil {
		return fmt.Errorf("downloading eksa tools image: %v", err)
	}

	images, err := d.Reader.ReadImagesFromBundles(ctx, b)
	if err != nil {
		return fmt.Errorf("downloading images: %v", err)
	}

	if err = d.BundlesImagesDownloader.Move(ctx, removeFromSlice(artifactNames(images), toolsImage)...); err != nil {
		return err
	}

	charts := d.Reader.ReadChartsFromBundles(ctx, b)

	d.ManifestDownloader.Download(ctx, b)

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

func removeFromSlice(s []string, toRemove string) []string {
	index := 0
	for _, i := range s {
		if i != toRemove {
			s[index] = i
			index++
		}
	}
	return s[:index]
}
