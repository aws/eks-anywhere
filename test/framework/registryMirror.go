package framework

import (
	"encoding/base64"
	"os"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	RegistryEndpointVar = "T_REGISTRY_MIRROR_ENDPOINT"
	RegistryCACertVar   = "T_REGISTRY_MIRROR_CA_CERT"
)

var registryMirrorRequiredEnvVars = []string{RegistryEndpointVar}

func WithRegistryMirrorEndpoint() E2ETestOpt {
	return func(e *E2ETest) {
		checkRequiredEnvVars(e.T, registryMirrorRequiredEnvVars)
		registryEndpoint := os.Getenv(RegistryEndpointVar)

		e.clusterFillers = append(e.clusterFillers,
			api.WithRegistryMirror(registryEndpoint, ""),
		)
	}
}

func WithRegistryMirrorEndpointAndCert() E2ETestOpt {
	return func(e *E2ETest) {
		checkRequiredEnvVars(e.T, append(registryMirrorRequiredEnvVars, RegistryCACertVar))
		registryEndpoint := os.Getenv(RegistryEndpointVar)
		registryCACert, err := base64.StdEncoding.DecodeString(os.Getenv(RegistryCACertVar))
		if err == nil {
			e.clusterFillers = append(e.clusterFillers,
				api.WithRegistryMirror(registryEndpoint, string(registryCACert)),
			)
		}
	}
}

func RequiredRegistryMirrorEnvVars() []string {
	return registryMirrorRequiredEnvVars
}
