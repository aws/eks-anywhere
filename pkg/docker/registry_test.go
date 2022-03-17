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

func TestNewRegistryDestination(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockDockerClient(ctrl)

	registry := "https://registry"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	for _, i := range images {
		client.EXPECT().PushImage(ctx, i, registry)
	}

	g.Expect(dstLoader.Write(ctx, images...)).To(Succeed())
}

func TestNewRegistryDestinationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockDockerClient(ctrl)

	registry := "https://registry"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewRegistryDestination(client, registry)
	client.EXPECT().PushImage(ctx, images[0], registry).Return(errors.New("error pushing"))

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
		client.EXPECT().PullImage(ctx, i)
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
	client.EXPECT().PullImage(ctx, images[0]).Return(errors.New("error pulling"))

	g.Expect(dstLoader.Load(ctx, images...)).To(MatchError(ContainSubstring("error pulling")))
}
