package registry

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestNewOCIRegistry(t *testing.T) {
	sut := NewOCIRegistry("localhost", "testdata/harbor.eksa.demo.crt", false)
	assert.Equal(t, "localhost", sut.Host)
	assert.Equal(t, "testdata/harbor.eksa.demo.crt", sut.CertFile)
	assert.False(t, sut.Insecure)

	err := sut.Init()
	assert.NoError(t, err)

	// Does not reinitialize
	err = sut.Init()
	assert.NoError(t, err)

	image := releasev1.Image{
		URI: "localhost/owner/name:latest",
	}
	destination := sut.Destination(image)
	assert.Equal(t, "localhost/owner/name:latest", destination)
	sut.SetProject("project/")
	assert.Equal(t, "project/", sut.Project)
	destination = sut.Destination(image)
	assert.Equal(t, "localhost/project/owner/name:latest", destination)

	_, err = sut.GetStorage(context.Background(), image)
	assert.NoError(t, err)
}

func TestNewOCIRegistryNoCertFile(t *testing.T) {
	sut := NewOCIRegistry("localhost", "", true)
	assert.Equal(t, "localhost", sut.Host)
	assert.Equal(t, "", sut.CertFile)
	assert.True(t, sut.Insecure)

	err := sut.Init()
	assert.NoError(t, err)
}

func TestNewOCIRegistry_InitError(t *testing.T) {
	sut := NewOCIRegistry("localhost", "bogus.crt", false)
	assert.Equal(t, "localhost", sut.Host)
	assert.Equal(t, "bogus.crt", sut.CertFile)
	assert.False(t, sut.Insecure)

	err := sut.Init()
	assert.EqualError(t, err, "error reading certificate file <bogus.crt>: open bogus.crt: no such file or directory")
}

func TestNewOCIRegistryCopy(t *testing.T) {
	ctx := context.Background()
	desc := ocispec.Descriptor{}
	image := releasev1.Image{
		URI: "public.ecr.aws/eks-anywhere/eks-anywhere-packages:0.2.22-eks-a-24",
	}
	sut := NewOCIRegistry("public.ecr.aws", "", true)
	mockOI := mocks.NewMockOrasInterface(gomock.NewController(t))
	mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, nil)
	mockOI.EXPECT().CopyGraph(ctx, gomock.Any(), gomock.Any(), desc, gomock.Any()).Return(nil)
	sut.OI = mockOI

	err := sut.Init()
	assert.NoError(t, err)
	dstRegistry := NewOCIRegistry("localhost", "", false)
	err = dstRegistry.Init()
	assert.NoError(t, err)

	err = sut.Copy(ctx, image, dstRegistry)
	assert.NoError(t, err)

	sut.DryRun = true
	mockOI.EXPECT().Resolve(ctx, gomock.Any(), gomock.Any()).Return(desc, nil)
	err = sut.Copy(ctx, image, dstRegistry)
	assert.NoError(t, err)
}
