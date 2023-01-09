package registry_test

import (
	"context"
	"crypto/x509"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
)

var (
	ctx   = context.Background()
	desc  = ocispec.Descriptor{}
	image = registry.Artifact{
		Registry:   "public.ecr.aws/",
		Repository: "eks-anywhere/eks-anywhere-packages",
		Digest:     "sha256:6efe21500abbfbb6b3e37b80dd5dea0b11a0d1b145e84298fee5d7784a77e967",
		Tag:        "0.2.22-eks-a-24",
	}
	credentialStore = registry.NewCredentialStore()
	certificates    = &x509.CertPool{}
)

func TestNewOCIRegistry(t *testing.T) {
	registryContext := registry.NewRegistryContext("localhost", credentialStore, certificates, false)
	sut := registry.NewOCIRegistry(registryContext)
	assert.Equal(t, "localhost", sut.GetHost())
	assert.False(t, sut.IsInsecure())

	err := sut.Init()
	assert.NoError(t, err)

	// Does not reinitialize
	err = sut.Init()
	assert.NoError(t, err)

	image := registry.Artifact{
		Registry:   "localhost",
		Repository: "owner/name",
		Digest:     "sha256:6efe21500abbfbb6b3e37b80dd5dea0b11a0d1b145e84298fee5d7784a77e967",
	}
	destination := sut.Destination(image)
	assert.Equal(t, "localhost/owner/name@sha256:6efe21500abbfbb6b3e37b80dd5dea0b11a0d1b145e84298fee5d7784a77e967", destination)
	sut.SetProject("project/")
	assert.Equal(t, "project/", sut.GetProject())
	destination = sut.Destination(image)
	assert.Equal(t, "localhost/project/owner/name@sha256:6efe21500abbfbb6b3e37b80dd5dea0b11a0d1b145e84298fee5d7784a77e967", destination)

	_, err = sut.GetStorage(context.Background(), image)
	assert.NoError(t, err)
}

func TestOCIRegistry_Copy(t *testing.T) {
	registryContext := registry.NewRegistryContext("public.ecr.aws", credentialStore, certificates, false)
	sut := registry.NewOCIRegistry(registryContext)
	assert.NoError(t, sut.Init())
	//mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
	//mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	//mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockSrcRepo, nil)
	//mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, nil)
	//mockOI.EXPECT().CopyGraph(ctx, &mockSrcRepo, gomock.Any(), desc, gomock.Any()).Return(nil)
	err := sut.Init()
	assert.NoError(t, err)

	registryContext = registry.NewRegistryContext("localhost", credentialStore, certificates, false)
	dstRegistry := registry.NewOCIRegistry(registryContext)
	assert.NoError(t, dstRegistry.Init())
	//mockOI = mocks.NewMockOrasInterface(gomock.NewController(t))
	//mockDstRepo := *mocks.NewMockRepository(gomock.NewController(t))
	//mockOI.EXPECT().Repository(ctx, gomock.Any(), "eks-anywhere/eks-anywhere-packages").Return(&mockDstRepo, nil)
	//dstRegistry.OI = mockOI

	assert.Equal(t, "", image.VersionedImage())
	err = sut.Copy(ctx, image, dstRegistry)
	assert.NoError(t, err)
}

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
