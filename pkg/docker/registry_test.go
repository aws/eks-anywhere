package docker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/docker/mocks"
)

func TestNewRegistryDestination(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	for _, i := range images {
		client.EXPECT().TagImage(test.AContext(), i, registry)
		client.EXPECT().PushImage(test.AContext(), i, registry)
	}

	g.Expect(dstLoader.Write(ctx, images...)).To(Succeed())
}

func TestNewRegistryDestinationErrorTag(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	client.EXPECT().TagImage(test.AContext(), images[0], registry).Return(errors.New("error tagging"))
	client.EXPECT().TagImage(test.AContext(), images[1], registry).MaxTimes(1)
	client.EXPECT().PushImage(test.AContext(), images[1], registry).MaxTimes(1)

	g.Expect(dstLoader.Write(ctx, images...)).To(MatchError(ContainSubstring("error tagging")))
}

func TestNewRegistryDestinationErrorPush(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	client.EXPECT().TagImage(test.AContext(), images[0], registry)
	client.EXPECT().PushImage(test.AContext(), images[0], registry).Return(errors.New("error pushing"))
	client.EXPECT().TagImage(test.AContext(), images[1], registry).MaxTimes(1)
	client.EXPECT().PushImage(test.AContext(), images[1], registry).MaxTimes(1)

	g.Expect(dstLoader.Write(ctx, images...)).To(MatchError(ContainSubstring("error pushing")))
}

func TestNewOriginalRegistrySource(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockDockerClient(ctrl)

	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewOriginalRegistrySource(client)
	for _, i := range images {
		client.EXPECT().PullImage(test.AContext(), i)
	}

	g.Expect(dstLoader.Load(ctx, images...)).To(Succeed())
}

func TestOriginalRegistrySourceError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockDockerClient(ctrl)

	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewOriginalRegistrySource(client)
	client.EXPECT().PullImage(test.AContext(), images[0]).Return(errors.New("error pulling"))
	client.EXPECT().PullImage(test.AContext(), images[1]).MaxTimes(1)

	g.Expect(dstLoader.Load(ctx, images...)).To(MatchError(ContainSubstring("error pulling")))
}
