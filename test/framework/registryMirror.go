package framework

import (
	"context"
	"encoding/base64"
	"net"
	"os"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	RegistryEndpointVar = "T_REGISTRY_MIRROR_ENDPOINT"
	RegistryPortVar     = "T_REGISTRY_MIRROR_PORT"
	RegistryUsernameVar = "T_REGISTRY_MIRROR_USERNAME"
	RegistryPasswordVar = "T_REGISTRY_MIRROR_PASSWORD"
	RegistryCACertVar   = "T_REGISTRY_MIRROR_CA_CERT"
)

var registryMirrorRequiredEnvVars = []string{RegistryEndpointVar, RegistryPortVar, RegistryUsernameVar, RegistryPasswordVar, RegistryCACertVar}

func WithRegistryMirrorEndpointAndCert() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, registryMirrorRequiredEnvVars)
		endpoint := os.Getenv(RegistryEndpointVar)
		hostPort := net.JoinHostPort(endpoint, os.Getenv(RegistryPortVar))
		username := os.Getenv(RegistryUsernameVar)
		password := os.Getenv(RegistryPasswordVar)
		err := buildDocker(e.T).Login(context.Background(), hostPort, username, password)
		if err != nil {
			e.T.Fatalf("error logging into docker registry %s: %v", hostPort, err)
		}
		certificate, err := base64.StdEncoding.DecodeString(os.Getenv(RegistryCACertVar))
		if err == nil {
			e.clusterFillers = append(e.clusterFillers,
				api.WithRegistryMirror(endpoint, string(certificate)),
			)
		}
	}
}

func RequiredRegistryMirrorEnvVars() []string {
	return registryMirrorRequiredEnvVars
}
