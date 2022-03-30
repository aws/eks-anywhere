package commands

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Reader interface {
	ReadImages(eksaVersion string) ([]releasev1.Image, error)
}

type Mover interface {
	Move(ctx context.Context, images ...string) error
}

type MoveImages struct {
	Reader  Reader
	Version version.Info
	Mover   Mover
}

func (m MoveImages) Run(ctx context.Context) error {
	images, err := m.Reader.ReadImages(m.Version.GitVersion)
	if err != nil {
		return fmt.Errorf("moving images: %v", err)
	}
	taggedImages := make([]string, 0, len(images))
	for _, i := range images {
		taggedImages = append(taggedImages, i.VersionedImage())
	}

	return m.Mover.Move(ctx, taggedImages...)
}
