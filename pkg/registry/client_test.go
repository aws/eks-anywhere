package registry_test

//
//import (
//	"context"
//	"fmt"
//	"testing"
//
//	"github.com/golang/mock/gomock"
//	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
//	"github.com/stretchr/testify/assert"
//
//	"github.com/aws/eks-anywhere/pkg/registry"
//	"github.com/aws/eks-anywhere/pkg/registry/mocks"
//)
//
//var (
//	ctx   = context.Background()
//	desc  = ocispec.Descriptor{}
//	image = registry.Artifact{
//		Registry:   "public.ecr.aws/",
//		Repository: "eks-anywhere/eks-anywhere-packages",
//		Tag:        "0.2.22-eks-a-24",
//	}
//)
//
//func TestNewOCIRegistry(t *testing.T) {
//	sut := registry.NewOCIRegistry("localhost", "testdata/harbor.eksa.demo.crt", false)
//	assert.Equal(t, "localhost", sut.GetHost())
//	assert.Equal(t, "testdata/harbor.eksa.demo.crt", sut.GetCertFile())
//	assert.False(t, sut.IsInsecure())
//
//	err := sut.Init()
//	assert.NoError(t, err)
//
//	// Does not reinitialize
//	err = sut.Init()
//	assert.NoError(t, err)
//
//	image := registry.Artifact{
//		Registry:   "localhost",
//		Repository: "owner/name",
//		Tag:        "latest",
//	}
//	destination := sut.Destination(image)
//	assert.Equal(t, "localhost/owner/name:latest", destination)
//	sut.SetProject("project/")
//	assert.Equal(t, "project/", sut.GetProject())
//	destination = sut.Destination(image)
//	assert.Equal(t, "localhost/project/owner/name:latest", destination)
//
//	_, err = sut.GetStorage(context.Background(), image)
//	assert.NoError(t, err)
//}
//
//
//func TestOCIRegistry_Copy(t *testing.T) {
//	sut := registry.NewOCIRegistry("public.ecr.aws", "", true)
//	mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockSrcRepo, nil)
//	mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, nil)
//	mockOI.EXPECT().CopyGraph(ctx, &mockSrcRepo, gomock.Any(), desc, gomock.Any()).Return(nil)
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
//	assert.NoError(t, err)
//}
//
//func TestOCIRegistry_CopDryRun(t *testing.T) {
//	sut := registry.NewOCIRegistry("public.ecr.aws", "", true)
//	mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockSrcRepo, nil)
//	mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, nil)
//	sut.OI = mockOI
//	err := sut.Init()
//	assert.NoError(t, err)
//	sut.SetDryRun(true)
//
//	dstRegistry := registry.NewOCIRegistry("localhost", "", false)
//	mockOI = mocks.NewMockOrasInterface(gomock.NewController(t))
//	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))
//	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockDstRepo, nil)
//	dstRegistry.OI = mockOI
//
//	err = sut.Copy(ctx, image, dstRegistry)
//	assert.NoError(t, err)
//}
//
//func TestOCIRegistry_CopyError(t *testing.T) {
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
//func TestOCIRegistry_CopyErrorDestination(t *testing.T) {
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
//func TestOCIRegistry_CopyErrorSource(t *testing.T) {
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
