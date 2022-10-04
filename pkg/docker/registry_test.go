package docker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/docker/mocks"
)

func TestNewRegistryDestination(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	namespace := constants.DefaultRegistryMirrorNamespace
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry, namespace)
	for _, i := range images {
		client.EXPECT().TagImage(test.AContext(), i, registry, namespace)
		client.EXPECT().PushImage(test.AContext(), i, registry, namespace)
	}

	g.Expect(dstLoader.Write(ctx, images...)).To(Succeed())
}

func TestNewRegistryDestinationWhenDigestSpecified(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	namespace := "custom"
	image := "image1@sha256:v1"
	expectedImage := "image1:v1"
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry, namespace)
	client.EXPECT().TagImage(test.AContext(), expectedImage, registry, namespace)
	client.EXPECT().PushImage(test.AContext(), expectedImage, registry, namespace)

	g.Expect(dstLoader.Write(ctx, image)).To(Succeed())
}

func TestNewRegistryDestinationWhenNamespaceNonempty(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	namespace := "custom"
	image := "857151390494.dkr.ecr.us-west-2.amazonaws.com:v1"
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry, namespace)
	client.EXPECT().TagImage(test.AContext(), image, registry, namespace)
	client.EXPECT().PushImage(test.AContext(), image, registry, namespace)

	g.Expect(dstLoader.Write(ctx, image)).To(Succeed())
}

func TestNewRegistryDestinationWhenNamespaceEmpty(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	namespace := ""
	image := "783794618700.dkr.ecr.us-west-2.amazonaws.com:v1"
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry, namespace)
	client.EXPECT().TagImage(test.AContext(), image, registry, namespace)
	client.EXPECT().PushImage(test.AContext(), image, registry, namespace)

	g.Expect(dstLoader.Write(ctx, image)).To(Succeed())
}

func TestNewRegistryDestinationErrorTag(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry, "")
	client.EXPECT().TagImage(test.AContext(), images[0], registry, "").Return(errors.New("error tagging"))
	client.EXPECT().TagImage(test.AContext(), images[1], registry, "").MaxTimes(1)
	client.EXPECT().PushImage(test.AContext(), images[1], registry, "").MaxTimes(1)

	g.Expect(dstLoader.Write(ctx, images...)).To(MatchError(ContainSubstring("error tagging")))
}

func TestNewRegistryDestinationErrorPush(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockImageTaggerPusher(ctrl)

	registry := "https://registry"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry, "")
	client.EXPECT().TagImage(test.AContext(), images[0], registry, "")
	client.EXPECT().PushImage(test.AContext(), images[0], registry, "").Return(errors.New("error pushing"))
	client.EXPECT().TagImage(test.AContext(), images[1], registry, "").MaxTimes(1)
	client.EXPECT().PushImage(test.AContext(), images[1], registry, "").MaxTimes(1)

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
