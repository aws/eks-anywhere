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
	images := []string{"public.ecr.aws/image1:1", "public.ecr.aws/image2:2"}
	localImages := []string{"https://registry/image1:1", "https://registry/image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	for index, i := range images {
		client.EXPECT().TagImage(test.AContext(), i, localImages[index])
		client.EXPECT().PushImage(test.AContext(), i, localImages[index])
	}

	g.Expect(dstLoader.Write(ctx, images...)).To(Succeed())
}

func TestNewRegistryDestinationWhenDigestSpecified(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	image := "public.ecr.aws/image1@sha256:v1"
	expectedImage := "https://registry/image1:v1"
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	client.EXPECT().TagImage(test.AContext(), image, expectedImage)
	client.EXPECT().PushImage(test.AContext(), image, expectedImage)

	g.Expect(dstLoader.Write(ctx, image)).To(Succeed())
}

func TestNewRegistryDestinationWhenPackagesDevProvided(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	image := "857151390494.dkr.ecr.us-west-2.amazonaws.com/image1:v1"
	expectedImage := "https://registry/l0g8r8j6/image1:v1"
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	client.EXPECT().TagImage(test.AContext(), image, expectedImage)
	client.EXPECT().PushImage(test.AContext(), image, expectedImage)

	g.Expect(dstLoader.Write(ctx, image)).To(Succeed())
}

func TestNewRegistryDestinationWhenPackagesProdProvided(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	image := "783794618700.dkr.ecr.us-west-2.amazonaws.com/image1:v1"
	expectedImage := "https://registry/eks-anywhere/image1:v1"
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	client.EXPECT().TagImage(test.AContext(), image, expectedImage)
	client.EXPECT().PushImage(test.AContext(), image, expectedImage)

	g.Expect(dstLoader.Write(ctx, image)).To(Succeed())
}

func TestNewRegistryDestinationErrorTag(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	images := []string{"public.ecr.aws/image1:1", "public.ecr.aws/image2:2"}
	localImages := []string{"https://registry/image1:1", "https://registry/image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	client.EXPECT().TagImage(test.AContext(), images[0], localImages[0]).Return(errors.New("error tagging"))
	client.EXPECT().TagImage(test.AContext(), images[1], localImages[1]).MaxTimes(1)
	client.EXPECT().PushImage(test.AContext(), images[1], localImages[1]).MaxTimes(1)

	g.Expect(dstLoader.Write(ctx, images...)).To(MatchError(ContainSubstring("error tagging")))
}

func TestNewRegistryDestinationErrorPush(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	images := []string{"public.ecr.aws/image1:1", "public.ecr.aws/image2:2"}
	localImages := []string{"https://registry/image1:1", "https://registry/image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	client.EXPECT().TagImage(test.AContext(), images[0], localImages[0])
	client.EXPECT().PushImage(test.AContext(), images[0], localImages[0]).Return(errors.New("error pushing"))
	client.EXPECT().TagImage(test.AContext(), images[1], localImages[1]).MaxTimes(1)
	client.EXPECT().PushImage(test.AContext(), images[1], localImages[1]).MaxTimes(1)

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
