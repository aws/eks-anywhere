package config

import (
	"errors"
	"os"

	"github.com/aws/eks-anywhere/pkg/constants"
)

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
