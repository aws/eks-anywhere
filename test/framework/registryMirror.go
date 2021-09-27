package framework

import (
	"encoding/base64"
	"os"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	registryEndpointVar = "T_REGISTRY_MIRROR_ENDPOINT"
	registryCACertVar   = "T_REGISTRY_MIRROR_CA_CERT"
)

var registryMirrorRequiredEnvVars = []string{registryEndpointVar}

func WithRegistryMirrorEndpoint() E2ETestOpt {
	return func(e *E2ETest) {
		checkRequiredEnvVars(e.T, registryMirrorRequiredEnvVars)
		registryEndpoint := os.Getenv(registryEndpointVar)

		e.clusterFillers = append(e.clusterFillers,
			api.WithRegistryMirror(registryEndpoint, ""),
		)
	}
}

func WithRegistryMirrorEndpointAndCert() E2ETestOpt {
	return func(e *E2ETest) {
		checkRequiredEnvVars(e.T, append(registryMirrorRequiredEnvVars, registryCACertVar))
		registryEndpoint := os.Getenv(registryEndpointVar)
		registryCACert, err := base64.StdEncoding.DecodeString(os.Getenv(registryCACertVar))
		if err == nil {
			e.clusterFillers = append(e.clusterFillers,
				api.WithRegistryMirror(registryEndpoint, string(registryCACert)),
			)
		}
	}
}
