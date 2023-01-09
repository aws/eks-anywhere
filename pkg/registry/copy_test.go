package registry_test

import (
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

func TestCopy(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))
	dstClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))

	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	dstClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockDstRepo, nil)
	var desc ocispec.Descriptor
	expectedImage := srcArtifact.VersionedImage()
	srcClient.EXPECT().Resolve(ctx, &mockSrcRepo, expectedImage).Return(desc, nil)
	srcClient.EXPECT().CopyGraph(ctx, &mockSrcRepo, &mockDstRepo, desc).Return(nil)

	err := registry.Copy(ctx, srcClient, dstClient, srcArtifact)
	assert.NoError(t, err)
}

//func TestCopyError(t *testing.T) {
//	sut := registry.NewOCIRegistry("public.ecr.aws", "", true)
//	mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockSrcRepo, nil)
//	mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, nil)
//	mockOI.EXPECT().CopyGraph(ctx, &mockSrcRepo, gomock.Any(), desc, gomock.Any()).Return(fmt.Errorf("oops"))
//	sut.OI = mockOI
//	err := sut.Init()
//	assert.NoError(t, err)
//
//	dstRegistry := registry.NewOCIRegistry("localhost", "", false)
//	mockOI = mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockDstRepo, nil)
//	dstRegistry.OI = mockOI
//
//	err = sut.Copy(ctx, image, dstRegistry)
//	assert.EqualError(t, err, "registry copy: oops")
//}
//
//func TestCopyErrorDestination(t *testing.T) {
//	sut := registry.NewOCIRegistry("public.ecr.aws", "", true)
//	mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockSrcRepo, nil)
//	mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, nil)
//	sut.OI = mockOI
//	err := sut.Init()
//	assert.NoError(t, err)
//
//	dstRegistry := registry.NewOCIRegistry("localhost", "", false)
//	mockOI = mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockDstRepo, fmt.Errorf("oops"))
//	dstRegistry.OI = mockOI
//
//	err = sut.Copy(ctx, image, dstRegistry)
//	assert.EqualError(t, err, "registry copy destination: error creating repository eks-anywhere/eks-anywhere-packages: oops")
//}
//
//func TestCopyErrorSource(t *testing.T) {
//	sut := registry.NewOCIRegistry("public.ecr.aws", "", true)
//	mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockSrcRepo, nil)
//	mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, fmt.Errorf("oops"))
//	sut.OI = mockOI
//	err := sut.Init()
//	assert.NoError(t, err)
//
//	dstRegistry := registry.NewOCIRegistry("localhost", "", false)
//
//	err = sut.Copy(ctx, image, dstRegistry)
//	assert.EqualError(t, err, "registry copy destination: oops")
//}
//
//func TestOCIRegistry_CopyErrorSourceRepository(t *testing.T) {
//	sut := registry.NewOCIRegistry("public.ecr.aws", "", true)
//	mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockSrcRepo, fmt.Errorf("ooops"))
//	sut.OI = mockOI
//	err := sut.Init()
//	assert.NoError(t, err)
//
//	dstRegistry := registry.NewOCIRegistry("localhost", "", false)
//
//	err = sut.Copy(ctx, image, dstRegistry)
//	assert.EqualError(t, err, "registry copy source: error creating repository eks-anywhere/eks-anywhere-packages: ooops")
//}
