package config

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/constants"
)

const registryAuthSecretName = "registry-credentials"

func ReadCredentials() (username, password string, err error) {
	username, ok := os.LookupEnv(constants.RegistryUsername)
	if !ok {
		return "", "", errors.New("please set REGISTRY_USERNAME env var")
	}

	password, ok = os.LookupEnv(constants.RegistryPassword)
	if !ok {
		return "", "", errors.New("please set REGISTRY_PASSWORD env var")
	}

	return username, password, nil
}

// ReadCredentialsFromSecret reads from Kubernetes secret registry-credentials.
// Returns the username and password, or error.
func ReadCredentialsFromSecret(ctx context.Context, client client.Client) (username, password string, err error) {
	registryAuthSecret := &corev1.Secret{}
	key := types.NamespacedName{Name: registryAuthSecretName, Namespace: constants.EksaSystemNamespace}
	if err := client.Get(ctx, key, registryAuthSecret); err != nil {
		return "", "", errors.Wrap(err, "fetching registry auth secret")
	}

	rUsername := registryAuthSecret.Data["username"]
	rPassword := registryAuthSecret.Data["password"]

	return string(rUsername), string(rPassword), nil
}

// SetCredentialsEnv sets the registry username and password env variables.
func SetCredentialsEnv(username, password string) error {
	if err := os.Setenv(constants.RegistryUsername, username); err != nil {
		return fmt.Errorf("failed setting env %s: %v", constants.RegistryUsername, err)
	}

	if err := os.Setenv(constants.RegistryPassword, password); err != nil {
		return fmt.Errorf("failed setting env %s: %v", constants.RegistryPassword, err)
	}

	return nil
}
