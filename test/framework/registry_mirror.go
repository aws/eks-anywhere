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
	RegistryEndpointVar                  = "T_REGISTRY_MIRROR_ENDPOINT"
	RegistryPortVar                      = "T_REGISTRY_MIRROR_PORT"
	RegistryUsernameVar                  = "T_REGISTRY_MIRROR_USERNAME"
	RegistryPasswordVar                  = "T_REGISTRY_MIRROR_PASSWORD"
	RegistryCACertVar                    = "T_REGISTRY_MIRROR_CA_CERT"
	RegistryEndpointTinkerbellVar        = "T_REGISTRY_MIRROR_ENDPOINT_TINKERBELL"
	RegistryPortTinkerbellVar            = "T_REGISTRY_MIRROR_PORT_TINKERBELL"
	RegistryUsernameTinkerbellVar        = "T_REGISTRY_MIRROR_USERNAME_TINKERBELL"
	RegistryPasswordTinkerbellVar        = "T_REGISTRY_MIRROR_PASSWORD_TINKERBELL"
	RegistryCACertTinkerbellVar          = "T_REGISTRY_MIRROR_CA_CERT_TINKERBELL"
	RegistryMirrorDefaultSecurityGroup   = "T_REGISTRY_MIRROR_DEFAULT_SECURITY_GROUP"
	RegistryMirrorAirgappedSecurityGroup = "T_REGISTRY_MIRROR_AIRGAPPED_SECURITY_GROUP"
	PrivateRegistryEndpointVar           = "T_PRIVATE_REGISTRY_MIRROR_ENDPOINT"
	PrivateRegistryPortVar               = "T_PRIVATE_REGISTRY_MIRROR_PORT"
	PrivateRegistryUsernameVar           = "T_PRIVATE_REGISTRY_MIRROR_USERNAME"
	PrivateRegistryPasswordVar           = "T_PRIVATE_REGISTRY_MIRROR_PASSWORD"
	PrivateRegistryCACertVar             = "T_PRIVATE_REGISTRY_MIRROR_CA_CERT"
	PrivateRegistryEndpointTinkerbellVar = "T_PRIVATE_REGISTRY_MIRROR_ENDPOINT_TINKERBELL"
	PrivateRegistryPortTinkerbellVar     = "T_PRIVATE_REGISTRY_MIRROR_PORT_TINKERBELL"
	PrivateRegistryUsernameTinkerbellVar = "T_PRIVATE_REGISTRY_MIRROR_USERNAME_TINKERBELL"
	PrivateRegistryPasswordTinkerbellVar = "T_PRIVATE_REGISTRY_MIRROR_PASSWORD_TINKERBELL"
	PrivateRegistryCACertTinkerbellVar   = "T_PRIVATE_REGISTRY_MIRROR_CA_CERT_TINKERBELL"
)

var (
	registryMirrorRequiredEnvVars                  = []string{RegistryEndpointVar, RegistryPortVar, RegistryUsernameVar, RegistryPasswordVar, RegistryCACertVar}
	registryMirrorTinkerbellRequiredEnvVars        = []string{RegistryEndpointTinkerbellVar, RegistryPortTinkerbellVar, RegistryUsernameTinkerbellVar, RegistryPasswordTinkerbellVar, RegistryCACertTinkerbellVar}
	registryMirrorDockerAirgappedRequiredEnvVars   = []string{RegistryMirrorDefaultSecurityGroup, RegistryMirrorAirgappedSecurityGroup}
	privateRegistryMirrorRequiredEnvVars           = []string{PrivateRegistryEndpointVar, PrivateRegistryPortVar, PrivateRegistryUsernameVar, PrivateRegistryPasswordVar, PrivateRegistryCACertVar}
	privateRegistryMirrorTinkerbellRequiredEnvVars = []string{PrivateRegistryEndpointTinkerbellVar, PrivateRegistryPortTinkerbellVar, PrivateRegistryUsernameTinkerbellVar, PrivateRegistryPasswordTinkerbellVar, PrivateRegistryCACertTinkerbellVar}
)

