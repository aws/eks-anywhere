package artifacts

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/pkg/logger"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ImportToolsImage struct {
	Bundles            *releasev1.Bundles
	ImageMover         ImageMover
	UnPackager         UnPackager
	InputFile          string
	TmpArtifactsFolder string
}

type UnPackager interface {
	UnPackage(orgFile, dstFolder string) error
}

func (i ImportToolsImage) Run(ctx context.Context) error {
	if err := os.MkdirAll(i.TmpArtifactsFolder, os.ModePerm); err != nil {
		return fmt.Errorf("creating tmp artifact folder to unpackage tools image: %v", err)
	}

	logger.Info("Unpackaging artifacts", "dst", i.TmpArtifactsFolder)
	if err := i.UnPackager.UnPackage(i.InputFile, i.TmpArtifactsFolder); err != nil {
		return err
	}

	toolsImage := i.Bundles.DefaultEksAToolsImage().VersionedImage()

	if err := i.ImageMover.Move(ctx, toolsImage); err != nil {
		return fmt.Errorf("importing tools image: %v", err)
	}

	return nil
}
