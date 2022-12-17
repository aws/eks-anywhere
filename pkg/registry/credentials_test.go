package registry

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialStore_Init(t *testing.T) {
	credentialStore := NewCredentialStore("testdata")

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
	assert.EqualError(t, err, "error getting credentials - err: exec: \"docker-credential-bogus\": executable file not found in $PATH, out: ``")
	assert.Equal(t, "", result.Username)
	assert.Equal(t, "", result.Password)
	assert.Equal(t, "", result.AccessToken)
	assert.Equal(t, "", result.RefreshToken)
}

func TestCredentialStore_InitEmpty(t *testing.T) {
	credentialStore := NewCredentialStore("testdata/empty")
	err := credentialStore.Init()
	assert.NoError(t, err)
}

func TestCredentialStore_InitNoPermissions(t *testing.T) {
	dir, err := ioutil.TempDir("testdata", "noperms")
	defer os.Remove(dir)
	assert.NoError(t, err)
	fileName := dir + "/config.json"
	err = os.WriteFile(fileName, []byte("{}"), 0000)
	defer os.Remove(fileName)
	assert.NoError(t, err)

	credentialStore := NewCredentialStore(dir)
	err = credentialStore.Init()
	assert.EqualError(t, err, dir+"/config.json: open "+dir+"/config.json: permission denied")
}
