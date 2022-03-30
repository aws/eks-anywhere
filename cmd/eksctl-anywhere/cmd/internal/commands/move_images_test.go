package commands_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands"
	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/mocks"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestMoveImagesRun(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	mover := mocks.NewMockMover(ctrl)

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
	reader.EXPECT().ReadImages("v1.0.0").Return(images, nil)

	mover.EXPECT().Move(ctx, "image1:1", "image2:1")

	c := commands.MoveImages{
		Reader:  reader,
		Mover:   mover,
		Version: version.Info{GitVersion: "v1.0.0"},
	}
	g.Expect(c.Run(ctx)).To(Succeed())
}

func TestMoveImagesRunError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	mover := mocks.NewMockMover(ctrl)

	reader.EXPECT().ReadImages("v1.0.0").Return(nil, errors.New("error reading images"))

	c := commands.MoveImages{
		Reader:  reader,
		Mover:   mover,
		Version: version.Info{GitVersion: "v1.0.0"},
	}
	g.Expect(c.Run(ctx)).To(MatchError(ContainSubstring("moving images: error reading images")))
}
