package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestReadConfig(t *testing.T) {
	_, _, err := ReadCredentials()
	assert.Error(t, err)

	expectedUser := "testuser"
	expectedPassword := "testpass"
	t.Setenv(constants.RegistryUsername, expectedUser)
	t.Setenv(constants.RegistryPassword, expectedPassword)

	username, password, err := ReadCredentials()
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, username)
	assert.Equal(t, expectedPassword, password)
}
