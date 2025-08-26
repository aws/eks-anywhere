package artifacts

import (
	"context"
	"fmt"
	"os"
	"strings"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Import struct {
	Reader             Reader
	Bundles            *releasev1.Bundles
	ImageMover         ImageMover
	ChartImporter      ChartImporter
	TmpArtifactsFolder string
	FileImporter       FileImporter
}

type ChartImporter interface {
	Import(ctx context.Context, charts ...string) error
}

type FileImporter interface {
	Push(ctx context.Context, bundles *releasev1.Bundles)
}

func (i Import) Run(ctx context.Context) error {
	images, err := i.Reader.ReadImagesFromBundles(ctx, i.Bundles)
	if err != nil {
		return fmt.Errorf("downloading images: %v", err)
	}

	// Filter out CSI component images as they're not used by EKS Anywhere
	// but are still referenced in the EKS-D release manifest
	var filteredImages []releasev1.Image
	for _, img := range images {
		if img.URI != "" && strings.Contains(img.URI, "public.ecr.aws/csi-components/") {
			// Skip CSI component images
			continue
		}
		filteredImages = append(filteredImages, img)
	}

	if err = i.ImageMover.Move(ctx, artifactNames(filteredImages)...); err != nil {
		return err
	}

	charts := i.Reader.ReadChartsFromBundles(ctx, i.Bundles)

	if err := i.ChartImporter.Import(ctx, artifactNames(charts)...); err != nil {
		return err
	}

	i.FileImporter.Push(ctx, i.Bundles)

	if err := os.RemoveAll(i.TmpArtifactsFolder); err != nil {
		return fmt.Errorf("deleting tmp artifact import folder: %v", err)
	}

	return nil
}
