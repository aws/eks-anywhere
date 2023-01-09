package registry_test

import (
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
)

func TestCopy(t *testing.T) {
	registryContext := registry.NewRegistryContext("public.ecr.aws", credentialStore, certificates, false)
	srcRegistry := registry.NewOCIRegistry(registryContext)
	assert.NoError(t, srcRegistry.Init())
	mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockSrcRepo, nil)
	mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, nil)
	mockOI.EXPECT().CopyGraph(ctx, &mockSrcRepo, gomock.Any(), desc, gomock.Any()).Return(nil)

	registryContext = registry.NewRegistryContext("localhost", credentialStore, certificates, false)
	dstRegistry := registry.NewOCIRegistry(registryContext)
	assert.NoError(t, dstRegistry.Init())
	mockOI = mocks.NewMockOrasInterface(gomock.NewController(t))
	mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))
	mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockDstRepo, nil)
	dstRegistry.OI = mockOI

	err := registry.Copy(ctx, srcRegistry, dstRegistry, image)
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
