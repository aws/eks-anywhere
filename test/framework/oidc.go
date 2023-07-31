package framework

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/oidc"
	"github.com/aws/eks-anywhere/pkg/executables"
)

const (
	OIDCIssuerUrlVar = "T_OIDC_ISSUER_URL"
	OIDCClientIdVar  = "T_OIDC_CLIENT_ID"
	OIDCKidVar       = "T_OIDC_KID"
	OIDCKeyFileVar   = "T_OIDC_KEY_FILE"
)

var oidcRequiredEnvVars = []string{
	OIDCIssuerUrlVar,
	OIDCClientIdVar,
	OIDCKidVar,
	OIDCKeyFileVar,
}

func WithOIDC() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.addClusterConfigFillers(WithOIDCClusterConfig(e.T))
	}
}

// WithOIDCClusterConfig returns a ClusterConfigFiller that adds the default
// OIDCConfig for E2E tests to the cluster Config and links it by name in the
// Cluster resource.
func WithOIDCClusterConfig(t T) api.ClusterConfigFiller {
	checkRequiredEnvVars(t, oidcRequiredEnvVars)
	name := defaultClusterName

	return api.JoinClusterConfigFillers(
		api.WithOIDCConfig(name,
			api.WithOIDCRequiredClaims("kubernetesAccess", "true"),
			api.WithOIDCGroupsPrefix("s3-oidc:"),
			api.WithOIDCGroupsClaim("groups"),
			api.WithOIDCUsernamePrefix("s3-oidc:"),
			api.WithOIDCUsernameClaim("email"),
			api.WithStringFromEnvVarOIDCConfig(OIDCIssuerUrlVar, api.WithOIDCIssuerUrl),
			api.WithStringFromEnvVarOIDCConfig(OIDCClientIdVar, api.WithOIDCClientId),
		),
		api.ClusterToConfigFiller(
			api.WithOIDCIdentityProviderRef(name),
		),
	)
}

func (e *ClusterE2ETest) ValidateOIDC() {
	ctx := context.Background()
	cluster := e.Cluster()
	e.T.Log("Creating roles for OIDC")
	err := e.KubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, oidcRoles)
	if err != nil {
		e.T.Errorf("Error applying roles for oids: %v", err)
		return
	}

	issuerUrl, err := url.Parse(os.Getenv(OIDCIssuerUrlVar))
	if err != nil {
		e.T.Errorf("Error parsing oidc issuer url: %v", err)
		return
	}

	kid := os.Getenv(OIDCKidVar)
	keyFile := os.Getenv(OIDCKeyFileVar)

	e.T.Log("Generating OIDC JWT token")
	jwt, err := oidc.NewJWT(
		path.Join(issuerUrl.Host, issuerUrl.Path),
		kid,
		keyFile,
		oidc.WithEmail("oidcuser@aws.com"),
		oidc.WithGroup("developers"),
		oidc.WithRole("dev"),
		oidc.WithKubernetesAccess(true),
		oidc.WithAudience(kid),
	)
	if err != nil {
		e.T.Errorf("Error generating JWT token for oidc: %v", err)
		return
	}

	apiServerUrl, err := e.KubectlClient.GetApiServerUrl(ctx, cluster)
	if err != nil {
		e.T.Errorf("Error getting api server url: %v", err)
		return
	}

	e.T.Log("Getting pods with OIDC token")
	_, err = e.KubectlClient.GetPods(
		ctx,
		executables.WithKubeconfig(filepath.Join(e.ClusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", e.ClusterName))),
		executables.WithServer(apiServerUrl),
		executables.WithToken(jwt),
		executables.WithSkipTLSVerify(),
		executables.WithAllNamespaces(),
	)
	if err != nil {
		e.T.Errorf("Error getting pods: %v", err)
	}

	e.T.Log("Getting deployments with OIDC token")
	_, err = e.KubectlClient.GetDeployments(
		ctx,
		executables.WithKubeconfig(filepath.Join(e.ClusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", e.ClusterName))),
		executables.WithServer(apiServerUrl),
		executables.WithToken(jwt),
		executables.WithSkipTLSVerify(),
		executables.WithAllNamespaces(),
	)
	if err != nil {
		e.T.Errorf("Error getting deployments: %v", err)
	}
}

// WithOIDCEnvVarCheck returns a ClusterE2ETestOpt that checks for the required env vars.
func WithOIDCEnvVarCheck() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, oidcRequiredEnvVars)
	}
}
