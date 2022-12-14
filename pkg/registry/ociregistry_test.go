package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestNewOCIRegistry(t *testing.T) {
	registry := NewOCIRegistry("localhost", "testdata/harbor.eksa.demo.crt", false)
	assert.Equal(t, "localhost", registry.Host)
	assert.Equal(t, "testdata/harbor.eksa.demo.crt", registry.CertFile)
	assert.False(t, registry.Insecure)

	err := registry.Init()
	assert.NoError(t, err)

	err = registry.Init()
	assert.NoError(t, err)

	image := releasev1.Image{
		URI: "localhost/owner/name:latest",
	}
	destination := registry.Destination(image)
	assert.Equal(t, "localhost/owner/name:latest", destination)
	registry.SetProject("project/")
	assert.Equal(t, "project/", registry.Project)
	destination = registry.Destination(image)
	assert.Equal(t, "localhost/project/owner/name:latest", destination)

	_, err = registry.GetStorage(context.Background(), image)
	assert.NoError(t, err)
}

func TestNewOCIRegistryNoCertFile(t *testing.T) {
	registry := NewOCIRegistry("localhost", "", true)
	assert.Equal(t, "localhost", registry.Host)
	assert.Equal(t, "", registry.CertFile)
	assert.True(t, registry.Insecure)

	err := registry.Init()
	assert.NoError(t, err)
}

func TestNewOCIRegistry_InitErrro(t *testing.T) {
	registry := NewOCIRegistry("localhost", "bogus.crt", false)
	assert.Equal(t, "localhost", registry.Host)
	assert.Equal(t, "bogus.crt", registry.CertFile)
	assert.False(t, registry.Insecure)

	err := registry.Init()
	assert.EqualError(t, err, "error reading certificate file <bogus.crt>: open bogus.crt: no such file or directory")
}

func TestNewOCIRegistrCopy(t *testing.T) {
	image := releasev1.Image{
		URI: "public.ecr.aws/eks-anywhere/eks-anywhere-packages:0.2.22-eks-a-24",
	}
	srcRegistry := NewOCIRegistry("public.ecr.aws", "", true)
	err := srcRegistry.Init()
	assert.NoError(t, err)
	dstRegistry := NewOCIRegistry("localhost", "", false)
	err = dstRegistry.Init()
	assert.NoError(t, err)
	srcRegistry.DryRun = true

	err = srcRegistry.Copy(context.Background(), image, dstRegistry)
	assert.NoError(t, err)
}
