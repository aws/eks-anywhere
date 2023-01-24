package registry_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
	"github.com/aws/eks-anywhere/pkg/registry/mocks"
)

var srcArtifact = registry.Artifact{
	Registry:   "public.ecr.aws",
	Repository: "l0g8r8j6/kube-vip/kube-vip",
	Tag:        "v0.5.5-eks-a-v0.0.0-dev-build.4452",
	Digest:     "sha256:6efe21500abbfbb6b3e37b80dd5dea0b11a0d1b145e84298fee5d7784a77e967",
}

var expectedSrcRef = srcArtifact.VersionedImage()

var desc = ocispec.Descriptor{
	URLs: []string{expectedSrcRef},
}

func TestCopy(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))
	dstClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))

	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	dstClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockDstRepo, nil)
	dstClient.EXPECT().Destination(srcArtifact).Return(expectedSrcRef)
	srcClient.EXPECT().CopyGraph(ctx, &mockSrcRepo, expectedSrcRef, &mockDstRepo, expectedSrcRef).Return(desc, nil)
	dstClient.EXPECT().Tag(ctx, &mockDstRepo, desc, srcArtifact.Tag).Return(nil)

	err := registry.Copy(ctx, srcClient, dstClient, srcArtifact)
	assert.NoError(t, err)
}

func TestCopyTagError(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))
	dstClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))

	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	dstClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockDstRepo, nil)
	dstClient.EXPECT().Destination(srcArtifact).Return(expectedSrcRef)
	srcClient.EXPECT().CopyGraph(ctx, &mockSrcRepo, expectedSrcRef, &mockDstRepo, expectedSrcRef).Return(desc, nil)
	dstClient.EXPECT().Tag(ctx, &mockDstRepo, desc, srcArtifact.Tag).Return(fmt.Errorf("oops"))

	err := registry.Copy(ctx, srcClient, dstClient, srcArtifact)
	assert.EqualError(t, err, "image tag: oops")
}

func TestCopyCopyGraphError(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))
	dstClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))

	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	dstClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockDstRepo, nil)
	dstClient.EXPECT().Destination(srcArtifact).Return(srcArtifact.VersionedImage())
	expectedSrcRef := srcArtifact.VersionedImage()
	srcClient.EXPECT().CopyGraph(ctx, &mockSrcRepo, expectedSrcRef, &mockDstRepo, expectedSrcRef).Return(desc, fmt.Errorf("oops"))

	err := registry.Copy(ctx, srcClient, dstClient, srcArtifact)
	assert.EqualError(t, err, "registry copy: oops")
}

func TestCopyDstGetStorageError(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))
	dstClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))

	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	dstClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockDstRepo, fmt.Errorf("oops"))

	err := registry.Copy(ctx, srcClient, dstClient, srcArtifact)
	assert.EqualError(t, err, "repository destination: oops")
}

func TestCopySrcGetStorageError(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))
	dstClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))

	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, fmt.Errorf("oops"))

	err := registry.Copy(ctx, srcClient, dstClient, srcArtifact)
	assert.EqualError(t, err, "repository source: oops")
}
