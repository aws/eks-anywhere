package docker_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/docker/mocks"
)

func TestNewDiskSource(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockDockerClient(ctrl)

	file := "file"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	sourceLoader := docker.NewDiskSource(client, file)
	client.EXPECT().LoadFromFile(ctx, file)

	g.Expect(sourceLoader.Load(ctx, images...)).To(Succeed())
}

func TestNewDiskDestination(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockDockerClient(ctrl)

	file := "file"
	images := []string{"image1:1", "image2:2"}
	ctx := context.Background()
	dstLoader := docker.NewDiskDestination(client, file)
	client.EXPECT().SaveToFile(ctx, file, images[0], images[1])

	g.Expect(dstLoader.Write(ctx, images...)).To(Succeed())
}
