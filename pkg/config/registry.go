package config

import (
	"errors"
	"os"
)

func ReadCredentials() (username, password string, err error) {
	username, ok := os.LookupEnv("REGISTRY_USERNAME")
	if !ok {
		return "", "", errors.New("please set REGISTRY_USERNAME env var")
	}

	password, ok = os.LookupEnv("REGISTRY_PASSWORD")
	if !ok {
		return "", "", errors.New("please set REGISTRY_PASSWORD env var")
	}

	return username, password, nil
}
