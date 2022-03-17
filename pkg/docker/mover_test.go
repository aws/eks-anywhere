package docker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/docker/mocks"
)

type moverTest struct {
	*WithT
	ctx    context.Context
	src    *mocks.MockImageSource
	dst    *mocks.MockImageDestination
	images []string
}

func newMoverTest(t *testing.T) *moverTest {
	ctrl := gomock.NewController(t)

	return &moverTest{
		WithT:  NewWithT(t),
		ctx:    context.Background(),
		src:    mocks.NewMockImageSource(ctrl),
		dst:    mocks.NewMockImageDestination(ctrl),
		images: []string{"image1:1", "image2:2"},
	}
}

func TestImageMoverMove(t *testing.T) {
	tt := newMoverTest(t)
	tt.src.EXPECT().Load(tt.ctx, tt.images[0], tt.images[1])
	tt.dst.EXPECT().Write(tt.ctx, tt.images[0], tt.images[1])

	m := docker.NewImageMover(tt.src, tt.dst)
	tt.Expect(m.Move(tt.ctx, tt.images...)).To(Succeed())
}

func TestImageMoverMoveErrorSource(t *testing.T) {
	tt := newMoverTest(t)
	errorMsg := "fake error"
	tt.src.EXPECT().Load(tt.ctx, tt.images[0], tt.images[1]).Return(errors.New(errorMsg))

	m := docker.NewImageMover(tt.src, tt.dst)

	tt.Expect(m.Move(tt.ctx, tt.images...)).To(MatchError("loading docker image mover source: fake error"))
}

func TestImageMoverMoveErrorDestination(t *testing.T) {
	tt := newMoverTest(t)
	errorMsg := "fake error"
	tt.src.EXPECT().Load(tt.ctx, tt.images[0], tt.images[1])
	tt.dst.EXPECT().Write(tt.ctx, tt.images[0], tt.images[1]).Return(errors.New(errorMsg))

	m := docker.NewImageMover(tt.src, tt.dst)

	tt.Expect(m.Move(tt.ctx, tt.images...)).To(MatchError("writing images to destination with image mover: fake error"))
}
