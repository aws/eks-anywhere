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

	registryContext := registry.NewStorageContext("localhost", credentialStore, certificates, false)
	result, err := cache.Get(registryContext)
	assert.NoError(t, err)
	ociRegistry, ok := result.(*registry.OCIRegistryClient)
	assert.True(t, ok)
	assert.Equal(t, "localhost", ociRegistry.GetHost())

	registryContext = registry.NewStorageContext("!@#$", credentialStore, certificates, true)
	result, err = cache.Get(registryContext)
	assert.EqualError(t, err, "error with registry <!@#$>: invalid reference: invalid registry")
	busted, ok := result.(*registry.OCIRegistryClient)
	assert.False(t, ok)
	assert.Nil(t, busted)

	artifact := registry.NewArtifactFromURI("localhost/owner/name:latest")
	assert.Equal(t, "localhost", artifact.Registry)

	cache.Set("localhost", result)
}
