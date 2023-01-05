package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_Get(t *testing.T) {
	cache := NewCache()

	result, err := cache.Get("localhost", "", false)

	assert.NoError(t, err)
	ociRegistry, ok := result.(*OCIRegistryClient)
	assert.True(t, ok)
	assert.Equal(t, "localhost", ociRegistry.GetHost())

	result, err = cache.Get("localhost", "", false)

	assert.NoError(t, err)
	sameOciRegistry, ok := result.(*OCIRegistryClient)
	assert.True(t, ok)
	assert.Equal(t, "localhost", sameOciRegistry.GetHost())
	assert.Equal(t, "", sameOciRegistry.GetCertFile())
	assert.False(t, sameOciRegistry.IsInsecure())
	assert.Equal(t, ociRegistry, sameOciRegistry)

	result, err = cache.Get("example.com", "bogus.file", true)

	assert.EqualError(t, err, "error with repository example.com: error reading certificate file <bogus.file>: open bogus.file: no such file or directory")
	busted, ok := result.(*OCIRegistryClient)
	assert.False(t, ok)
	assert.Nil(t, busted)
}
