package registry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
)

func TestCredentialStore_Init(t *testing.T) {
	credentialStore := registry.NewCredentialStore()
	credentialStore.SetDirectory("testdata")

	err := credentialStore.Init()
	assert.NoError(t, err)

	result, err := credentialStore.Credential("localhost")
	assert.NoError(t, err)
	assert.Equal(t, "user", result.Username)
	assert.Equal(t, "pass", result.Password)
	assert.Equal(t, "", result.AccessToken)
	assert.Equal(t, "", result.RefreshToken)

	result, err = credentialStore.Credential("harbor.eksa.demo:30003")
	assert.NoError(t, err)
	assert.Equal(t, "captain", result.Username)
	assert.Equal(t, "haddock", result.Password)
	assert.Equal(t, "", result.AccessToken)
	assert.Equal(t, "", result.RefreshToken)

	result, err = credentialStore.Credential("bogus")
	assert.NoError(t, err)
	assert.Equal(t, "", result.Username)
	assert.Equal(t, "", result.Password)
	assert.Equal(t, "", result.AccessToken)
	assert.Equal(t, "", result.RefreshToken)

	result, err = credentialStore.Credential("5551212.dkr.ecr.us-west-2.amazonaws.com")
	// This is a generic error, so using errors.Is won't work, and this is as
	// much of the string as we can reliably match against in a cross-platform
	// fashion. Until they change it, then everything will break.
	assert.ErrorContains(t, err, "error getting credentials - err")
	assert.Equal(t, "", result.Username)
	assert.Equal(t, "", result.Password)
	assert.Equal(t, "", result.AccessToken)
	assert.Equal(t, "", result.RefreshToken)
}

func TestCredentialStore_InitEmpty(t *testing.T) {
	credentialStore := registry.NewCredentialStore()
	credentialStore.SetDirectory("testdata/empty")
	err := credentialStore.Init()
	assert.NoError(t, err)
}
