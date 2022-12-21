package framework

import (
	"context"
	"encoding/base64"
	"net"
	"os"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	RegistryEndpointVar           = "T_REGISTRY_MIRROR_ENDPOINT"
	RegistryPortVar               = "T_REGISTRY_MIRROR_PORT"
	RegistryUsernameVar           = "T_REGISTRY_MIRROR_USERNAME"
	RegistryPasswordVar           = "T_REGISTRY_MIRROR_PASSWORD"
	RegistryCACertVar             = "T_REGISTRY_MIRROR_CA_CERT"
	RegistryEndpointTinkerbellVar = "T_REGISTRY_MIRROR_ENDPOINT_TINKERBELL"
	RegistryPortTinkerbellVar     = "T_REGISTRY_MIRROR_PORT_TINKERBELL"
	RegistryUsernameTinkerbellVar = "T_REGISTRY_MIRROR_USERNAME_TINKERBELL"
	RegistryPasswordTinkerbellVar = "T_REGISTRY_MIRROR_PASSWORD_TINKERBELL"
	RegistryCACertTinkerbellVar   = "T_REGISTRY_MIRROR_CA_CERT_TINKERBELL"
)

var (
	registryMirrorRequiredEnvVars           = []string{RegistryEndpointVar, RegistryPortVar, RegistryUsernameVar, RegistryPasswordVar, RegistryCACertVar}
	registryMirrorTinkerbellRequiredEnvVars = []string{RegistryEndpointTinkerbellVar, RegistryPortTinkerbellVar, RegistryUsernameTinkerbellVar, RegistryPasswordTinkerbellVar, RegistryCACertTinkerbellVar}
)

func WithRegistryMirrorEndpointAndCert(providerName string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		var endpoint, hostPort, username, password, registryCert string
		port := "443"

		switch providerName {
		case constants.TinkerbellProviderName:
			checkRequiredEnvVars(e.T, registryMirrorTinkerbellRequiredEnvVars)
			endpoint = os.Getenv(RegistryEndpointTinkerbellVar)
			hostPort = net.JoinHostPort(endpoint, os.Getenv(RegistryPortTinkerbellVar))
			username = os.Getenv(RegistryUsernameTinkerbellVar)
			password = os.Getenv(RegistryPasswordTinkerbellVar)
			registryCert = os.Getenv(RegistryCACertTinkerbellVar)
			if os.Getenv(RegistryPortTinkerbellVar) != "" {
				port = os.Getenv(RegistryPortTinkerbellVar)
			}
		default:
			checkRequiredEnvVars(e.T, registryMirrorRequiredEnvVars)
			endpoint = os.Getenv(RegistryEndpointVar)
			hostPort = net.JoinHostPort(endpoint, os.Getenv(RegistryPortVar))
			username = os.Getenv(RegistryUsernameVar)
			password = os.Getenv(RegistryPasswordVar)
			registryCert = os.Getenv(RegistryCACertVar)
			if os.Getenv(RegistryPortVar) != "" {
				port = os.Getenv(RegistryPortVar)
			}
		}

		err := buildDocker(e.T).Login(context.Background(), hostPort, username, password)
		if err != nil {
			e.T.Fatalf("error logging into docker registry %s: %v", hostPort, err)
		}
		certificate, err := base64.StdEncoding.DecodeString(registryCert)
		if err == nil {
			e.clusterFillers = append(e.clusterFillers,
				api.WithRegistryMirror(endpoint, port, string(certificate)),
			)
		}
		// Set env vars for helm login/push
		err = os.Setenv("REGISTRY_USERNAME", username)
		if err != nil {
			e.T.Fatalf("unable to set REGISTRY_USERNAME: %v", err)
		}
		err = os.Setenv("REGISTRY_PASSWORD", password)
		if err != nil {
			e.T.Fatalf("unable to set REGISTRY_PASSWORD: %v", err)
		}
	}
}

func RequiredRegistryMirrorEnvVars() []string {
	return append(registryMirrorRequiredEnvVars, registryMirrorTinkerbellRequiredEnvVars...)
}
