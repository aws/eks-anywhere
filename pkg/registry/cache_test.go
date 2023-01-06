package registry_test

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
)

func TestCache_Get(t *testing.T) {
	cache := registry.NewCache()
	credentialStore := registry.NewCredentialStore()
	certificates := &x509.CertPool{}

	registryContext := registry.NewRegistryContext("localhost", credentialStore, certificates, false)
	result, err := cache.Get(registryContext)
	assert.NoError(t, err)
	ociRegistry, ok := result.(*registry.OCIRegistryClient)
	assert.True(t, ok)
	assert.Equal(t, "localhost", ociRegistry.GetHost())

	registryContext = registry.NewRegistryContext("", credentialStore, certificates, true)
	result, err = cache.Get(registryContext)
	assert.EqualError(t, err, "error with repository example.com: error reading certificate file <bogus.file>: open bogus.file: no such file or directory")
	busted, ok := result.(*registry.OCIRegistryClient)
	assert.False(t, ok)
	assert.Nil(t, busted)
}
