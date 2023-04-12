package config

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestReadConfig(t *testing.T) {
	os.Unsetenv(constants.RegistryUsername)
	os.Unsetenv(constants.RegistryPassword)
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

func TestSetCredentialsEnv(t *testing.T) {
	uName := ""
	uPass := ""
	err := SetCredentialsEnv(uName, uPass)
	assert.NoError(t, err)
}

func TestReadCredentialsFromSecret(t *testing.T) {
	ctx := context.Background()
	expectedUser := "testuser"
	expectedPassword := "testpass"
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryAuthSecretName,
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string][]byte{
			"username": []byte(expectedUser),
			"password": []byte(expectedPassword),
		},
	}

	objs := []runtime.Object{sec}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()
	u, p, err := ReadCredentialsFromSecret(ctx, cl)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, u)
	assert.Equal(t, expectedPassword, p)
}

func TestReadCredentialsFromSecretNotFound(t *testing.T) {
	ctx := context.Background()
	cb := fake.NewClientBuilder()
	cl := cb.Build()
	u, p, err := ReadCredentialsFromSecret(ctx, cl)
	assert.ErrorContains(t, err, "fetching registry auth secret")
	assert.Empty(t, u)
	assert.Empty(t, p)
}
