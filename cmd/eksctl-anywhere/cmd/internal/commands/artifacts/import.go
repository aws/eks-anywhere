package artifacts

import (
	"context"
	"fmt"
	"os"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Import struct {
	Reader             Reader
	Bundles            *releasev1.Bundles
	ImageMover         ImageMover
	ChartImporter      ChartImporter
	TmpArtifactsFolder string
}

type ChartImporter interface {
	Import(ctx context.Context, charts ...string) error
}

func (i Import) Run(ctx context.Context) error {
	images, err := i.Reader.ReadImagesFromBundles(i.Bundles)
	if err != nil {
		return fmt.Errorf("downloading images: %v", err)
	}

	if err = i.ImageMover.Move(ctx, artifactNames(images)...); err != nil {
		return err
	}

	charts := i.Reader.ReadChartsFromBundles(ctx, i.Bundles)

	if err := i.ChartImporter.Import(ctx, artifactNames(charts)...); err != nil {
		return err
	}

	if err := os.RemoveAll(i.TmpArtifactsFolder); err != nil {
		return fmt.Errorf("deleting tmp artifact import folder: %v", err)
	}

	return nil
}