// WithRegistryMirrorInsecureSkipVerify sets up e2e for registry mirrors with InsecureSkipVerify option.
func WithRegistryMirrorInsecureSkipVerify(providerName string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		setupRegistryMirrorEndpointAndCert(e, providerName, true)
	}
}

// WithRegistryMirrorEndpointAndCert sets up e2e for registry mirrors.
func WithRegistryMirrorEndpointAndCert(providerName string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		setupRegistryMirrorEndpointAndCert(e, providerName, false)
	}
}

// WithAuthenticatedRegistryMirror sets up e2e for authenticated registry mirrors.
func WithAuthenticatedRegistryMirror(providerName string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		var endpoint, hostPort, username, password, registryCert string
		port := "443"

		switch providerName {
		case constants.TinkerbellProviderName:
			checkRequiredEnvVars(e.T, privateRegistryMirrorTinkerbellRequiredEnvVars)
			endpoint = os.Getenv(PrivateRegistryEndpointTinkerbellVar)
			hostPort = net.JoinHostPort(endpoint, os.Getenv(PrivateRegistryPortTinkerbellVar))
			username = os.Getenv(PrivateRegistryUsernameTinkerbellVar)
			password = os.Getenv(PrivateRegistryPasswordTinkerbellVar)
			registryCert = os.Getenv(PrivateRegistryCACertTinkerbellVar)
			if os.Getenv(PrivateRegistryPortTinkerbellVar) != "" {
				port = os.Getenv(PrivateRegistryPortTinkerbellVar)
			}
		default:
			checkRequiredEnvVars(e.T, privateRegistryMirrorRequiredEnvVars)
			endpoint = os.Getenv(PrivateRegistryEndpointVar)
			hostPort = net.JoinHostPort(endpoint, os.Getenv(PrivateRegistryPortVar))
			username = os.Getenv(PrivateRegistryUsernameVar)
			password = os.Getenv(PrivateRegistryPasswordVar)
			registryCert = os.Getenv(PrivateRegistryCACertVar)
			if os.Getenv(PrivateRegistryPortVar) != "" {
				port = os.Getenv(PrivateRegistryPortVar)
			}
		}

		// Set env vars for helm login/push
		err := os.Setenv("REGISTRY_USERNAME", username)
		if err != nil {
			e.T.Fatalf("unable to set REGISTRY_USERNAME: %v", err)
		}
		err = os.Setenv("REGISTRY_PASSWORD", password)
		if err != nil {
			e.T.Fatalf("unable to set REGISTRY_PASSWORD: %v", err)
		}

		err = buildDocker(e.T).Login(context.Background(), hostPort, username, password)
		if err != nil {
			e.T.Fatalf("error logging into docker registry %s: %v", hostPort, err)
		}
		certificate, err := base64.StdEncoding.DecodeString(registryCert)
		if err == nil {
			e.clusterFillers = append(e.clusterFillers,
				api.WithRegistryMirror(endpoint, port, string(certificate), true, false),
			)
		}
	}
}

func RequiredRegistryMirrorEnvVars() []string {
	registryMirrorRequiredEnvVars = append(registryMirrorRequiredEnvVars, registryMirrorTinkerbellRequiredEnvVars...)
	registryMirrorRequiredEnvVars = append(registryMirrorRequiredEnvVars, privateRegistryMirrorRequiredEnvVars...)
	registryMirrorRequiredEnvVars = append(registryMirrorRequiredEnvVars, privateRegistryMirrorTinkerbellRequiredEnvVars...)
	return append(registryMirrorRequiredEnvVars, registryMirrorDockerAirgappedRequiredEnvVars...)
}

func setupRegistryMirrorEndpointAndCert(e *ClusterE2ETest, providerName string, insecureSkipVerify bool) {
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
			api.WithRegistryMirror(endpoint, port, string(certificate), false, insecureSkipVerify),
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

// SetRegistryMirrorDefaultInstanceSecurityGroupOnCleanup sets the instance security group to the registry mirror default security group on cleanup.
func (e *ClusterE2ETest) SetRegistryMirrorDefaultInstanceSecurityGroupOnCleanup(opts ...CommandOpt) {
	e.T.Cleanup(func() {
		e.ChangeInstanceSecurityGroup(os.Getenv(RegistryMirrorDefaultSecurityGroup))
	})
}
